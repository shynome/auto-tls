package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/libdns/cloudflare"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/shynome/auto-tls/db"
	"github.com/shynome/auto-tls/tools"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	"golang.org/x/sync/errgroup"
)

const neverRenewTime = time.Hour * 24 * 365 * 100 // 100 year

func bindTLS(se *core.ServeEvent) error {
	app := se.App

	storage := &certmagic.FileStorage{Path: filepath.Join(app.DataDir(), "certmagic")}
	cfg := certmagic.Config{Storage: storage}
	cache := certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(cert certmagic.Certificate) (*certmagic.Config, error) {
			return nil, nil // 不需要 cache 进行 renew, 每天的定时任务会进行 renew
		},
		RenewCheckInterval: neverRenewTime,
	})
	// 每天执行一次续期任务
	app.Cron().MustAdd("certmagic", "0 0 * * *", func() {
		ManageAsync(app, cache, cfg)
	})

	se.Router.GET("/api/cert/{id}", func(e *core.RequestEvent) (err error) {
		defer err0.Then(&err, nil, nil)
		r := e.Request
		id := r.PathValue("id")
		record := try.To1(e.App.FindRecordById(db.TableDomains, id))
		token := record.GetString("token")
		if token == "" {
			return apis.NewUnauthorizedError("此记录未开启访问权限", nil)
		}
		rtoken := r.FormValue("token")
		if token != rtoken {
			return apis.NewUnauthorizedError("认证失败", nil)
		}
		expired := record.GetDateTime("expired").Time()
		if expired.IsZero() {
			return apis.NewNotFoundError("证书尚未准备好", nil)
		}
		if time.Now().After(expired) {
			return apis.NewNotFoundError("证书已过期", nil)
		}

		pem, fn := getPEMTry(e.App, record)

		disposition := mime.FormatMediaType("inline", map[string]string{"filename": fn})
		h := e.Response.Header()
		h.Set("Content-Disposition", disposition)

		return e.Blob(http.StatusOK, "text/plain", pem)
	})

	return se.Next()
}

func getPEMTry(app core.App, record *core.Record) (pem []byte, fn string) {
	domain := record.GetString("domain")
	f := filepath.Join(record.BaseFilesPath(), record.GetString("pem"))
	fs := try.To1(app.NewFilesystem())
	defer fs.Close()

	attrs := try.To1(fs.Attributes(f))
	fn = attrs.Metadata["original-filename"]
	if fn == "" {
		fn = domain + ".pem"
	}

	r := try.To1(fs.GetReader(f))
	defer r.Close()
	pem = try.To1(io.ReadAll(r))
	return pem, fn
}

func parseUint16Str[T ~uint16](s string) ([]T, error) {
	ss := strings.Split(s, ",")
	uu := make([]T, len(ss))
	for i, s := range ss {
		u, err := strconv.ParseUint(s, 16, 16)
		if err != nil {
			return nil, err
		}
		uu[i] = T(u)
	}
	return uu, nil
}

func ManageAsync(app core.App, cache *certmagic.Cache, cfg certmagic.Config) (err error) {
	logger := app.Logger()
	defer err0.Then(&err, nil, func() {
		logger.Error("issue certs failed", "error", err)
	})

	q := dbx.Not(dbx.HashExp{"dns_provider": ""})
	domains := try.To1(app.FindAllRecords(db.TableDomains, q))

	var eg errgroup.Group
	for _, domain := range domains {
		eg.Go(func() (err error) {
			logger := logger.With("domain", domain)
			defer err0.Then(&err, nil, func() {
				logger.Error("issue cert failed", "error", err)
			})

			domain := MagicDomain(domain)
			magic := try.To1(domain.Magic(app, cache, cfg))
			d := domain.GetString("domain")
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
			defer cancel()
			err = magic.ManageSync(ctx, []string{d})
			try.To(err)

			ctx2 := context.Background()
			c := try.To1(magic.CacheManagedCertificate(ctx2, d))
			if c.Leaf != nil {
				expired := try.To1(types.ParseDateTime(c.Leaf.NotAfter))
				expiredR := domain.GetDateTime("expired")
				if expired.Equal(expiredR) {
					return // 相同的证书, 不用更新
				}
				pem := try.To1(tools.ToPEM(c.Certificate))
				fn := fmt.Sprintf("%s-%s.pem", expired.Time().Format("2006-01-02"), d)
				f := try.To1(filesystem.NewFileFromBytes(pem, fn))
				err := app.RunInTransaction(func(tx core.App) error {
					domain, err := tx.FindRecordById(db.TableDomains, domain.Id)
					if err != nil {
						return err
					}
					domain.Set("expired", expired)
					domain.Set("pem", f)
					err = tx.Save(domain)
					return err
				})
				try.To(err)
			}

			return
		})
	}
	try.To(eg.Wait())

	return nil
}

type Domain struct {
	core.BaseRecordProxy
}

func MagicDomain(r *core.Record) *Domain {
	d := &Domain{}
	d.SetProxyRecord(r)
	return d
}

func (domain *Domain) Magic(app core.App, cache *certmagic.Cache, cfg certmagic.Config) (_ *certmagic.Config, err error) {
	defer err0.Then(&err, nil, nil)

	acme := try.To1(app.FindRecordById(db.TableACMEs, domain.GetString("acme")))
	email := acme.GetString("email")
	CA := db.MagicCA(acme.GetString("CA"))

	dnsp := try.To1(app.FindRecordById(db.TableDNSP, domain.GetString("dns_provider")))
	p := dnsp.GetString("provider")
	var provider certmagic.DNSProvider
	switch p {
	case db.DNSPCloudflare:
		v := dnsp.GetString("value")
		var p cloudflare.Provider
		json.Unmarshal([]byte(v), &p)
		provider = &p
	default:
		return nil, fmt.Errorf("unsupported dns provider: %s", p)
	}

	magic := certmagic.New(cache, cfg)
	issuer := certmagic.NewACMEIssuer(magic, certmagic.ACMEIssuer{
		CA:     CA,
		Email:  email,
		Agreed: true,
		DNS01Solver: &certmagic.DNS01Solver{
			DNSManager: certmagic.DNSManager{DNSProvider: provider},
		},
	})
	magic.Issuers = []certmagic.Issuer{issuer}

	return magic, nil
}
