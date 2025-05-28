package main

import (
	"log"

	"github.com/GorgoFramework/gorgo/pkg/gorgo"
)

func main() {
	app := gorgo.New()

	app.Get("/", func(ctx *gorgo.Context) error {
		ctx.JSON(gorgo.Map{"message": "Hello, World!", "Demo": 1})
		return nil
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
