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

		acmes := try.To1(app.FindCollectionByNameOrId(db.TableACMEs))
		acmes.Fields.AddAt(getFieldNext(acmes, "agreed"),
			&core.JSONField{
				Name: "EAB", Id: ID("EAB"), System: true,
				Required: false,
			},
		)
		try.To(app.Save(acmes))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("no init rollback")
	})
}
