package container

import (
	"fmt"
	"reflect"
	"sync"
)

type Container struct {
	services map[string]interface{}
	mu       sync.RWMutex
}

func NewContainer() *Container {
	return &Container{
		services: make(map[string]interface{}),
	}
}

func (c *Container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

func (c *Container) Get(name string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	service, ok := c.services[name]
	return service, ok
}

func (c *Container) GetTyped(name string, target interface{}) error {
	service, exists := c.Get(name)
	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target is not a pointer")
	}

	serviceValue := reflect.ValueOf(service)
	targetValue.Elem().Set(serviceValue)
	return nil
}
