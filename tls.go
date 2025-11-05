package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/caddyserver/certmagic"
	"github.com/libdns/cloudflare"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/shynome/auto-tls/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	"golang.org/x/sync/errgroup"
)

func bindTLS(se *core.ServeEvent) error {
	app := se.App

	g := certmagic.NewDefault()
	cache := certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(certmagic.Certificate) (*certmagic.Config, error) {
			return g, nil
		},
	})
	storage := &certmagic.FileStorage{Path: filepath.Join(app.DataDir(), "certmagic")}
	magicGen := GenMagic(app, cache, storage)
	// 每天执行一次续期任务
	app.Cron().MustAdd("certmagic", "0 0 * * *", func() {
		ManageAsync(app, magicGen)
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
		domain := record.GetString("domain")
		magic := magicGen("", nil)
		ctx := r.Context()

		cc := try.To1(magic.CacheManagedCertificate(ctx, domain)) // 证书不存在时抛出的文件不存在错误会直接被转换成404错误, 不用另外处理了
		var cert tls.Certificate = cc.Certificate

		body := &bytes.Buffer{}
		for _, der := range cert.Certificate {
			block := &pem.Block{Type: "CERTIFICATE", Bytes: der}
			try.To(pem.Encode(body, block))
		}
		switch key := cert.PrivateKey.(type) {
		case *rsa.PrivateKey:
			der := x509.MarshalPKCS1PrivateKey(key)
			block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}
			try.To(pem.Encode(body, block))
		case *ecdsa.PrivateKey:
			der := try.To1(x509.MarshalECPrivateKey(key))
			block := &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}
			try.To(pem.Encode(body, block))
		case ed25519.PrivateKey:
			der := try.To1(x509.MarshalPKCS8PrivateKey(key))
			block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
			try.To(pem.Encode(body, block))
		default:
			return apis.NewInternalServerError("尚未支持的私钥类型", nil)
		}

		disposition := mime.FormatMediaType("inline", map[string]string{"filename": domain + ".pem"})
		h := e.Response.Header()
		h.Set("Content-Disposition", disposition)
		return e.Blob(http.StatusOK, "text/plain", body.Bytes())
	})

	return se.Next()
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

func ManageAsync(app core.App, genMagic MagicGen) (err error) {
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

			acme := try.To1(app.FindRecordById(db.TableACMEs, domain.GetString("acme")))
			email := acme.GetString("email")

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
				return fmt.Errorf("unsupported dns provider: %s", p)
			}

			magic := genMagic(email, provider)
			d := domain.GetString("domain")
			ctx := context.Background()
			err = magic.ManageSync(ctx, []string{d})
			return err
		})
	}
	try.To(eg.Wait())

	return nil
}

type MagicGen func(email string, provider certmagic.DNSProvider) *certmagic.Config

func GenMagic(app core.App, cache *certmagic.Cache, storage *certmagic.FileStorage) MagicGen {
	cfg := certmagic.Config{Storage: storage}
	return func(email string, provider certmagic.DNSProvider) *certmagic.Config {
		magic := certmagic.New(cache, cfg)
		issuer := certmagic.NewACMEIssuer(magic, certmagic.ACMEIssuer{
			Email:  email,
			Agreed: true,
			DNS01Solver: &certmagic.DNS01Solver{
				DNSManager: certmagic.DNSManager{DNSProvider: provider},
			},
		})
		if app.IsDev() {
			issuer.CA = certmagic.LetsEncryptStagingCA
		}
		magic.Issuers = []certmagic.Issuer{issuer}
		return magic
	}
}
