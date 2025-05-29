package gorgo

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/GorgoFramework/gorgo/internal/container"
)

// PluginPriority defines the plugin loading priority
type PluginPriority int

const (
	PriorityLowest  PluginPriority = 0
	PriorityLow     PluginPriority = 25
	PriorityNormal  PluginPriority = 50
	PriorityHigh    PluginPriority = 75
	PriorityHighest PluginPriority = 100
)

// PluginState represents the plugin state
type PluginState int

const (
	StateUninitialized PluginState = iota
	StateInitializing
	StateInitialized
	StateStarting
	StateRunning
	StateStopping
	StateStopped
	StateError
)

// Event represents an event in the system
type Event struct {
	Name string
	Data map[string]interface{}
	ctx  context.Context
}

// EventHandler event handler
type EventHandler func(event *Event) error

// PluginMetadata contains plugin metadata
type PluginMetadata struct {
	Name         string
	Version      string
	Description  string
	Author       string
	Dependencies []string
	Priority     PluginPriority
	Tags         []string
}

// LifecycleHooks defines lifecycle hooks
type LifecycleHooks interface {
	OnBeforeInit(ctx context.Context) error
	OnAfterInit(ctx context.Context) error
	OnBeforeStart(ctx context.Context) error
	OnAfterStart(ctx context.Context) error
	OnBeforeStop(ctx context.Context) error
	OnAfterStop(ctx context.Context) error
}

// MiddlewareProvider allows a plugin to provide middleware
type MiddlewareProvider interface {
	GetMiddleware() []MiddlewareFunc
}

// EventSubscriber allows a plugin to subscribe to events
type EventSubscriber interface {
	GetEventSubscriptions() map[string]EventHandler
}

// ConfigurablePlugin allows a plugin to validate and handle configuration
type ConfigurablePlugin interface {
	ValidateConfig(config map[string]interface{}) error
	GetDefaultConfig() map[string]interface{}
}

// HotReloadable allows a plugin to support hot reload
type HotReloadable interface {
	CanHotReload() bool
	OnHotReload(newConfig map[string]interface{}) error
}

// ServiceProvider allows a plugin to register services
type ServiceProvider interface {
	GetServices() map[string]interface{}
}

// Plugin extended plugin interface
type Plugin interface {
	GetMetadata() PluginMetadata
	Initialize(container *container.Container, config map[string]interface{}) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetState() PluginState
}

// BasePlugin base plugin implementation
type BasePlugin struct {
	metadata PluginMetadata
	state    PluginState
	mu       sync.RWMutex
}

func NewBasePlugin(metadata PluginMetadata) BasePlugin {
	return BasePlugin{
		metadata: metadata,
		state:    StateUninitialized,
	}
}

func (p *BasePlugin) GetMetadata() PluginMetadata {
	return p.metadata
}

func (p *BasePlugin) Initialize(container *container.Container, config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = StateInitialized
	return nil
}

func (p *BasePlugin) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = StateRunning
	return nil
}

func (p *BasePlugin) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = StateStopped
	return nil
}

func (p *BasePlugin) GetState() PluginState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// Default hook implementations
func (p *BasePlugin) OnBeforeInit(ctx context.Context) error  { return nil }
func (p *BasePlugin) OnAfterInit(ctx context.Context) error   { return nil }
func (p *BasePlugin) OnBeforeStart(ctx context.Context) error { return nil }
func (p *BasePlugin) OnAfterStart(ctx context.Context) error  { return nil }
func (p *BasePlugin) OnBeforeStop(ctx context.Context) error  { return nil }
func (p *BasePlugin) OnAfterStop(ctx context.Context) error   { return nil }

// EventBus event system
type EventBus struct {
	subscribers map[string][]EventHandler
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]EventHandler),
	}
}

func (eb *EventBus) Subscribe(eventName string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[eventName] = append(eb.subscribers[eventName], handler)
}

func (eb *EventBus) Publish(ctx context.Context, eventName string, data map[string]interface{}) error {
	eb.mu.RLock()
	handlers := eb.subscribers[eventName]
	eb.mu.RUnlock()

	event := &Event{
		Name: eventName,
		Data: data,
		ctx:  ctx,
	}

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			return fmt.Errorf("event handler error for %s: %w", eventName, err)
		}
	}

	return nil
}

// PluginManager manages plugins
type PluginManager struct {
	plugins   map[string]Plugin
	eventBus  *EventBus
	container *container.Container
	mu        sync.RWMutex
}

func NewPluginManager(container *container.Container) *PluginManager {
	return &PluginManager{
		plugins:   make(map[string]Plugin),
		eventBus:  NewEventBus(),
		container: container,
	}
}

func (pm *PluginManager) RegisterPlugin(plugin Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	metadata := plugin.GetMetadata()

	// Check dependencies
	for _, dep := range metadata.Dependencies {
		if _, exists := pm.plugins[dep]; !exists {
			return fmt.Errorf("dependency %s not found for plugin %s", dep, metadata.Name)
		}
	}

	pm.plugins[metadata.Name] = plugin
	return nil
}

func (pm *PluginManager) InitializePlugins(configs map[string]map[string]interface{}) error {
	// Sort plugins by priority and dependencies
	sortedPlugins := pm.getSortedPlugins()

	for _, plugin := range sortedPlugins {
		metadata := plugin.GetMetadata()
		config := configs[metadata.Name]
		if config == nil {
			config = make(map[string]interface{})
		}

		// Configuration validation
		if configurable, ok := plugin.(ConfigurablePlugin); ok {
			if err := configurable.ValidateConfig(config); err != nil {
				return fmt.Errorf("config validation failed for plugin %s: %w", metadata.Name, err)
			}
		}

		// Lifecycle hooks
		if hooks, ok := plugin.(LifecycleHooks); ok {
			if err := hooks.OnBeforeInit(context.Background()); err != nil {
				return fmt.Errorf("OnBeforeInit failed for plugin %s: %w", metadata.Name, err)
			}
		}

		// Initialization
		if err := plugin.Initialize(pm.container, config); err != nil {
			return fmt.Errorf("initialization failed for plugin %s: %w", metadata.Name, err)
		}

		// Service registration
		if serviceProvider, ok := plugin.(ServiceProvider); ok {
			services := serviceProvider.GetServices()
			for name, service := range services {
				pm.container.Register(name, service)
			}
		}

		// Event subscription
		if subscriber, ok := plugin.(EventSubscriber); ok {
			subscriptions := subscriber.GetEventSubscriptions()
			for eventName, handler := range subscriptions {
				pm.eventBus.Subscribe(eventName, handler)
			}
		}

		// Post-initialization hooks
		if hooks, ok := plugin.(LifecycleHooks); ok {
			if err := hooks.OnAfterInit(context.Background()); err != nil {
				return fmt.Errorf("OnAfterInit failed for plugin %s: %w", metadata.Name, err)
			}
		}
	}

	return nil
}

func (pm *PluginManager) StartPlugins(ctx context.Context) error {
	sortedPlugins := pm.getSortedPlugins()

	for _, plugin := range sortedPlugins {
		metadata := plugin.GetMetadata()

		// Pre-start hooks
		if hooks, ok := plugin.(LifecycleHooks); ok {
			if err := hooks.OnBeforeStart(ctx); err != nil {
				return fmt.Errorf("OnBeforeStart failed for plugin %s: %w", metadata.Name, err)
			}
		}

		// Start
		if err := plugin.Start(ctx); err != nil {
			return fmt.Errorf("start failed for plugin %s: %w", metadata.Name, err)
		}

		// Post-start hooks
		if hooks, ok := plugin.(LifecycleHooks); ok {
			if err := hooks.OnAfterStart(ctx); err != nil {
				return fmt.Errorf("OnAfterStart failed for plugin %s: %w", metadata.Name, err)
			}
		}

		// Publish plugin started event
		pm.eventBus.Publish(ctx, "plugin.started", map[string]interface{}{
			"plugin": metadata.Name,
		})
	}

	return nil
}

func (pm *PluginManager) StopPlugins(ctx context.Context) error {
	// Stop in reverse order
	sortedPlugins := pm.getSortedPlugins()
	for i := len(sortedPlugins) - 1; i >= 0; i-- {
		plugin := sortedPlugins[i]
		metadata := plugin.GetMetadata()

		// Pre-stop hooks
		if hooks, ok := plugin.(LifecycleHooks); ok {
			if err := hooks.OnBeforeStop(ctx); err != nil {
				return fmt.Errorf("OnBeforeStop failed for plugin %s: %w", metadata.Name, err)
			}
		}

		// Stop
		if err := plugin.Stop(ctx); err != nil {
			return fmt.Errorf("stop failed for plugin %s: %w", metadata.Name, err)
		}

		// Post-stop hooks
		if hooks, ok := plugin.(LifecycleHooks); ok {
			if err := hooks.OnAfterStop(ctx); err != nil {
				return fmt.Errorf("OnAfterStop failed for plugin %s: %w", metadata.Name, err)
			}
		}

		// Publish plugin stopped event
		pm.eventBus.Publish(ctx, "plugin.stopped", map[string]interface{}{
			"plugin": metadata.Name,
		})
	}

	return nil
}

func (pm *PluginManager) GetPlugin(name string) (Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	plugin, exists := pm.plugins[name]
	return plugin, exists
}

func (pm *PluginManager) GetEventBus() *EventBus {
	return pm.eventBus
}

func (pm *PluginManager) GetMiddleware() []MiddlewareFunc {
	var middleware []MiddlewareFunc

	sortedPlugins := pm.getSortedPlugins()
	for _, plugin := range sortedPlugins {
		if provider, ok := plugin.(MiddlewareProvider); ok {
			middleware = append(middleware, provider.GetMiddleware()...)
		}
	}

	return middleware
}

func (pm *PluginManager) HotReloadPlugin(name string, newConfig map[string]interface{}) error {
	pm.mu.RLock()
	plugin, exists := pm.plugins[name]
	pm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if reloadable, ok := plugin.(HotReloadable); ok && reloadable.CanHotReload() {
		return reloadable.OnHotReload(newConfig)
	}

	return fmt.Errorf("plugin %s does not support hot reload", name)
}

// getSortedPlugins returns plugins sorted by priority and dependencies
func (pm *PluginManager) getSortedPlugins() []Plugin {
	var plugins []Plugin
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}

	// Topological sort by dependencies + priority
	sort.Slice(plugins, func(i, j int) bool {
		metaI := plugins[i].GetMetadata()
		metaJ := plugins[j].GetMetadata()

		// First by priority
		if metaI.Priority != metaJ.Priority {
			return metaI.Priority > metaJ.Priority
		}

		// Then by dependencies
		for _, dep := range metaJ.Dependencies {
			if dep == metaI.Name {
				return true // i should be before j
			}
		}

		return false
	})

	return plugins
}
