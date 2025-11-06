package main

import (
	"github.com/pocketbase/pocketbase"
	_ "github.com/shynome/auto-tls/db/migrations"
	"github.com/shynome/err0/try"
)

func main() {
	app := pocketbase.New()
	app.OnServe().BindFunc(bindTLS)
	app.OnServe().BindFunc(bindDeploy)
	try.To(app.Start())
}
