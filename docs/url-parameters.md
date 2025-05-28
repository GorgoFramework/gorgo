# URL Parameter Parsing

Gorgo Framework supports extracting parameters from URLs using dynamic route segments.

## Defining Routes with Parameters

Parameters in routes are defined using the `:` prefix:

```go
app.Get("/users/:id", handler)
app.Get("/users/:userId/posts/:postId", handler)
app.Get("/api/:version/users/:id/profile", handler)
```

## Parameter Methods

### `ctx.Param(key string) string`

Gets parameter value by key:

```go
app.Get("/users/:id", func(ctx *gorgo.Context) error {
    id := ctx.Param("id")
    return ctx.String(fmt.Sprintf("User ID: %s", id))
})
```

### `ctx.ParamDefault(key, defaultValue string) string`

Gets parameter value or returns default value if parameter is not found:

```go
app.Get("/users/:id", func(ctx *gorgo.Context) error {
    id := ctx.Param("id")
    format := ctx.ParamDefault("format", "json")
    
    if format == "xml" {
        return ctx.String(fmt.Sprintf("<user><id>%s</id></user>", id))
    }
    
    return ctx.JSON(gorgo.Map{"userId": id})
})
```

### `ctx.HasParam(key string) bool`

Checks if parameter exists:

```go
app.Get("/users/:id", func(ctx *gorgo.Context) error {
    if !ctx.HasParam("id") {
        return ctx.JSON(gorgo.Map{"error": "ID parameter is required"})
    }
    
    id := ctx.Param("id")
    return ctx.JSON(gorgo.Map{"userId": id})
})
```

### `ctx.Params() map[string]string`

Returns all parameters as a map:

```go
app.Get("/users/:userId/posts/:postId", func(ctx *gorgo.Context) error {
    params := ctx.Params()
    
    return ctx.JSON(gorgo.Map{
        "message": "All parameters",
        "params":  params,
    })
})
```

## Usage Examples

### Simple Parameter

```go
app.Get("/hello/:name", func(ctx *gorgo.Context) error {
    name := ctx.Param("name")
    return ctx.String(fmt.Sprintf("Hello, %s!", name))
})

// GET /hello/John -> "Hello, John!"
```

### Multiple Parameters

```go
app.Get("/users/:userId/posts/:postId", func(ctx *gorgo.Context) error {
    userId := ctx.Param("userId")
    postId := ctx.Param("postId")
    
    return ctx.JSON(gorgo.Map{
        "userId": userId,
        "postId": postId,
    })
})

// GET /users/123/posts/456 -> {"userId": "123", "postId": "456"}
```

### Mixed Static and Dynamic Segments

```go
app.Get("/api/v1/users/:id/profile", func(ctx *gorgo.Context) error {
    id := ctx.Param("id")
    
    return ctx.JSON(gorgo.Map{
        "message": "User profile",
        "userId":  id,
        "version": "v1",
    })
})

// GET /api/v1/users/789/profile -> {"message": "User profile", "userId": "789", "version": "v1"}
```

### Parameter Validation

```go
app.Get("/api/:version/users/:userId", func(ctx *gorgo.Context) error {
    requiredParams := []string{"version", "userId"}
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
        "version": ctx.Param("version"),
        "userId":  ctx.Param("userId"),
    })
})
```

## Features

1. **Exact Matching**: The number of segments in the route must exactly match the number of segments in the request
2. **Security**: The `Params()` method returns a copy of the parameters map, preventing external modifications
3. **Performance**: Parameters are extracted only once during route matching
4. **Flexibility**: Support for any number of parameters in a route

## Limitations

- Parameters can only contain one path segment (wildcard parameters are not supported)
- Parameters cannot contain the `/` character
- The order of parameters in the route is important for matching 