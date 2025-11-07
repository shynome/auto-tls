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
		email := acmes.Fields.GetByName("email").(*core.EmailField)
		email.Required = true // email 必填
		CAs := []string{
			db.CALetsEncryptProduction,
			db.CAZeroSSLProduction,
			db.CAGoogleTrustProduction,
			// Staging 测试用的证书放后面
			db.CALetsEncryptStaging,
			db.CAGoogleTrustStaging,
		}
		acmes.Fields.AddAt(0,
			&core.SelectField{
				Name: "CA", Id: ID("CA"), System: true,
				Required: true, Presentable: true,
				Values: CAs, MaxSelect: 1,
			},
		)
		try.To(app.Save(acmes))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("no init rollback")
	})
}
