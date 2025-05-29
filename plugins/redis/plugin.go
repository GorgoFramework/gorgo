package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/GorgoFramework/gorgo/internal/container"
	"github.com/GorgoFramework/gorgo/pkg/gorgo"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

type RedisPlugin struct {
	gorgo.BasePlugin
	client *redis.Client
	config RedisConfig
}

type RedisConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
	PoolSize int    `toml:"pool_size"`
}

func NewRedisPlugin() *RedisPlugin {
	metadata := gorgo.PluginMetadata{
		Name:        "redis",
		Version:     "1.0.0",
		Description: "Redis caching and session storage plugin",
		Author:      "Gorgo Framework",
		Priority:    gorgo.PriorityNormal,
		Tags:        []string{"cache", "redis", "session"},
	}

	return &RedisPlugin{
		BasePlugin: gorgo.NewBasePlugin(metadata),
	}
}

// ConfigurablePlugin implementation
func (p *RedisPlugin) ValidateConfig(config map[string]interface{}) error {
	host, _ := config["host"].(string)
	if host == "" {
		return fmt.Errorf("host is required")
	}
	return nil
}

func (p *RedisPlugin) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"host":      "localhost",
		"port":      6379,
		"password":  "",
		"db":        0,
		"pool_size": 10,
	}
}

// ServiceProvider implementation
func (p *RedisPlugin) GetServices() map[string]interface{} {
	return map[string]interface{}{
		"redis":    p.client,
		"cache":    p.client,
		"rediscfg": p.config,
	}
}

// EventSubscriber implementation
func (p *RedisPlugin) GetEventSubscriptions() map[string]gorgo.EventHandler {
	return map[string]gorgo.EventHandler{
		"request.completed": p.onRequestCompleted,
		"app.stopping":      p.onAppStopping,
	}
}

func (p *RedisPlugin) onRequestCompleted(event *gorgo.Event) error {
	// Can add logic for caching responses
	return nil
}

func (p *RedisPlugin) onAppStopping(event *gorgo.Event) error {
	log.Println("Redis Plugin: Application stopping, clearing temporary cache...")
	return nil
}

// MiddlewareProvider implementation
func (p *RedisPlugin) GetMiddleware() []gorgo.MiddlewareFunc {
	return []gorgo.MiddlewareFunc{
		p.cacheMiddleware(),
	}
}

func (p *RedisPlugin) cacheMiddleware() gorgo.MiddlewareFunc {
	return func(next gorgo.HandlerFunc) gorgo.HandlerFunc {
		return func(ctx *gorgo.Context) error {
			// Simple caching for GET requests
			if ctx.Method() == "GET" {
				cacheKey := fmt.Sprintf("cache:%s", ctx.Path())

				// Check cache
				cached, err := p.client.Get(context.Background(), cacheKey).Result()
				if err == nil {
					ctx.Header("X-Cache", "HIT")
					return ctx.String(cached)
				}
			}

			// Execute handler
			return next(ctx)
		}
	}
}

// HotReloadable implementation
func (p *RedisPlugin) CanHotReload() bool {
	return true
}

func (p *RedisPlugin) OnHotReload(newConfig map[string]interface{}) error {
	log.Println("Redis Plugin: Hot reloading configuration...")
	// Here you can update settings without reconnection
	return nil
}

// Main plugin methods
func (p *RedisPlugin) Initialize(container *container.Container, config map[string]interface{}) error {
	p.config = RedisConfig{
		Host:     getStringConfig(config, "host", "localhost"),
		Port:     getIntConfig(config, "port", 6379),
		Password: getStringConfig(config, "password", ""),
		DB:       getIntConfig(config, "db", 0),
		PoolSize: getIntConfig(config, "pool_size", 10),
	}

	// Create Redis client
	p.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", p.config.Host, p.config.Port),
		Password: p.config.Password,
		DB:       p.config.DB,
		PoolSize: p.config.PoolSize,
	})

	return p.BasePlugin.Initialize(container, config)
}

func (p *RedisPlugin) Start(ctx context.Context) error {
	// Check connection
	if err := p.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Println("Redis Plugin: Connected successfully")
	return p.BasePlugin.Start(ctx)
}

func (p *RedisPlugin) Stop(ctx context.Context) error {
	if p.client != nil {
		if err := p.client.Close(); err != nil {
			log.Printf("Error closing Redis client: %v", err)
		}
	}

	return p.BasePlugin.Stop(ctx)
}

// Additional methods
func (p *RedisPlugin) GetClient() *redis.Client {
	return p.client
}

func (p *RedisPlugin) Set(key string, value interface{}, expiration time.Duration) error {
	return p.client.Set(context.Background(), key, value, expiration).Err()
}

func (p *RedisPlugin) Get(key string) (string, error) {
	return p.client.Get(context.Background(), key).Result()
}

func (p *RedisPlugin) Delete(key string) error {
	return p.client.Del(context.Background(), key).Err()
}

// Session middleware
func (p *RedisPlugin) SessionMiddleware(sessionName string) gorgo.MiddlewareFunc {
	return func(next gorgo.HandlerFunc) gorgo.HandlerFunc {
		return func(ctx *gorgo.Context) error {
			sessionID := ctx.GetCookie(sessionName)
			if sessionID == "" {
				// Create new session
				sessionID = generateSessionID()
				cookie := &fasthttp.Cookie{}
				cookie.SetKey(sessionName)
				cookie.SetValue(sessionID)
				cookie.SetHTTPOnly(true)
				ctx.Cookie(cookie)
			}

			// Load session data
			sessionData, err := p.Get(fmt.Sprintf("session:%s", sessionID))
			if err == nil {
				ctx.Set("session_data", sessionData)
			}
			ctx.Set("session_id", sessionID)

			return next(ctx)
		}
	}
}

// Helper functions
func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if value, ok := config[key].(string); ok {
		return value
	}
	return defaultValue
}

func getIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	if value, ok := config[key].(int); ok {
		return value
	}
	if value, ok := config[key].(float64); ok {
		return int(value)
	}
	return defaultValue
}

func generateSessionID() string {
	// Simple session ID generation (use crypto/rand in production)
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}
