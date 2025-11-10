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

		dnsp := try.To1(app.FindCollectionByNameOrId(db.TableDNSP))
		pp := dnsp.Fields.GetByName("provider").(*core.SelectField)
		pp.Values = append(pp.Values, db.DNSPAlidns)
		try.To(app.Save(dnsp))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("no init rollback")
	})
}
