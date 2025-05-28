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
type Map map[string]interface{}

type Application struct {
	container *container.Container
	plugins   map[string]Plugin
	config    Config
	server    *fasthttp.Server
	router    *Router
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
		container: container.NewContainer(),
		plugins:   make(map[string]Plugin),
		server:    &fasthttp.Server{},
		router:    NewRouter(),
	}

	app.loadConfig()

	app.printBanner()

	return app
}

func (a *Application) loadConfig() {
	a.config = Config{}

	a.config.App.Name = "Gorgo Framework"
	a.config.App.Version = "0.0.3"
	a.config.App.Debug = true

	a.config.Server.Host = "localhost"
	a.config.Server.Port = 8080

	// TODO: Add custom config path
	if _, err := os.Stat("config/app.toml"); err == nil {
		if _, err := toml.DecodeFile("config/app.toml", &a.config); err != nil {
			log.Printf("Warning: failed to load config/app.toml: %v", err)
		}
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

func (a *Application) AddPlugin(plugin Plugin) *Application {
	a.plugins[plugin.Name()] = plugin
	return a
}

func (a *Application) Run() error {
	for name, plugin := range a.plugins {
		pluginConfig := a.config.Plugins[name]
		if pluginConfig == nil {
			pluginConfig = make(map[string]interface{})
		}

		if err := plugin.Initialize(a.container, pluginConfig); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %v", name, err)
		}

		log.Printf("Plugin %s initialized successfully", name)
	}

	a.server = &fasthttp.Server{
		Handler: a.handleRequest,
	}

	go func() {
		addr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)
		log.Printf("Server starting on %s", addr)
		if err := a.server.ListenAndServe(addr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	a.waitForShutdown()

	return nil
}

func (a *Application) handleRequest(ctx *fasthttp.RequestCtx) {
	gorgoCtx := NewContext(ctx, a.container, a.plugins)

	method := string(ctx.Method())
	path := string(ctx.Path())

	handler, params := a.router.FindHandler(method, path)
	if handler == nil {
		ctx.SetStatusCode(404)
		ctx.SetBodyString("Not Found")
		return
	}

	// Set URL parameters in context
	if params != nil {
		for key, value := range params {
			gorgoCtx.SetParam(key, value)
		}
	}

	if err := handler(gorgoCtx); err != nil {
		log.Printf("Handler error: %v", err)
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Internal Server Error")
		return
	}
}

func (a *Application) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	if len(a.plugins) > 0 {
		for name, plugin := range a.plugins {
			if err := plugin.Shutdown(); err != nil {
				log.Printf("Error shutting down plugin %s: %v", name, err)
			}
		}
	}

	if err := a.server.ShutdownWithContext(context.Background()); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	log.Println("Server stopped")
}

func (a *Application) Get(path string, handler HandlerFunc) {
	a.router.AddRoute("GET", path, handler)
}

func (a *Application) Post(path string, handler HandlerFunc) {
	a.router.AddRoute("POST", path, handler)
}

func (a *Application) Put(path string, handler HandlerFunc) {
	a.router.AddRoute("PUT", path, handler)
}

func (a *Application) Delete(path string, handler HandlerFunc) {
	a.router.AddRoute("DELETE", path, handler)
}

func (a *Application) Patch(path string, handler HandlerFunc) {
	a.router.AddRoute("PATCH", path, handler)
}
