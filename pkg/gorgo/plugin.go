package gorgo

import "github.com/GorgoFramework/gorgo/internal/container"

type Plugin interface {
	Name() string
	Initialize(container *container.Container, config map[string]interface{}) error
	Shutdown() error
}

type BasePlugin struct {
	name string
}

func NewBasePlugin(name string) BasePlugin {
	return BasePlugin{name: name}
}

func (p BasePlugin) Name() string {
	return p.name
}

func (p BasePlugin) Initialize(container *container.Container, config map[string]interface{}) error {
	return nil
}

func (p BasePlugin) Shutdown() error {
	return nil
}
