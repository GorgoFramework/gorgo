package main

import (
	"fmt"
	"log"

	"github.com/GorgoFramework/gorgo/pkg/gorgo"
)

func main() {
	app := gorgo.New()

	// Simple parameter example
	app.Get("/hello/:name", func(ctx *gorgo.Context) error {
		name := ctx.Param("name")
		return ctx.String(fmt.Sprintf("Hello, %s!", name))
	})

	// Multiple parameters example
	app.Get("/users/:userId/posts/:postId", func(ctx *gorgo.Context) error {
		userId := ctx.Param("userId")
		postId := ctx.Param("postId")

		return ctx.JSON(gorgo.Map{
			"message": "User post retrieved",
			"userId":  userId,
			"postId":  postId,
		})
	})

	// Mixed static and dynamic segments
	app.Get("/api/v1/users/:id/profile", func(ctx *gorgo.Context) error {
		id := ctx.Param("id")

		return ctx.JSON(gorgo.Map{
			"message": "User profile",
			"userId":  id,
			"version": "v1",
		})
	})

	// Nested parameters
	app.Get("/categories/:category/products/:productId/reviews/:reviewId", func(ctx *gorgo.Context) error {
		category := ctx.Param("category")
		productId := ctx.Param("productId")
		reviewId := ctx.Param("reviewId")

		return ctx.JSON(gorgo.Map{
			"category":  category,
			"productId": productId,
			"reviewId":  reviewId,
		})
	})

	// Root route without parameters
	app.Get("/", func(ctx *gorgo.Context) error {
		return ctx.JSON(gorgo.Map{
			"message": "Welcome to Gorgo Framework Parameter Example!",
			"routes": []string{
				"GET /hello/:name",
				"GET /users/:userId/posts/:postId",
				"GET /api/v1/users/:id/profile",
				"GET /categories/:category/products/:productId/reviews/:reviewId",
			},
		})
	})

	log.Println("Starting server with parameter parsing examples...")
	log.Println("Try these URLs:")
	log.Println("  http://localhost:8080/")
	log.Println("  http://localhost:8080/hello/John")
	log.Println("  http://localhost:8080/users/123/posts/456")
	log.Println("  http://localhost:8080/api/v1/users/789/profile")
	log.Println("  http://localhost:8080/categories/electronics/products/laptop123/reviews/review456")

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
