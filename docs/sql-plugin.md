# SQL Plugin

The SQL plugin provides PostgreSQL database integration for Gorgo Framework using the high-performance `pgx` driver with connection pooling.

## Features

- **Connection Pooling** - Built-in connection pool management using `pgxpool`
- **PostgreSQL Support** - Full PostgreSQL compatibility via `pgx/v5`
- **Automatic Cleanup** - Graceful connection pool shutdown
- **Configuration-based** - Easy setup through TOML configuration
- **Dependency Injection** - Seamless integration with Gorgo's DI container

## Installation

The SQL plugin is included with Gorgo Framework. No additional installation is required.

## Configuration

Add the SQL plugin configuration to your `config/app.toml` file:

```toml
[plugins.sql]
host = "localhost"
port = 5432
user = "postgres"
password = "your_password"
db = "your_database"
```

### Configuration Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `host` | string | Yes | - | PostgreSQL server hostname |
| `port` | int | No | 5432 | PostgreSQL server port |
| `user` | string | Yes | - | Database username |
| `password` | string | Yes | - | Database password |
| `db` | string | Yes | - | Database name |

## Usage

### Basic Setup

```go
package main

import (
    "log"
    "github.com/GorgoFramework/gorgo/pkg/gorgo"
    "github.com/GorgoFramework/gorgo/plugins/sql"
)

func main() {
    // Create and add SQL plugin
    sqlPlugin := sql.NewSqlPlugin()
    app := gorgo.New().AddPlugin(sqlPlugin)
    
    // Your routes here
    app.Get("/users", getUsersHandler)
    
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### Accessing the Database

The SQL plugin registers a connection pool in the dependency container under the name `"sql"`. You can access it in your handlers:

```go
app.Get("/users", func(ctx *gorgo.Context) error {
    // Get the database service
    db, exists := ctx.GetService("sql")
    if !exists {
        return ctx.JSON(gorgo.Map{"error": "Database not available"})
    }
    
    // Cast to pgxpool.Pool
    pool := db.(*pgxpool.Pool)
    
    // Use the pool for database operations
    // ... your database logic here
    
    return ctx.JSON(gorgo.Map{"status": "success"})
})
```

## Database Operations

### Query Single Row

```go
app.Get("/user/:id", func(ctx *gorgo.Context) error {
    db, _ := ctx.GetService("sql")
    pool := db.(*pgxpool.Pool)
    
    userID := ctx.Param("id")
    
    var username string
    var email string
    
    query := "SELECT username, email FROM users WHERE id = $1"
    err := pool.QueryRow(context.Background(), query, userID).Scan(&username, &email)
    if err != nil {
        return ctx.JSON(gorgo.Map{"error": err.Error()})
    }
    
    return ctx.JSON(gorgo.Map{
        "id": userID,
        "username": username,
        "email": email,
    })
})
```

### Query Multiple Rows

```go
app.Get("/users", func(ctx *gorgo.Context) error {
    db, _ := ctx.GetService("sql")
    pool := db.(*pgxpool.Pool)
    
    rows, err := pool.Query(context.Background(), "SELECT id, username, email FROM users")
    if err != nil {
        return ctx.JSON(gorgo.Map{"error": err.Error()})
    }
    defer rows.Close()
    
    var users []map[string]interface{}
    
    for rows.Next() {
        var id int
        var username, email string
        
        if err := rows.Scan(&id, &username, &email); err != nil {
            return ctx.JSON(gorgo.Map{"error": err.Error()})
        }
        
        users = append(users, map[string]interface{}{
            "id": id,
            "username": username,
            "email": email,
        })
    }
    
    return ctx.JSON(gorgo.Map{"users": users})
})
```

### Insert Data

```go
app.Post("/users", func(ctx *gorgo.Context) error {
    db, _ := ctx.GetService("sql")
    pool := db.(*pgxpool.Pool)
    
    // Parse request body (you might want to add proper JSON parsing)
    username := string(ctx.PostArgs().Peek("username"))
    email := string(ctx.PostArgs().Peek("email"))
    
    var userID int
    query := "INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id"
    err := pool.QueryRow(context.Background(), query, username, email).Scan(&userID)
    if err != nil {
        return ctx.JSON(gorgo.Map{"error": err.Error()})
    }
    
    return ctx.JSON(gorgo.Map{
        "id": userID,
        "username": username,
        "email": email,
        "message": "User created successfully",
    })
})
```

### Update Data

```go
app.Put("/users/:id", func(ctx *gorgo.Context) error {
    db, _ := ctx.GetService("sql")
    pool := db.(*pgxpool.Pool)
    
    userID := ctx.Param("id")
    username := string(ctx.PostArgs().Peek("username"))
    email := string(ctx.PostArgs().Peek("email"))
    
    query := "UPDATE users SET username = $1, email = $2 WHERE id = $3"
    result, err := pool.Exec(context.Background(), query, username, email, userID)
    if err != nil {
        return ctx.JSON(gorgo.Map{"error": err.Error()})
    }
    
    if result.RowsAffected() == 0 {
        return ctx.JSON(gorgo.Map{"error": "User not found"})
    }
    
    return ctx.JSON(gorgo.Map{"message": "User updated successfully"})
})
```

### Delete Data

```go
app.Delete("/users/:id", func(ctx *gorgo.Context) error {
    db, _ := ctx.GetService("sql")
    pool := db.(*pgxpool.Pool)
    
    userID := ctx.Param("id")
    
    query := "DELETE FROM users WHERE id = $1"
    result, err := pool.Exec(context.Background(), query, userID)
    if err != nil {
        return ctx.JSON(gorgo.Map{"error": err.Error()})
    }
    
    if result.RowsAffected() == 0 {
        return ctx.JSON(gorgo.Map{"error": "User not found"})
    }
    
    return ctx.JSON(gorgo.Map{"message": "User deleted successfully"})
})
```

## Transactions

```go
app.Post("/transfer", func(ctx *gorgo.Context) error {
    db, _ := ctx.GetService("sql")
    pool := db.(*pgxpool.Pool)
    
    // Begin transaction
    tx, err := pool.Begin(context.Background())
    if err != nil {
        return ctx.JSON(gorgo.Map{"error": err.Error()})
    }
    defer tx.Rollback(context.Background())
    
    // Perform multiple operations
    _, err = tx.Exec(context.Background(), 
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2", 
        100, 1)
    if err != nil {
        return ctx.JSON(gorgo.Map{"error": err.Error()})
    }
    
    _, err = tx.Exec(context.Background(), 
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2", 
        100, 2)
    if err != nil {
        return ctx.JSON(gorgo.Map{"error": err.Error()})
    }
    
    // Commit transaction
    if err = tx.Commit(context.Background()); err != nil {
        return ctx.JSON(gorgo.Map{"error": err.Error()})
    }
    
    return ctx.JSON(gorgo.Map{"message": "Transfer completed successfully"})
})
```

## Error Handling

Always handle database errors appropriately:

```go
app.Get("/user/:id", func(ctx *gorgo.Context) error {
    db, exists := ctx.GetService("sql")
    if !exists {
        ctx.SetStatusCode(500)
        return ctx.JSON(gorgo.Map{"error": "Database service not available"})
    }
    
    pool := db.(*pgxpool.Pool)
    userID := ctx.Param("id")
    
    var username string
    query := "SELECT username FROM users WHERE id = $1"
    err := pool.QueryRow(context.Background(), query, userID).Scan(&username)
    
    if err != nil {
        if err == pgx.ErrNoRows {
            ctx.SetStatusCode(404)
            return ctx.JSON(gorgo.Map{"error": "User not found"})
        }
        
        ctx.SetStatusCode(500)
        return ctx.JSON(gorgo.Map{"error": "Database error"})
    }
    
    return ctx.JSON(gorgo.Map{"username": username})
})
```

## Best Practices

1. **Always use context**: Pass `context.Background()` or a proper context to database operations
2. **Handle errors**: Always check and handle database errors appropriately
3. **Use parameterized queries**: Prevent SQL injection by using `$1`, `$2`, etc. placeholders
4. **Close resources**: Always close rows when iterating over query results
5. **Use transactions**: For operations that require atomicity
6. **Connection pooling**: The plugin automatically manages connection pooling

## Complete Example

See the `examples/postgres_example/` directory for a complete working example with:

- Database configuration
- CRUD operations
- Error handling
- Proper project structure

## Troubleshooting

### Common Issues

1. **Connection refused**: Check if PostgreSQL is running and accessible
2. **Authentication failed**: Verify username and password in configuration
3. **Database does not exist**: Ensure the specified database exists
4. **Permission denied**: Check user permissions for the database

### Debug Mode

Enable debug mode in your configuration to see detailed logs:

```toml
[app]
debug = true
```

## Dependencies

The SQL plugin uses the following dependencies:

- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/jackc/pgx/v5/pgxpool` - Connection pooling

These are automatically included when you use Gorgo Framework. 