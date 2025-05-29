# Enhanced Plugin System for Gorgo Framework

Gorgo Framework now supports a powerful and flexible plugin system with multiple capabilities for extending functionality.

## Core Features

### 1. Lifecycle Hooks
Plugins can hook into various stages of the application lifecycle:

```go
type LifecycleHooks interface {
    OnBeforeInit(ctx context.Context) error
    OnAfterInit(ctx context.Context) error
    OnBeforeStart(ctx context.Context) error
    OnAfterStart(ctx context.Context) error
    OnBeforeStop(ctx context.Context) error
    OnAfterStop(ctx context.Context) error
}
```

### 2. Event System (Event Bus)
Plugins can subscribe to events and publish their own:

```go
type EventSubscriber interface {
    GetEventSubscriptions() map[string]EventHandler
}

// Built-in events:
// - app.starting, app.stopping
// - server.started
// - request.incoming, request.completed, request.error, request.not_found
// - plugin.started, plugin.stopped
```

### 3. Middleware System
Plugins can provide middleware:

```go
type MiddlewareProvider interface {
    GetMiddleware() []MiddlewareFunc
}
```

### 4. Dependency Management
Plugins can register services in the dependency container:

```go
type ServiceProvider interface {
    GetServices() map[string]interface{}
}
```

### 5. Plugin Configuration
Plugins can validate and provide default configuration:

```go
type ConfigurablePlugin interface {
    ValidateConfig(config map[string]interface{}) error
    GetDefaultConfig() map[string]interface{}
}
```

### 6. Hot Reload
Plugins can support hot configuration reload:

```go
type HotReloadable interface {
    CanHotReload() bool
    OnHotReload(newConfig map[string]interface{}) error
}
```

### 7. Priorities and Dependencies
Plugins are loaded in order of priority and dependencies:

```go
type PluginMetadata struct {
    Name         string
    Version      string
    Description  string
    Author       string
    Dependencies []string
    Priority     PluginPriority
    Tags         []string
}
```

## Creating a Plugin

### Basic Plugin

```go
package myplugin

import (
    "context"
    "github.com/GorgoFramework/gorgo/pkg/gorgo"
    "github.com/GorgoFramework/gorgo/internal/container"
)

type MyPlugin struct {
    gorgo.BasePlugin
}

func NewMyPlugin() *MyPlugin {
    metadata := gorgo.PluginMetadata{
        Name:        "myplugin",
        Version:     "1.0.0",
        Description: "My awesome plugin",
        Author:      "Your Name",
        Priority:    gorgo.PriorityNormal,
        Tags:        []string{"utility"},
    }

    return &MyPlugin{
        BasePlugin: gorgo.NewBasePlugin(metadata),
    }
}

func (p *MyPlugin) Initialize(container *container.Container, config map[string]interface{}) error {
    // Plugin initialization
    return p.BasePlugin.Initialize(container, config)
}

func (p *MyPlugin) Start(ctx context.Context) error {
    // Plugin startup
    return p.BasePlugin.Start(ctx)
}

func (p *MyPlugin) Stop(ctx context.Context) error {
    // Plugin shutdown
    return p.BasePlugin.Stop(ctx)
}
```

### Plugin with Middleware

```go
func (p *MyPlugin) GetMiddleware() []gorgo.MiddlewareFunc {
    return []gorgo.MiddlewareFunc{
        func(next gorgo.HandlerFunc) gorgo.HandlerFunc {
            return func(ctx *gorgo.Context) error {
                // Middleware logic
                return next(ctx)
            }
        },
    }
}
```

### Plugin with Events

```go
func (p *MyPlugin) GetEventSubscriptions() map[string]gorgo.EventHandler {
    return map[string]gorgo.EventHandler{
        "request.completed": p.onRequestCompleted,
        "custom.event":      p.onCustomEvent,
    }
}

func (p *MyPlugin) onRequestCompleted(event *gorgo.Event) error {
    // Event handling
    return nil
}
```

### Plugin with Services

```go
func (p *MyPlugin) GetServices() map[string]interface{} {
    return map[string]interface{}{
        "myservice": p.myService,
        "myconfig":  p.config,
    }
}
```

## Using Plugins

### Plugin Registration

```go
app := gorgo.New()

// Add plugins
app.AddPlugin(sql.NewSqlPlugin()).
    AddPlugin(redis.NewRedisPlugin()).
    AddPlugin(monitoring.NewMonitoringPlugin())
```

### Accessing Plugin Services

```go
app.Get("/users", func(ctx *gorgo.Context) error {
    // Get service from plugin
    if db, exists := ctx.GetService("sql"); exists {
        // Use database
    }
    
    // Get the plugin itself
    if plugin, exists := ctx.GetPlugin("sql"); exists {
        // Use plugin methods
    }
    
    return ctx.JSON(gorgo.Map{"users": []string{}})
})
```

### Publishing Events

```go
app.Get("/trigger", func(ctx *gorgo.Context) error {
    eventBus := app.GetEventBus()
    
    err := eventBus.Publish(ctx.FastHTTP(), "custom.event", gorgo.Map{
        "user_id": ctx.Query("user_id"),
        "action":  "custom_action",
    })
    
    return ctx.JSON(gorgo.Map{"success": err == nil})
})
```

### Hot Reload

```go
app.Post("/admin/reload/:plugin", func(ctx *gorgo.Context) error {
    pluginName := ctx.Param("plugin")
    newConfig := map[string]interface{}{
        "enabled": true,
    }
    
    err := app.HotReloadPlugin(pluginName, newConfig)
    return ctx.JSON(gorgo.Map{"success": err == nil})
})
```

## Built-in Plugins

### SQL Plugin
- PostgreSQL connection
- Connection pool
- Transaction middleware
- Hot reload

### Redis Plugin
- Caching
- Sessions
- Automatic caching middleware

### Monitoring Plugin
- Metrics collection
- Request logging
- Periodic reports
- Metrics endpoint

## Configuration

```toml
[plugins.sql]
host = "localhost"
port = 5432
user = "postgres"
password = "password"
db = "myapp"
max_conns = 25
min_conns = 5

[plugins.redis]
host = "localhost"
port = 6379
password = ""
db = 0
pool_size = 10

[plugins.monitoring]
enabled = true
report_interval = 60
log_requests = true
```

## Best Practices

1. **Use priorities** for proper loading order
2. **Declare dependencies** between plugins
3. **Validate configuration** in the `ValidateConfig` method
4. **Handle errors** in lifecycle hooks
5. **Use events** for loose coupling
6. **Provide default configuration**
7. **Document your plugin's API**

## Examples

See the `examples/advanced_plugins_example/` folder for a complete example of using the enhanced plugin system. 