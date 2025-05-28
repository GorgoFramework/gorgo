# Gorgo Framework

A modern and fast web framework for Go, built on top of FastHTTP.

## Features

- 🚀 **High Performance** - built on FastHTTP
- 🛣️ **Flexible Routing** - URL parameter support
- 🔌 **Plugin System** - easily extensible architecture
- 🏗️ **Dependency Injection** - built-in dependency container
- 📝 **Simple API** - intuitive interface

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

### Text Response

```go
app.Get("/text", func(ctx *gorgo.Context) error {
    return ctx.String("Hello, World!")
})
```

## Plugin System

Gorgo supports modular architecture through a plugin system:

```go
import "github.com/GorgoFramework/gorgo/plugins/sql"

func main() {
    sqlPlugin := sql.NewSqlPlugin()
    app := gorgo.New().AddPlugin(sqlPlugin)
    
    app.Get("/users", func(ctx *gorgo.Context) error {
        db, _ := ctx.GetService("sql")
        // Database operations
        return ctx.JSON(gorgo.Map{"users": []string{}})
    })
    
    app.Run()
}
```

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
```

## Examples

In the `examples/` directory you'll find various usage examples:

- `hello_world_example/` - basic example
- `echo_example/` - URL parameter example
- `params_example/` - extended parameter examples
- `advanced_params_example/` - advanced parameter handling
- `postgres_example/` - PostgreSQL integration

## Documentation

- [URL Parameters](docs/url-parameters.md) - detailed guide for working with URL parameters

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

We welcome contributions to the project! Please create issues and pull requests.

## Support

If you have questions or suggestions, create an issue in the repository.