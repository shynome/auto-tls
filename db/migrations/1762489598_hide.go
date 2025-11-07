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

		{
			acmes := try.To1(app.FindCollectionByNameOrId(db.TableACMEs))
			EAB := acmes.Fields.GetByName("EAB").(*core.JSONField)
			EAB.Hidden = true
			try.To(app.Save(acmes))
		}

		{
			deploys := try.To1(app.FindCollectionByNameOrId(db.TableDeploys))
			value := deploys.Fields.GetByName("value").(*core.JSONField)
			value.Hidden = true
			try.To(app.Save(deploys))
		}

		{
			dnsp := try.To1(app.FindCollectionByNameOrId(db.TableDNSP))
			value := dnsp.Fields.GetByName("value").(*core.JSONField)
			value.Hidden = true
			try.To(app.Save(dnsp))
		}

		{
			domains := try.To1(app.FindCollectionByNameOrId(db.TableDomains))
			token := domains.Fields.GetByName("token").(*core.TextField)
			token.Hidden = true
			try.To(app.Save(domains))
		}

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("no init rollback")
	})
}
