package main

import (
	"context"
	"log"

	"github.com/GorgoFramework/gorgo/pkg/gorgo"
	"github.com/GorgoFramework/gorgo/plugins/sql"
)

func main() {
	sqlPlugin := sql.NewSqlPlugin()
	app := gorgo.New().AddPlugin(sqlPlugin)

	app.Get("/user", func(ctx *gorgo.Context) error {
		db, _ := ctx.GetService("sql")

		pool := db.(*sql.SqlPlugin).GetPool()

		stmt := "SELECT * FROM users WHERE id = $1"

		var username string
		err := pool.QueryRow(context.Background(), stmt, 1).Scan(&username)
		if err != nil {
			return ctx.JSON(gorgo.Map{"error": err.Error()})
		}

		return ctx.JSON(gorgo.Map{"username": username})
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
