package main

import (
	"net/netip"
	"strings"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/libdns/libdns"
	"github.com/miekg/dns"
	"github.com/pocketbase/pocketbase/core"
	"github.com/shynome/auto-tls/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	"go.uber.org/zap"
)

func bindDDNS(se *core.ServeEvent) error {
	se.App.OnRecordViewRequest(db.TableDDNS).BindFunc(func(e *core.RecordRequestEvent) (err error) {
		defer err0.Then(&err, nil, nil)

		ip4 := e.RealIP()
		ip, err := netip.ParseAddr(ip4)
		wantUpdate := e.Request.URL.Query().Has("token")
		switch {
		case err != nil, // 地址解析出错跳过
			!ip.Is4(),   // 只支持更新ip4现在
			!wantUpdate: // 不是更新请求不进行更新
			return e.Next()
		}

		e.Record.Set("ip4", ip4)
		e.Record.Set("apply", true)
		try.To(e.App.SaveWithContext(e.Request.Context(), e.Record))
		return e.Next()
	})
	se.App.OnRecordUpdate(db.TableDDNS).BindFunc(func(e *core.RecordEvent) (err error) {
		app := e.App
		ctx := e.Context
		defer err0.Then(&err, nil, nil)

		if !e.Record.GetBool("apply") {
			return e.Next()
		}
		e.Record.Set("apply", false)

		old := try.To1(app.FindRecordById(db.TableDDNS, e.Record.Id))
		ip4old := old.GetString("ip4")
		ip4 := e.Record.GetString("ip4")

		if ip4old == ip4 {
			// 地址未变动不更新
			return e.Next()
		}

		domain := e.Record.GetString("domain")
		fqdn := dns.Fqdn(domain)
		zone := try.To1(certmagic.FindZoneByFQDN(ctx, zap.L(), fqdn, certmagic.RecursiveNameservers(nil)))
		dn, _ := strings.CutSuffix(fqdn, "."+zone)
		provider := getDNSProviderTry(app, e.Record.GetString("dns_provider"))

		if ip4old != "" {
			ip := try.To1(netip.ParseAddr(ip4old))
			_, err := provider.DeleteRecords(ctx, zone, []libdns.Record{
				libdns.Address{Name: dn, IP: ip, TTL: time.Minute},
			})
			if err != nil {
				app.Logger().Warn("移除旧记录出错", "ip4", ip4old, "error", err)
			}
		}
		ip := try.To1(netip.ParseAddr(ip4))
		try.To1(provider.AppendRecords(ctx, zone, []libdns.Record{
			libdns.Address{Name: dn, IP: ip, TTL: time.Minute},
		}))

		return e.Next()
	})
	return se.Next()
}
