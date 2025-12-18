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
		caList := acmes.Fields.GetByName("CA").(*core.SelectField)
		vv := []string{}
		for _, v := range caList.Values {
			vv = append(vv, v)
			if v == db.CAZeroSSLProduction {
				vv = append(vv, db.CATrustAsia) // 放在 ZeroSSL 后面
			}
		}
		caList.Values = vv
		try.To(app.Save(acmes))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("no init rollback")
	})
}
