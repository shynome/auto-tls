package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/migrations"
	"github.com/shynome/auto-tls/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

func init() {
	migrations.Register(func(app core.App) (err error) {
		defer err0.Then(&err, nil, nil)

		acmes := core.NewBaseCollection(db.TableACMEs, ID(db.TableACMEs))
		acmes.Fields.Add(
			&core.EmailField{
				Name: "email", Id: ID("email"), System: true,
				Presentable: true,
			},
			&core.BoolField{
				Name: "agreed", Id: ID("agreed"), System: true,
				Required: true, // 必须要同意
			},
		)
		addUpdatedFields(acmes)
		try.To(app.Save(acmes))

		dnsp := core.NewBaseCollection(db.TableDNSP, ID(db.TableDNSP))
		dnsp.Fields.Add(
			&core.TextField{
				Name: "name", Id: ID("name"), System: true,
				Presentable: true,
			},
			&core.SelectField{
				Name: "provider", Id: ID("provider"), System: true,
				Required: true, Presentable: true,
				Values:    []string{db.DNSPCloudflare}, // 目前只支持 Cloudflare
				MaxSelect: 1,
			},
			&core.JSONField{
				Name: "value", Id: ID("value"), System: true,
				Required: true,
			},
		)
		addUpdatedFields(dnsp)
		try.To(app.Save(dnsp))

		domains := core.NewBaseCollection(db.TableDomains, ID(db.TableDomains))
		domains.Fields.Add(
			&core.TextField{
				Name: "domain", Id: ID("domain"), System: true,
				Required: true, Presentable: true,
			},
			&core.RelationField{
				Name: "acme", Id: ID("acme"), System: true,
				Required:     true,
				CollectionId: acmes.Id, MaxSelect: 1,
			},
			&core.RelationField{
				Name: "dns_provider", Id: ID("dns_provider"), System: true,
				Required:     false, // 没有则不申请证书
				CollectionId: dnsp.Id, MaxSelect: 1,
			},
			&core.TextField{
				Name: "token", Id: ID("token"), System: true,
				Required: false, // 未设置token时则不允许下载此证书
			},
		)
		addUpdatedFields(domains)
		try.To(app.Save(domains))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("no init tls rollback")
	})
}
