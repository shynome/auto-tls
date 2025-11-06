package aliyun

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	"github.com/shynome/auto-tls/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

func Deploy(product string, payload []byte) ([]byte, error) {
	switch product {
	case db.ProductCDN, db.ProductDCDN:
		return deployCDN(product, payload)
	}
	return nil, fmt.Errorf("")
}

func deployCDN(product string, payload []byte) (_ []byte, err error) {
	return
}

type Aliyun struct {
	core.BaseRecordProxy
}

func (d *Aliyun) Deploy(app core.App, cert tls.Certificate) (err error) {
	defer err0.Then(&err, nil, nil)

	domainR := try.To1(app.FindRecordById(db.TableDomains, d.GetString("domain")))
	suffix := domainR.GetString("domain")
	suffix = strings.Replace(suffix, "*.", "", 1)

	products := d.GetStringSlice("products")
	for _, p := range products {
		switch p {
		case db.ProductCDN:
			// create cdn deploy task
		case db.ProductDCDN:
			// create dcdn deploy task
		}
	}

	return nil
}
