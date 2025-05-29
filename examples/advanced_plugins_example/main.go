package main

import (
	"log"

	"github.com/GorgoFramework/gorgo/pkg/gorgo"
	"github.com/GorgoFramework/gorgo/plugins/monitoring"
	"github.com/GorgoFramework/gorgo/plugins/sql"
)

func main() {
	// Create application
	app := gorgo.New()

	// Add plugins
	monitoringPlugin := monitoring.NewMonitoringPlugin()

	app.AddPlugin(monitoringPlugin)

	// Enable CORS
	app.EnableCORS()

	// Enable rate limiting
	app.EnableRateLimit(gorgo.RateLimitOptions{
		RequestsPerMinute: 100,
		BurstSize:         10,
	})

	// Create API route group
	api := app.Group("/api/v1")

	// Basic routes
	app.Get("/", func(ctx *gorgo.Context) error {
		return ctx.JSON(gorgo.Map{
			"message": "Welcome to Gorgo Framework with Enhanced Plugins!",
			"version": "0.0.4",
			"plugins": []string{"sql", "monitoring"},
		})
	})

	// Endpoint for metrics
	app.Get("/metrics", func(ctx *gorgo.Context) error {
		// Simple metrics implementation without direct plugin access
		return ctx.JSON(gorgo.Map{
			"message": "Metrics endpoint",
			"note":    "Monitoring plugin provides automatic metrics collection",
		})
	})

	// API routes with middleware
	api.Get("/users", getUsersHandler)
	api.Post("/users", createUserHandler,
		// Transaction middleware (if SQL plugin is available)
		func(next gorgo.HandlerFunc) gorgo.HandlerFunc {
			return func(ctx *gorgo.Context) error {
				if _, exists := ctx.GetService("sql"); exists {
					if sqlPlugin, ok := ctx.GetPlugin("sql"); ok {
						if plugin, ok := sqlPlugin.(*sql.SqlPlugin); ok {
							return plugin.TransactionMiddleware()(next)(ctx)
						}
					}
				}
				return next(ctx)
			}
		},
	)

	// Route for event demonstration
	app.Get("/trigger-event", func(ctx *gorgo.Context) error {
		eventBus := app.GetEventBus()

		// Publish custom event
		err := eventBus.Publish(ctx.FastHTTP(), "custom.event", gorgo.Map{
			"user_id": ctx.Query("user_id"),
			"action":  "test_event",
		})

		if err != nil {
			return ctx.Status(gorgo.InternalServerErrorStatus).JSON(gorgo.Map{"error": err.Error()})
		}

		return ctx.JSON(gorgo.Map{"message": "Event triggered successfully"})
	})

	// Route for plugin hot reload
	app.Post("/admin/reload-plugin/:name", func(ctx *gorgo.Context) error {
		pluginName := ctx.Param("name")

		// New configuration (in real app, get from request body)
		newConfig := map[string]interface{}{
			"enabled":         true,
			"report_interval": 30,
		}

		err := app.HotReloadPlugin(pluginName, newConfig)
		if err != nil {
			return ctx.Status(gorgo.BadRequestStatus).JSON(gorgo.Map{"error": err.Error()})
		}

		return ctx.JSON(gorgo.Map{"message": "Plugin reloaded successfully"})
	})

	// Route for plugin information
	app.Get("/admin/plugins", func(ctx *gorgo.Context) error {
		plugins := []gorgo.Map{}

		if sqlPlugin, exists := app.GetPlugin("sql"); exists {
			metadata := sqlPlugin.GetMetadata()
			plugins = append(plugins, gorgo.Map{
				"name":        metadata.Name,
				"version":     metadata.Version,
				"description": metadata.Description,
				"state":       sqlPlugin.GetState(),
				"priority":    metadata.Priority,
				"tags":        metadata.Tags,
			})
		}

		if monPlugin, exists := app.GetPlugin("monitoring"); exists {
			metadata := monPlugin.GetMetadata()
			plugins = append(plugins, gorgo.Map{
				"name":        metadata.Name,
				"version":     metadata.Version,
				"description": metadata.Description,
				"state":       monPlugin.GetState(),
				"priority":    metadata.Priority,
				"tags":        metadata.Tags,
			})
		}

		return ctx.JSON(gorgo.Map{"plugins": plugins})
	})

	// Start server
	log.Fatal(app.Run())
}

func getUsersHandler(ctx *gorgo.Context) error {
	// Demonstrate context data usage
	userAgent := ctx.UserAgent()
	clientIP := ctx.ClientIP()

	users := []gorgo.Map{
		{"id": 1, "name": "John Doe", "email": "john@example.com"},
		{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
	}

	return ctx.JSON(gorgo.Map{
		"users":      users,
		"user_agent": userAgent,
		"client_ip":  clientIP,
		"total":      len(users),
	})
}

func createUserHandler(ctx *gorgo.Context) error {
	// Demonstrate JSON handling
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := ctx.BindJSON(&user); err != nil {
		return ctx.Status(gorgo.BadRequestStatus).JSON(gorgo.Map{"error": "Invalid JSON"})
	}

	// Simulate user creation
	newUser := gorgo.Map{
		"id":    3,
		"name":  user.Name,
		"email": user.Email,
	}

	return ctx.Status(gorgo.CreatedStatus).JSON(gorgo.Map{
		"message": "User created successfully",
		"user":    newUser,
	})
}
