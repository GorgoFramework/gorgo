package main

import (
	"github.com/GorgoFramework/gorgo/pkg/gorgo"
)

func main() {
	app := gorgo.New()

	app.Get("/", func(ctx *gorgo.Context) error {
		ctx.JSON(gorgo.Map{"message": "Hello, World!", "Demo": 1})
		return nil
	})

	app.Run()
}
