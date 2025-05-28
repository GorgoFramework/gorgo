package gorgo

import "github.com/GorgoFramework/gorgo/internal/container"

type Plugin interface {
	Name() string
	Initialize(container *container.Container, config map[string]interface{}) error
	Shutdown() error
}
