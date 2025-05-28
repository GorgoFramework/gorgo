package main

import (
	"fmt"
	"log"

	"github.com/GorgoFramework/gorgo/pkg/gorgo"
)

func main() {
	app := gorgo.New()

	// Example demonstrating all parameter methods
	app.Get("/demo/:name/:age", func(ctx *gorgo.Context) error {
		// Basic parameter access
		name := ctx.Param("name")
		age := ctx.Param("age")

		// Parameter with default value
		role := ctx.ParamDefault("role", "user")

		// Check if parameter exists
		hasName := ctx.HasParam("name")
		hasRole := ctx.HasParam("role")

		// Get all parameters
		allParams := ctx.Params()

		return ctx.JSON(gorgo.Map{
			"name":      name,
			"age":       age,
			"role":      role,
			"hasName":   hasName,
			"hasRole":   hasRole,
			"allParams": allParams,
		})
	})

	// Optional parameter example
	app.Get("/users/:id", func(ctx *gorgo.Context) error {
		id := ctx.Param("id")
		format := ctx.ParamDefault("format", "json")

		response := gorgo.Map{
			"userId": id,
			"format": format,
		}

		if format == "xml" {
			return ctx.String(fmt.Sprintf("<user><id>%s</id></user>", id))
		}

		return ctx.JSON(response)
	})

	// Complex nested route
	app.Get("/api/:version/users/:userId/posts/:postId/comments/:commentId", func(ctx *gorgo.Context) error {
		// Get all parameters
		params := ctx.Params()

		// Validate required parameters
		requiredParams := []string{"version", "userId", "postId", "commentId"}
		missing := []string{}

		for _, param := range requiredParams {
			if !ctx.HasParam(param) {
				missing = append(missing, param)
			}
		}

		if len(missing) > 0 {
			return ctx.JSON(gorgo.Map{
				"error":         "Missing required parameters",
				"missingParams": missing,
			})
		}

		return ctx.JSON(gorgo.Map{
			"message":    "Comment retrieved successfully",
			"parameters": params,
			"endpoint": fmt.Sprintf("API %s - User %s - Post %s - Comment %s",
				ctx.Param("version"),
				ctx.Param("userId"),
				ctx.Param("postId"),
				ctx.Param("commentId")),
		})
	})

	// Root route with documentation
	app.Get("/", func(ctx *gorgo.Context) error {
		return ctx.JSON(gorgo.Map{
			"message": "Advanced Parameter Parsing Example",
			"features": []string{
				"ctx.Param(key) - Get parameter value",
				"ctx.ParamDefault(key, default) - Get parameter with default",
				"ctx.HasParam(key) - Check if parameter exists",
				"ctx.Params() - Get all parameters",
			},
			"examples": []string{
				"GET /demo/John/25 - Basic parameter demo",
				"GET /users/123 - Optional parameter demo",
				"GET /api/v1/users/123/posts/456/comments/789 - Complex nested route",
			},
		})
	})

	log.Println("Starting advanced parameter parsing example...")
	log.Println("Available endpoints:")
	log.Println("  GET / - Documentation")
	log.Println("  GET /demo/:name/:age - Parameter methods demo")
	log.Println("  GET /users/:id - Optional parameters")
	log.Println("  GET /api/:version/users/:userId/posts/:postId/comments/:commentId - Complex route")

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
