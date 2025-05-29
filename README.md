# Gorgo Framework

A modern and fast web framework for Go, built on top of FastHTTP.

## Features

- 🚀 **High Performance** - built on FastHTTP
- 🛣️ **Flexible Routing** - URL parameter support
- 🔌 **Enhanced Plugin System** - powerful and extensible architecture
- 🏗️ **Dependency Injection** - built-in dependency container
- 📝 **Simple API** - intuitive interface
- 🎯 **Middleware Support** - comprehensive middleware system
- 📊 **Event System** - pub/sub event bus
- 🔄 **Hot Reload** - plugin hot reloading support

## Quick Start

### Installation

```bash
go mod init your-project
go get github.com/GorgoFramework/gorgo@latest
```

### Simple Example

```go
package main

import (
    "log"
    "github.com/GorgoFramework/gorgo/pkg/gorgo"
)

func main() {
    app := gorgo.New()

    app.Get("/", func(ctx *gorgo.Context) error {
        return ctx.JSON(gorgo.Map{"message": "Hello, World!"})
    })

    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Enhanced Plugin System

Gorgo now features a powerful plugin system with:

- **Lifecycle Hooks** - OnBeforeInit, OnAfterInit, OnBeforeStart, etc.
- **Event Bus** - Subscribe to and publish events
- **Middleware Providers** - Plugins can provide middleware
- **Service Registration** - Automatic dependency injection
- **Hot Reloading** - Runtime configuration updates
- **Priority & Dependencies** - Controlled loading order

### Plugin Example

```go
package main

import (
    "log"
    "github.com/GorgoFramework/gorgo/pkg/gorgo"
    "github.com/GorgoFramework/gorgo/plugins/sql"
    "github.com/GorgoFramework/gorgo/plugins/monitoring"
)

func main() {
    app := gorgo.New()

    // Add plugins
    app.AddPlugin(sql.NewSqlPlugin()).
        AddPlugin(monitoring.NewMonitoringPlugin())

    // Enable built-in middleware
    app.EnableCORS().
        EnableRateLimit(gorgo.RateLimitOptions{
            RequestsPerMinute: 100,
            BurstSize:         10,
        })

    app.Get("/", func(ctx *gorgo.Context) error {
        // Access plugin services
        if db, exists := ctx.GetService("sql"); exists {
            // Use database
        }
        
        return ctx.JSON(gorgo.Map{"message": "Hello with plugins!"})
    })

    log.Fatal(app.Run())
}
```

## Routing with Parameters

Gorgo supports extracting parameters from URLs:

```go
// Simple parameter
app.Get("/hello/:name", func(ctx *gorgo.Context) error {
    name := ctx.Param("name")
    return ctx.String(fmt.Sprintf("Hello, %s!", name))
})

// Multiple parameters
app.Get("/users/:userId/posts/:postId", func(ctx *gorgo.Context) error {
    userId := ctx.Param("userId")
    postId := ctx.Param("postId")
    
    return ctx.JSON(gorgo.Map{
        "userId": userId,
        "postId": postId,
    })
})
```

### Parameter Methods

- `ctx.Param(key)` - get parameter value
- `ctx.ParamDefault(key, default)` - get parameter with default value
- `ctx.HasParam(key)` - check if parameter exists
- `ctx.Params()` - get all parameters

## HTTP Methods

```go
app.Get("/users", getUsersHandler)
app.Post("/users", createUserHandler)
app.Put("/users/:id", updateUserHandler)
app.Delete("/users/:id", deleteUserHandler)
app.Patch("/users/:id", patchUserHandler)
```

## Middleware System

### Built-in Middleware

```go
// Enable CORS
app.EnableCORS(gorgo.CORSOptions{
    AllowOrigin: "*",
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
})

// Rate limiting
app.EnableRateLimit(gorgo.RateLimitOptions{
    RequestsPerMinute: 100,
    BurstSize: 10,
})

// Custom middleware
app.Use(gorgo.LoggerMiddleware())
app.Use(gorgo.RecoveryMiddleware())
```

### Route-specific Middleware

```go
app.Get("/protected", protectedHandler, 
    gorgo.AuthMiddleware(authFunc),
    customMiddleware(),
)
```

### Route Groups

```go
api := app.Group("/api/v1", 
    gorgo.AuthMiddleware(authFunc),
    gorgo.LoggerMiddleware(),
)

api.Get("/users", getUsersHandler)
api.Post("/users", createUserHandler)
```

## Responses

### JSON Response

```go
app.Get("/api/data", func(ctx *gorgo.Context) error {
    return ctx.JSON(gorgo.Map{
        "status": "success",
        "data":   []string{"item1", "item2"},
    })
})
```

### Other Response Types

```go
// Text response
return ctx.String("Hello, World!")

// HTML response
return ctx.HTML("<h1>Hello, World!</h1>")

// Status with chaining
return ctx.Status(201).JSON(gorgo.Map{"created": true})

// Headers
return ctx.Header("X-Custom", "value").JSON(data)
```

## Event System

```go
// Subscribe to events in plugins
func (p *MyPlugin) GetEventSubscriptions() map[string]gorgo.EventHandler {
    return map[string]gorgo.EventHandler{
        "request.completed": p.onRequestCompleted,
        "custom.event": p.onCustomEvent,
    }
}

// Publish events
app.Get("/trigger", func(ctx *gorgo.Context) error {
    eventBus := app.GetEventBus()
    return eventBus.Publish(ctx.FastHTTP(), "custom.event", gorgo.Map{
        "user_id": ctx.Query("user_id"),
    })
})
```

## Built-in Plugins

### SQL Plugin
- PostgreSQL connection pooling
- Transaction middleware
- Hot reloadable configuration

### Redis Plugin  
- Caching middleware
- Session management
- Connection pooling

### Monitoring Plugin
- Request metrics collection
- Performance monitoring
- Health check endpoints

## Configuration

Create a `config/app.toml` file:

```toml
[app]
name = "My App"
version = "1.0.0"
debug = true

[server]
host = "localhost"
port = 8080

[plugins.sql]
host = "localhost"
port = 5432
user = "postgres"
password = "password"
db = "myapp"
max_conns = 25

[plugins.monitoring]
enabled = true
report_interval = 60
log_requests = true
```

## Examples

In the `examples/` directory you'll find various usage examples:

- `hello_world_example/` - basic example
- `echo_example/` - URL parameter example
- `params_example/` - extended parameter examples
- `advanced_params_example/` - advanced parameter handling
- `postgres_example/` - PostgreSQL integration
- `advanced_plugins_example/` - enhanced plugin system demo

## Documentation

- [URL Parameters](docs/url-parameters.md) - detailed guide for working with URL parameters
- [SQL Plugin](docs/sql-plugin.md) - PostgreSQL database integration guide
- [Enhanced Plugin System](docs/enhanced-plugin-system.md) - comprehensive plugin development guide

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

We welcome contributions to the project! Please create issues and pull requests.

## Support

If you have questions or suggestions, create an issue in the repository.