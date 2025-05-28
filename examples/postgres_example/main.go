package main

import (
	"context"
	"log"

	"github.com/GorgoFramework/gorgo/pkg/gorgo"
	"github.com/GorgoFramework/gorgo/plugins/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	sqlPlugin := sql.NewSqlPlugin()
	app := gorgo.New().AddPlugin(sqlPlugin)

	// Get user by ID
	app.Get("/user/:id", func(ctx *gorgo.Context) error {
		db, exists := ctx.GetService("sql")
		if !exists {
			ctx.SetStatusCode(500)
			return ctx.JSON(gorgo.Map{"error": "Database service not available"})
		}

		pool := db.(*pgxpool.Pool)
		userID := ctx.Param("id")

		var username, email string
		query := "SELECT username, email FROM users WHERE id = $1"
		err := pool.QueryRow(context.Background(), query, userID).Scan(&username, &email)

		if err != nil {
			if err == pgx.ErrNoRows {
				ctx.SetStatusCode(404)
				return ctx.JSON(gorgo.Map{"error": "User not found"})
			}

			ctx.SetStatusCode(500)
			return ctx.JSON(gorgo.Map{"error": "Database error"})
		}

		return ctx.JSON(gorgo.Map{
			"id":       userID,
			"username": username,
			"email":    email,
		})
	})

	// Get all users
	app.Get("/users", func(ctx *gorgo.Context) error {
		db, exists := ctx.GetService("sql")
		if !exists {
			ctx.SetStatusCode(500)
			return ctx.JSON(gorgo.Map{"error": "Database service not available"})
		}

		pool := db.(*pgxpool.Pool)

		rows, err := pool.Query(context.Background(), "SELECT id, username, email FROM users")
		if err != nil {
			ctx.SetStatusCode(500)
			return ctx.JSON(gorgo.Map{"error": "Database error"})
		}
		defer rows.Close()

		var users []map[string]interface{}

		for rows.Next() {
			var id int
			var username, email string

			if err := rows.Scan(&id, &username, &email); err != nil {
				ctx.SetStatusCode(500)
				return ctx.JSON(gorgo.Map{"error": "Database error"})
			}

			users = append(users, map[string]interface{}{
				"id":       id,
				"username": username,
				"email":    email,
			})
		}

		return ctx.JSON(gorgo.Map{"users": users})
	})

	// Root endpoint with documentation
	app.Get("/", func(ctx *gorgo.Context) error {
		return ctx.JSON(gorgo.Map{
			"message": "PostgreSQL Example API",
			"endpoints": []string{
				"GET / - This documentation",
				"GET /users - Get all users",
				"GET /user/:id - Get user by ID",
			},
			"note": "Make sure to create a 'users' table with id, username, and email columns",
		})
	})

	log.Println("Starting PostgreSQL example server...")
	log.Println("Available endpoints:")
	log.Println("  GET / - Documentation")
	log.Println("  GET /users - Get all users")
	log.Println("  GET /user/:id - Get user by ID")

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
