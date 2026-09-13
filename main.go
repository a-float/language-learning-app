package main

import (
	"log"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		se.Router.GET("/static/{path...}", apis.Static(os.DirFS("./pb_public"), false))
		se.Router.GET("/favicon.ico", func(e *core.RequestEvent) error {
			http.ServeFile(e.Response, e.Request, "./pb_public/favicon.ico")
			return nil
		})

		se.Router.GET("/hello/{name}", func(e *core.RequestEvent) error {
			name := e.Request.PathValue("name")
			return e.String(http.StatusOK, "Hello "+name)
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
