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

		domains := try.To1(app.FindCollectionByNameOrId(db.TableDomains))
		domains.Fields.AddAt(getFieldNext(domains, "domain"),
			&core.DateField{
				Id: ID("expired"), Name: "expired", System: true,
			},
		)
		try.To(app.Save(domains))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("no init rollback")
	})
}
