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

		products := []string{
			db.ProductCDN,
			db.ProductDCDN,
		}
		targets := []string{
			db.TargetAliyun,
		}
		deploys := core.NewBaseCollection(db.TableDeploys, ID(db.TableDeploys))
		deploys.Fields.Add(
			&core.SelectField{
				Name: "target", Id: ID("target"), System: true,
				Required: true, Presentable: true,
				Values: targets, MaxSelect: 1,
			},
			&core.RelationField{
				Name: "domain", Id: ID("domain"), System: true,
				Required: true, Presentable: true,
				CollectionId: domains.Id, MaxSelect: 1,
			},
			&core.SelectField{
				Name: "products", Id: ID("products"), System: true,
				Required: false,
				Values:   products, MaxSelect: len(products),
			},
			&core.JSONField{
				Name: "value", Id: ID("value"), System: true,
				Required: false,
			},
		)
		addUpdatedFields(deploys)
		try.To(app.Save(deploys))

		tasks := core.NewBaseCollection(db.TableTasks, ID(db.TableTasks))
		tasks.Fields.Add(
			&core.RelationField{
				Name: "deploy", Id: ID("deploy"), System: true,
				Required: true, Presentable: true,
				CollectionId: deploys.Id, CascadeDelete: true, MaxSelect: 1,
			},
			&core.TextField{
				Name: "product", Id: ID("product"), System: true,
				Required: false, Presentable: true,
			},
			&core.BoolField{
				Name: "deployed", Id: ID("deployed"), System: true,
				Required: false,
			},
			&core.JSONField{
				Name: "payload", Id: ID("payload"), System: true,
				Required: false,
			},
			&core.JSONField{
				Name: "result", Id: ID("result"), System: true,
				Required: false,
			},
		)
		addUpdatedFields(tasks)
		try.To(app.Save(tasks))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("no init deploy rollback")
	})
}
