package sql

import (
	"context"
	"fmt"
	"log"

	"github.com/GorgoFramework/gorgo/internal/container"
	"github.com/GorgoFramework/gorgo/pkg/gorgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SqlPlugin struct {
	gorgo.BasePlugin
	pool   *pgxpool.Pool
	config SqlConfig
}

type SqlConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"db"`
	MaxConns int    `toml:"max_conns"`
	MinConns int    `toml:"min_conns"`
}

func NewSqlPlugin() *SqlPlugin {
	metadata := gorgo.PluginMetadata{
		Name:        "sql",
		Version:     "1.0.0",
		Description: "PostgreSQL database plugin with connection pooling",
		Author:      "Gorgo Framework",
		Priority:    gorgo.PriorityHigh,
		Tags:        []string{"database", "postgresql", "sql"},
	}

	return &SqlPlugin{
		BasePlugin: gorgo.NewBasePlugin(metadata),
	}
}

// ConfigurablePlugin implementation
func (p *SqlPlugin) ValidateConfig(config map[string]interface{}) error {
	host, _ := config["host"].(string)
	if host == "" {
		return fmt.Errorf("host is required")
	}

	user, _ := config["user"].(string)
	if user == "" {
		return fmt.Errorf("user is required")
	}

	password, _ := config["password"].(string)
	if password == "" {
		return fmt.Errorf("password is required")
	}

	dbName, _ := config["db"].(string)
	if dbName == "" {
		return fmt.Errorf("db is required")
	}

	return nil
}

func (p *SqlPlugin) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"host":      "localhost",
		"port":      5432,
		"user":      "postgres",
		"password":  "",
		"db":        "",
		"max_conns": 25,
		"min_conns": 5,
	}
}

// ServiceProvider implementation
func (p *SqlPlugin) GetServices() map[string]interface{} {
	return map[string]interface{}{
		"sql":    p.pool,
		"db":     p.pool, // Alternative name
		"sqlcfg": p.config,
	}
}

// EventSubscriber implementation
func (p *SqlPlugin) GetEventSubscriptions() map[string]gorgo.EventHandler {
	return map[string]gorgo.EventHandler{
		"app.stopping":      p.onAppStopping,
		"request.completed": p.onRequestCompleted,
	}
}

func (p *SqlPlugin) onAppStopping(event *gorgo.Event) error {
	log.Println("SQL Plugin: Application is stopping, preparing to close connections...")
	return nil
}

func (p *SqlPlugin) onRequestCompleted(event *gorgo.Event) error {
	// Can add logic for monitoring database requests
	return nil
}

// LifecycleHooks implementation
func (p *SqlPlugin) OnBeforeInit(ctx context.Context) error {
	log.Println("SQL Plugin: Preparing to initialize...")
	return nil
}

func (p *SqlPlugin) OnAfterInit(ctx context.Context) error {
	log.Printf("SQL Plugin: Successfully initialized with %d max connections", p.config.MaxConns)
	return nil
}

func (p *SqlPlugin) OnBeforeStart(ctx context.Context) error {
	log.Println("SQL Plugin: Starting database connection monitoring...")
	return nil
}

func (p *SqlPlugin) OnAfterStart(ctx context.Context) error {
	// Check connection
	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	log.Println("SQL Plugin: Database connection verified")
	return nil
}

func (p *SqlPlugin) OnBeforeStop(ctx context.Context) error {
	log.Println("SQL Plugin: Preparing to stop...")
	return nil
}

func (p *SqlPlugin) OnAfterStop(ctx context.Context) error {
	log.Println("SQL Plugin: Successfully stopped")
	return nil
}

// HotReloadable implementation
func (p *SqlPlugin) CanHotReload() bool {
	return true
}

func (p *SqlPlugin) OnHotReload(newConfig map[string]interface{}) error {
	log.Println("SQL Plugin: Hot reloading configuration...")

	// Validate new configuration
	if err := p.ValidateConfig(newConfig); err != nil {
		return fmt.Errorf("hot reload validation failed: %w", err)
	}

	// Here you can implement logic for updating configuration
	// without full connection pool reload
	log.Println("SQL Plugin: Configuration hot reloaded successfully")
	return nil
}

// Main plugin methods
func (p *SqlPlugin) Initialize(container *container.Container, config map[string]interface{}) error {
	// Parse configuration
	p.config = SqlConfig{
		Host:     getStringConfig(config, "host", "localhost"),
		Port:     getIntConfig(config, "port", 5432),
		User:     getStringConfig(config, "user", ""),
		Password: getStringConfig(config, "password", ""),
		Database: getStringConfig(config, "db", ""),
		MaxConns: getIntConfig(config, "max_conns", 25),
		MinConns: getIntConfig(config, "min_conns", 5),
	}

	// Create connection string
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?pool_max_conns=%d&pool_min_conns=%d",
		p.config.User,
		p.config.Password,
		p.config.Host,
		p.config.Port,
		p.config.Database,
		p.config.MaxConns,
		p.config.MinConns,
	)

	// Create connection pool
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	p.pool = pool

	// Call base initialization
	return p.BasePlugin.Initialize(container, config)
}

func (p *SqlPlugin) Start(ctx context.Context) error {
	// Check connection
	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return p.BasePlugin.Start(ctx)
}

func (p *SqlPlugin) Stop(ctx context.Context) error {
	if p.pool != nil {
		p.pool.Close()
	}

	return p.BasePlugin.Stop(ctx)
}

// Additional methods for database operations
func (p *SqlPlugin) GetPool() *pgxpool.Pool {
	return p.pool
}

func (p *SqlPlugin) GetConfig() SqlConfig {
	return p.config
}

// Middleware for automatic transaction management
func (p *SqlPlugin) TransactionMiddleware() gorgo.MiddlewareFunc {
	return func(next gorgo.HandlerFunc) gorgo.HandlerFunc {
		return func(ctx *gorgo.Context) error {
			tx, err := p.pool.Begin(context.Background())
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			// Add transaction to context
			ctx.Set("tx", tx)

			// Execute handler
			err = next(ctx)

			if err != nil {
				// Rollback transaction on error
				if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
					log.Printf("Failed to rollback transaction: %v", rollbackErr)
				}
				return err
			}

			// Commit transaction on success
			if commitErr := tx.Commit(context.Background()); commitErr != nil {
				return fmt.Errorf("failed to commit transaction: %w", commitErr)
			}

			return nil
		}
	}
}

// MiddlewareProvider implementation
func (p *SqlPlugin) GetMiddleware() []gorgo.MiddlewareFunc {
	return []gorgo.MiddlewareFunc{
		// Can add middleware for SQL query logging
		p.sqlLoggingMiddleware(),
	}
}

func (p *SqlPlugin) sqlLoggingMiddleware() gorgo.MiddlewareFunc {
	return func(next gorgo.HandlerFunc) gorgo.HandlerFunc {
		return func(ctx *gorgo.Context) error {
			// Here you can add logic for tracking SQL queries
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
