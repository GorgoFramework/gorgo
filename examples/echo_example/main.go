package main

import (
	"fmt"
	"log"

	"github.com/GorgoFramework/gorgo/pkg/gorgo"
)

func main() {
	app := gorgo.New()

	app.Get("/:username", func(ctx *gorgo.Context) error {
		return ctx.String(fmt.Sprintf("Hello, %s!", ctx.Param("username")))
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

}
