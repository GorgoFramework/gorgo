package gorgo

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/GorgoFramework/gorgo/internal/container"
	"github.com/valyala/fasthttp"
)

type HandlerFunc func(ctx *Context) error
type Map map[string]any

type Application struct {
	container       *container.Container
	pluginManager   *PluginManager
	config          Config
	server          *fasthttp.Server
	router          *Router
	middlewareChain *MiddlewareChain
}

type Config struct {
	App struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
		Debug   bool   `toml:"debug"`
	} `toml:"app"`

	Server struct {
		Host string `toml:"host"`
		Port int    `toml:"port"`
	} `toml:"server"`

	Plugins map[string]map[string]interface{} `toml:"plugins"`
}

func New() *Application {
	app := &Application{
		container:       container.NewContainer(),
		config:          Config{},
		router:          NewRouter(),
		middlewareChain: NewMiddlewareChain(),
	}

	app.pluginManager = NewPluginManager(app.container)

	app.loadConfig()
	app.setupDefaultMiddleware()
	app.printBanner()

	return app
}

func (a *Application) loadConfig() {
	// Set defaults
	a.config.App.Name = "Gorgo Application"
	a.config.App.Version = "1.0.0"
	a.config.App.Debug = false
	a.config.Server.Host = "localhost"
	a.config.Server.Port = 3000

	// TODO: Add custom config path
	if _, err := os.Stat("config/app.toml"); err == nil {
		if _, err := toml.DecodeFile("config/app.toml", &a.config); err != nil {
			log.Printf("Warning: failed to load config/app.toml: %v", err)
		}
	}
}

func (a *Application) setupDefaultMiddleware() {
	// Add basic middleware
	a.middlewareChain.Add(RecoveryMiddleware())

	if a.config.App.Debug {
		a.middlewareChain.Add(LoggerMiddleware())
	}
}

func (a *Application) printBanner() {
	banner := `
 ██████╗  ██████╗ ██████╗  ██████╗  ██████╗ 
██╔════╝ ██╔═══██╗██╔══██╗██╔════╝ ██╔═══██╗
██║  ███╗██║   ██║██████╔╝██║  ███╗██║   ██║
██║   ██║██║   ██║██╔══██╗██║   ██║██║   ██║
╚██████╔╝╚██████╔╝██║  ██║╚██████╔╝╚██████╔╝
 ╚═════╝  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝  ╚═════╝ 

%s v%s
Powered by Gorgo Framework
`
	fmt.Printf(banner, a.config.App.Name, a.config.App.Version)
}

// Methods for working with plugins
func (a *Application) AddPlugin(plugin Plugin) *Application {
	if err := a.pluginManager.RegisterPlugin(plugin); err != nil {
		log.Printf("Failed to register plugin: %v", err)
	}
	return a
}

func (a *Application) GetPlugin(name string) (Plugin, bool) {
	return a.pluginManager.GetPlugin(name)
}

func (a *Application) GetEventBus() *EventBus {
	return a.pluginManager.GetEventBus()
}

func (a *Application) HotReloadPlugin(name string, newConfig map[string]interface{}) error {
	return a.pluginManager.HotReloadPlugin(name, newConfig)
}

// Methods for working with middleware
func (a *Application) Use(middleware MiddlewareFunc) *Application {
	a.middlewareChain.Add(middleware)
	return a
}

func (a *Application) UseMiddleware(middlewares ...MiddlewareFunc) *Application {
	for _, middleware := range middlewares {
		a.middlewareChain.Add(middleware)
	}
	return a
}

// CORS methods
func (a *Application) EnableCORS(options ...CORSOptions) *Application {
	var corsOptions CORSOptions
	if len(options) > 0 {
		corsOptions = options[0]
	} else {
		corsOptions = DefaultCORSOptions()
	}

	a.middlewareChain.Add(CORSMiddleware(corsOptions))
	return a
}

// Rate limiting methods
func (a *Application) EnableRateLimit(options RateLimitOptions) *Application {
	a.middlewareChain.Add(RateLimitMiddleware(options))
	return a
}

// Authentication methods
func (a *Application) EnableAuth(authFunc func(ctx *Context) (interface{}, error)) *Application {
	a.middlewareChain.Add(AuthMiddleware(authFunc))
	return a
}

func (a *Application) Run() error {
	// Initialize plugins
	if err := a.pluginManager.InitializePlugins(a.config.Plugins); err != nil {
		return fmt.Errorf("failed to initialize plugins: %v", err)
	}

	// Add middleware from plugins
	pluginMiddleware := a.pluginManager.GetMiddleware()
	for _, middleware := range pluginMiddleware {
		a.middlewareChain.Add(middleware)
	}

	// Start plugins
	ctx := context.Background()
	if err := a.pluginManager.StartPlugins(ctx); err != nil {
		return fmt.Errorf("failed to start plugins: %v", err)
	}

	// Publish application starting event
	a.pluginManager.GetEventBus().Publish(ctx, "app.starting", map[string]interface{}{
		"config": a.config,
	})

	a.server = &fasthttp.Server{
		Handler: a.handleRequest,
	}

	go func() {
		addr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)
		log.Printf("Server starting on %s", addr)

		// Publish server started event
		a.pluginManager.GetEventBus().Publish(ctx, "server.started", map[string]interface{}{
			"address": addr,
		})

		if err := a.server.ListenAndServe(addr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	a.waitForShutdown()

	return nil
}

func (a *Application) handleRequest(ctx *fasthttp.RequestCtx) {
	gorgoCtx := NewContext(ctx, a.container, a.pluginManager.plugins)

	method := string(ctx.Method())
	path := string(ctx.Path())

	// Publish incoming request event
	a.pluginManager.GetEventBus().Publish(context.Background(), "request.incoming", map[string]interface{}{
		"method": method,
		"path":   path,
		"ip":     gorgoCtx.ClientIP(),
	})

	handler, params := a.router.FindHandler(method, path)
	if handler == nil {
		ctx.SetStatusCode(404)
		ctx.SetBodyString("Not Found")

		// Publish 404 event
		a.pluginManager.GetEventBus().Publish(context.Background(), "request.not_found", map[string]interface{}{
			"method": method,
			"path":   path,
		})
		return
	}

	// Set URL parameters in context
	for key, value := range params {
		gorgoCtx.SetParam(key, value)
	}

	// Apply middleware chain
	finalHandler := a.middlewareChain.Execute(handler)

	if err := finalHandler(gorgoCtx); err != nil {
		log.Printf("Handler error: %v", err)
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Internal Server Error")

		// Publish error event
		a.pluginManager.GetEventBus().Publish(context.Background(), "request.error", map[string]interface{}{
			"method": method,
			"path":   path,
			"error":  err.Error(),
		})
		return
	}

	// Publish successful request event
	a.pluginManager.GetEventBus().Publish(context.Background(), "request.completed", map[string]interface{}{
		"method": method,
		"path":   path,
		"status": ctx.Response.StatusCode(),
	})
}

func (a *Application) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx := context.Background()

	// Publish application stopping event
	a.pluginManager.GetEventBus().Publish(ctx, "app.stopping", map[string]interface{}{})

	// Stop plugins
	if err := a.pluginManager.StopPlugins(ctx); err != nil {
		log.Printf("Error stopping plugins: %v", err)
	}

	if err := a.server.ShutdownWithContext(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	log.Println("Server stopped")
}

// HTTP methods with route-level middleware support
func (a *Application) Get(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	finalHandler := a.applyRouteMiddleware(handler, middleware...)
	a.router.AddRoute("GET", path, finalHandler)
}

func (a *Application) Post(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	finalHandler := a.applyRouteMiddleware(handler, middleware...)
	a.router.AddRoute("POST", path, finalHandler)
}

func (a *Application) Put(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	finalHandler := a.applyRouteMiddleware(handler, middleware...)
	a.router.AddRoute("PUT", path, finalHandler)
}

func (a *Application) Delete(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	finalHandler := a.applyRouteMiddleware(handler, middleware...)
	a.router.AddRoute("DELETE", path, finalHandler)
}

func (a *Application) Patch(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	finalHandler := a.applyRouteMiddleware(handler, middleware...)
	a.router.AddRoute("PATCH", path, finalHandler)
}

func (a *Application) applyRouteMiddleware(handler HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	if len(middleware) == 0 {
		return handler
	}

	chain := NewMiddlewareChain(middleware...)
	return chain.Execute(handler)
}

// Route grouping
type RouteGroup struct {
	app        *Application
	prefix     string
	middleware []MiddlewareFunc
}

func (a *Application) Group(prefix string, middleware ...MiddlewareFunc) *RouteGroup {
	return &RouteGroup{
		app:        a,
		prefix:     prefix,
		middleware: middleware,
	}
}

func (rg *RouteGroup) Get(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	fullPath := rg.prefix + path
	allMiddleware := append(rg.middleware, middleware...)
	rg.app.Get(fullPath, handler, allMiddleware...)
}

func (rg *RouteGroup) Post(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	fullPath := rg.prefix + path
	allMiddleware := append(rg.middleware, middleware...)
	rg.app.Post(fullPath, handler, allMiddleware...)
}

func (rg *RouteGroup) Put(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	fullPath := rg.prefix + path
	allMiddleware := append(rg.middleware, middleware...)
	rg.app.Put(fullPath, handler, allMiddleware...)
}

func (rg *RouteGroup) Delete(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	fullPath := rg.prefix + path
	allMiddleware := append(rg.middleware, middleware...)
	rg.app.Delete(fullPath, handler, allMiddleware...)
}

func (rg *RouteGroup) Patch(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	fullPath := rg.prefix + path
	allMiddleware := append(rg.middleware, middleware...)
	rg.app.Patch(fullPath, handler, allMiddleware...)
}
