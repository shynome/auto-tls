package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/shynome/auto-tls/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

func init() {
	migrations.Register(func(app core.App) (err error) {
		defer err0.Then(&err, nil, nil)

		dnsp := try.To1(app.FindCollectionByNameOrId(db.TableDNSP))

		ddns := core.NewBaseCollection(db.TableDDNS, ID(db.TableDDNS))
		ddns.Fields.Add(
			&core.TextField{
				Name: "domain", Id: ID("domain"), System: true,
				Required: true, Presentable: true,
			},
			&core.RelationField{
				Name: "dns_provider", Id: ID("dns_provider"), System: true,
				Required:     true,
				CollectionId: dnsp.Id, MaxSelect: 1,
			},
			&core.TextField{
				Name: "token", Id: ID("token"), System: true,
				Required: true, Hidden: true,
			},
			&core.TextField{
				Name: "ip4", Id: ID("ip4"), System: true,
				Required: false,
			},
			&core.BoolField{
				Name: "apply", Id: ID("apply"), System: true,
				Required: false,
			},
		)
		addUpdatedFields(ddns)
		ddns.ViewRule = types.Pointer("@request.query.token = token")
		try.To(app.Save(ddns))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("no init rollback")
	})
}
