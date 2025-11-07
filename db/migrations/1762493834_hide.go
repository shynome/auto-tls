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

		s := app.Settings()
		s.Meta.HideControls = true
		try.To(app.Save(s))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("no init rollback")
	})
}
