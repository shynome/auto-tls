package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/migrations"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

func init() {
	migrations.Register(func(app core.App) (err error) {
		defer err0.Then(&err, nil, nil)

		superusers := try.To1(app.FindCollectionByNameOrId(core.CollectionNameSuperusers))
		superusers.Fields.AddAt(getFieldIndex(superusers, "created"),
			&core.BoolField{
				Name: "notify_muted", Id: ID("notify_muted"), System: true,
				Required: false,
			},
		)
		try.To(app.Save(superusers))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("no init rollback")
	})
}
