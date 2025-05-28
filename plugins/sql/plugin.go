package sql

import (
	"context"
	"fmt"

	"github.com/GorgoFramework/gorgo/internal/container"
	"github.com/GorgoFramework/gorgo/pkg/gorgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SqlPlugin struct {
	gorgo.BasePlugin
	pool *pgxpool.Pool
}

func NewSqlPlugin() *SqlPlugin {
	return &SqlPlugin{
		BasePlugin: gorgo.NewBasePlugin("sql"),
	}
}

func (p *SqlPlugin) Initialize(container *container.Container, config map[string]interface{}) error {
	host, _ := config["host"].(string)
	if host == "" {
		return fmt.Errorf("host is required")
	}

	port, _ := config["port"].(int)
	if port == 0 {
		port = 5432
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

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, dbName)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("failed to create pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	p.pool = pool
	container.Register("db", pool)

	return nil
}

func (p *SqlPlugin) Shutdown() error {
	if p.pool != nil {
		p.pool.Close()
	}

	return nil
}

func (p *SqlPlugin) GetPool() *pgxpool.Pool {
	return p.pool
}
