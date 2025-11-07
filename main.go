package main

import (
	"github.com/pocketbase/pocketbase"
	_ "github.com/shynome/auto-tls/db/migrations"
	"github.com/shynome/err0/try"
)

var Version = "dev"

func main() {
	app := pocketbase.New()
	app.RootCmd.Version = Version
	app.OnServe().BindFunc(bindTLS)
	app.OnServe().BindFunc(bindDeploy)
	try.To(app.Start())
}
