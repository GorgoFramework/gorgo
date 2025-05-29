package gorgo

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/GorgoFramework/gorgo/internal/container"
)

// Mock plugin implementations for testing

// MockPlugin - basic mock plugin
type MockPlugin struct {
	BasePlugin
	initError  error
	startError error
	stopError  error
}

func NewMockPlugin(name string, priority PluginPriority) *MockPlugin {
	metadata := PluginMetadata{
		Name:     name,
		Version:  "1.0.0",
		Priority: priority,
	}
	return &MockPlugin{
		BasePlugin: NewBasePlugin(metadata),
	}
}

func (mp *MockPlugin) Initialize(container *container.Container, config map[string]interface{}) error {
	if mp.initError != nil {
		return mp.initError
	}
	return mp.BasePlugin.Initialize(container, config)
}

func (mp *MockPlugin) Start(ctx context.Context) error {
	if mp.startError != nil {
		return mp.startError
	}
	return mp.BasePlugin.Start(ctx)
}

func (mp *MockPlugin) Stop(ctx context.Context) error {
	if mp.stopError != nil {
		return mp.stopError
	}
	return mp.BasePlugin.Stop(ctx)
}

// MockConfigurablePlugin - mock plugin with configuration
type MockConfigurablePlugin struct {
	*MockPlugin
	validateConfigError error
	defaultConfig       map[string]interface{}
}

func NewMockConfigurablePlugin(name string) *MockConfigurablePlugin {
	return &MockConfigurablePlugin{
		MockPlugin: NewMockPlugin(name, PriorityNormal),
		defaultConfig: map[string]interface{}{
			"enabled": true,
			"timeout": 30,
		},
	}
}

func (mcp *MockConfigurablePlugin) ValidateConfig(config map[string]interface{}) error {
	if mcp.validateConfigError != nil {
		return mcp.validateConfigError
	}
	if timeout, ok := config["timeout"]; ok {
		if timeoutInt, ok := timeout.(int); ok && timeoutInt < 0 {
			return errors.New("timeout cannot be negative")
		}
	}
	return nil
}

func (mcp *MockConfigurablePlugin) GetDefaultConfig() map[string]interface{} {
	return mcp.defaultConfig
}

// MockServiceProvider - mock plugin that provides services
type MockServiceProvider struct {
	*MockPlugin
	services map[string]interface{}
}

func NewMockServiceProvider(name string, services map[string]interface{}) *MockServiceProvider {
	return &MockServiceProvider{
		MockPlugin: NewMockPlugin(name, PriorityNormal),
		services:   services,
	}
}

func (msp *MockServiceProvider) GetServices() map[string]interface{} {
	return msp.services
}

// MockEventSubscriber - mock plugin that subscribes to events
type MockEventSubscriber struct {
	*MockPlugin
	subscriptions map[string]EventHandler
}

func NewMockEventSubscriber(name string) *MockEventSubscriber {
	return &MockEventSubscriber{
		MockPlugin: NewMockPlugin(name, PriorityNormal),
		subscriptions: map[string]EventHandler{
			"test.event": func(event *Event) error {
				return nil
			},
		},
	}
}

func (mes *MockEventSubscriber) GetEventSubscriptions() map[string]EventHandler {
	return mes.subscriptions
}

// MockLifecycleHooks - mock plugin with lifecycle hooks
type MockLifecycleHooks struct {
	*MockPlugin
	beforeInitCalled  bool
	afterInitCalled   bool
	beforeStartCalled bool
	afterStartCalled  bool
	beforeStopCalled  bool
	afterStopCalled   bool
	hookError         error
}

func NewMockLifecycleHooks(name string) *MockLifecycleHooks {
	return &MockLifecycleHooks{
		MockPlugin: NewMockPlugin(name, PriorityNormal),
	}
}

func (mlh *MockLifecycleHooks) OnBeforeInit(ctx context.Context) error {
	if mlh.hookError != nil {
		return mlh.hookError
	}
	mlh.beforeInitCalled = true
	return nil
}

func (mlh *MockLifecycleHooks) OnAfterInit(ctx context.Context) error {
	if mlh.hookError != nil {
		return mlh.hookError
	}
	mlh.afterInitCalled = true
	return nil
}

func (mlh *MockLifecycleHooks) OnBeforeStart(ctx context.Context) error {
	if mlh.hookError != nil {
		return mlh.hookError
	}
	mlh.beforeStartCalled = true
	return nil
}

func (mlh *MockLifecycleHooks) OnAfterStart(ctx context.Context) error {
	if mlh.hookError != nil {
		return mlh.hookError
	}
	mlh.afterStartCalled = true
	return nil
}

func (mlh *MockLifecycleHooks) OnBeforeStop(ctx context.Context) error {
	if mlh.hookError != nil {
		return mlh.hookError
	}
	mlh.beforeStopCalled = true
	return nil
}

func (mlh *MockLifecycleHooks) OnAfterStop(ctx context.Context) error {
	if mlh.hookError != nil {
		return mlh.hookError
	}
	mlh.afterStopCalled = true
	return nil
}

// MockMiddlewareProvider - mock plugin that provides middleware
type MockMiddlewareProvider struct {
	*MockPlugin
	middleware []MiddlewareFunc
}

func NewMockMiddlewareProvider(name string) *MockMiddlewareProvider {
	middleware := []MiddlewareFunc{
		func(next HandlerFunc) HandlerFunc {
			return func(ctx *Context) error {
				ctx.Set("middleware", "applied")
				return next(ctx)
			}
		},
	}
	return &MockMiddlewareProvider{
		MockPlugin: NewMockPlugin(name, PriorityNormal),
		middleware: middleware,
	}
}

func (mmp *MockMiddlewareProvider) GetMiddleware() []MiddlewareFunc {
	return mmp.middleware
}

// MockHotReloadable - mock plugin that supports hot reload
type MockHotReloadable struct {
	*MockPlugin
	canReload    bool
	reloadError  error
	reloadedWith map[string]interface{}
	reloadCalled bool
}

func NewMockHotReloadable(name string, canReload bool) *MockHotReloadable {
	return &MockHotReloadable{
		MockPlugin: NewMockPlugin(name, PriorityNormal),
		canReload:  canReload,
	}
}

func (mhr *MockHotReloadable) CanHotReload() bool {
	return mhr.canReload
}

func (mhr *MockHotReloadable) OnHotReload(newConfig map[string]interface{}) error {
	if mhr.reloadError != nil {
		return mhr.reloadError
	}
	mhr.reloadCalled = true
	mhr.reloadedWith = newConfig
	return nil
}

// Test Event struct
func TestEvent(t *testing.T) {
	ctx := context.Background()
	event := &Event{
		Name: "test.event",
		Data: map[string]interface{}{
			"key": "value",
		},
		ctx: ctx,
	}

	if event.Name != "test.event" {
		t.Errorf("expected event name 'test.event', got '%s'", event.Name)
	}

	if event.Data["key"] != "value" {
		t.Errorf("expected data value 'value', got '%v'", event.Data["key"])
	}

	if event.ctx != ctx {
		t.Error("event context is not the same as provided")
	}
}

// Test PluginMetadata
func TestPluginMetadata(t *testing.T) {
	metadata := PluginMetadata{
		Name:         "test-plugin",
		Version:      "1.0.0",
		Description:  "Test plugin",
		Author:       "Test Author",
		Dependencies: []string{"dep1", "dep2"},
		Priority:     PriorityHigh,
		Tags:         []string{"tag1", "tag2"},
	}

	if metadata.Name != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got '%s'", metadata.Name)
	}

	if metadata.Priority != PriorityHigh {
		t.Errorf("expected priority %d, got %d", PriorityHigh, metadata.Priority)
	}

	if len(metadata.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(metadata.Dependencies))
	}
}

// Test BasePlugin
func TestBasePlugin(t *testing.T) {
	metadata := PluginMetadata{
		Name:    "base-test",
		Version: "1.0.0",
	}

	plugin := NewBasePlugin(metadata)

	// Test initial state
	if plugin.GetState() != StateUninitialized {
		t.Errorf("expected initial state %d, got %d", StateUninitialized, plugin.GetState())
	}

	// Test metadata
	meta := plugin.GetMetadata()
	if meta.Name != "base-test" {
		t.Errorf("expected name 'base-test', got '%s'", meta.Name)
	}

	// Test Initialize
	c := container.NewContainer()
	err := plugin.Initialize(c, map[string]interface{}{})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}
	if plugin.GetState() != StateInitialized {
		t.Errorf("expected state %d after Initialize, got %d", StateInitialized, plugin.GetState())
	}

	// Test Start
	ctx := context.Background()
	err = plugin.Start(ctx)
	if err != nil {
		t.Errorf("Start failed: %v", err)
	}
	if plugin.GetState() != StateRunning {
		t.Errorf("expected state %d after Start, got %d", StateRunning, plugin.GetState())
	}

	// Test Stop
	err = plugin.Stop(ctx)
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
	if plugin.GetState() != StateStopped {
		t.Errorf("expected state %d after Stop, got %d", StateStopped, plugin.GetState())
	}

	// Test lifecycle hooks (should not fail)
	if err := plugin.OnBeforeInit(ctx); err != nil {
		t.Errorf("OnBeforeInit failed: %v", err)
	}
	if err := plugin.OnAfterInit(ctx); err != nil {
		t.Errorf("OnAfterInit failed: %v", err)
	}
	if err := plugin.OnBeforeStart(ctx); err != nil {
		t.Errorf("OnBeforeStart failed: %v", err)
	}
	if err := plugin.OnAfterStart(ctx); err != nil {
		t.Errorf("OnAfterStart failed: %v", err)
	}
	if err := plugin.OnBeforeStop(ctx); err != nil {
		t.Errorf("OnBeforeStop failed: %v", err)
	}
	if err := plugin.OnAfterStop(ctx); err != nil {
		t.Errorf("OnAfterStop failed: %v", err)
	}
}

// Test EventBus
func TestEventBus(t *testing.T) {
	eventBus := NewEventBus()

	var receivedEvent *Event
	handler := func(event *Event) error {
		receivedEvent = event
		return nil
	}

	// Test Subscribe
	eventBus.Subscribe("test.event", handler)

	// Test Publish
	ctx := context.Background()
	data := map[string]interface{}{"key": "value"}
	err := eventBus.Publish(ctx, "test.event", data)
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	// Check if event was received
	if receivedEvent == nil {
		t.Fatal("event was not received")
	}
	if receivedEvent.Name != "test.event" {
		t.Errorf("expected event name 'test.event', got '%s'", receivedEvent.Name)
	}
	if receivedEvent.Data["key"] != "value" {
		t.Errorf("expected data value 'value', got '%v'", receivedEvent.Data["key"])
	}
}

func TestEventBus_HandlerError(t *testing.T) {
	eventBus := NewEventBus()

	expectedError := errors.New("handler error")
	handler := func(event *Event) error {
		return expectedError
	}

	eventBus.Subscribe("error.event", handler)

	ctx := context.Background()
	err := eventBus.Publish(ctx, "error.event", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error from handler")
	}

	if !errors.Is(err, expectedError) {
		t.Errorf("expected wrapped handler error, got: %v", err)
	}
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	eventBus := NewEventBus()

	var callCount int
	handler1 := func(event *Event) error {
		callCount++
		return nil
	}
	handler2 := func(event *Event) error {
		callCount++
		return nil
	}

	eventBus.Subscribe("multi.event", handler1)
	eventBus.Subscribe("multi.event", handler2)

	ctx := context.Background()
	err := eventBus.Publish(ctx, "multi.event", map[string]interface{}{})
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 handler calls, got %d", callCount)
	}
}

// Test PluginManager
func TestNewPluginManager(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	if pm == nil {
		t.Fatal("NewPluginManager returned nil")
	}

	if pm.plugins == nil {
		t.Fatal("plugins map is nil")
	}

	if pm.eventBus == nil {
		t.Fatal("eventBus is nil")
	}

	if pm.container != c {
		t.Fatal("container is not the same as provided")
	}
}

func TestPluginManager_RegisterPlugin(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockPlugin("test-plugin", PriorityNormal)

	// Test successful registration
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Errorf("RegisterPlugin failed: %v", err)
	}

	// Check if plugin was registered
	retrievedPlugin, exists := pm.GetPlugin("test-plugin")
	if !exists {
		t.Fatal("plugin was not registered")
	}
	if retrievedPlugin != plugin {
		t.Fatal("retrieved plugin is not the same as registered")
	}
}

func TestPluginManager_RegisterPlugin_WithDependencies(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	// Register dependency first
	dep := NewMockPlugin("dependency", PriorityNormal)
	err := pm.RegisterPlugin(dep)
	if err != nil {
		t.Errorf("RegisterPlugin failed for dependency: %v", err)
	}

	// Create plugin with dependency
	metadata := PluginMetadata{
		Name:         "dependent-plugin",
		Dependencies: []string{"dependency"},
		Priority:     PriorityNormal,
	}
	dependent := &MockPlugin{BasePlugin: NewBasePlugin(metadata)}

	// Test successful registration with satisfied dependency
	err = pm.RegisterPlugin(dependent)
	if err != nil {
		t.Errorf("RegisterPlugin failed for dependent plugin: %v", err)
	}
}

func TestPluginManager_RegisterPlugin_MissingDependency(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	// Create plugin with missing dependency
	metadata := PluginMetadata{
		Name:         "dependent-plugin",
		Dependencies: []string{"missing-dependency"},
		Priority:     PriorityNormal,
	}
	dependent := &MockPlugin{BasePlugin: NewBasePlugin(metadata)}

	// Test registration failure with missing dependency
	err := pm.RegisterPlugin(dependent)
	if err == nil {
		t.Fatal("expected error for missing dependency")
	}

	expectedError := "dependency missing-dependency not found for plugin dependent-plugin"
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestPluginManager_InitializePlugins(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockPlugin("test-plugin", PriorityNormal)
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	configs := map[string]map[string]interface{}{
		"test-plugin": {
			"enabled": true,
		},
	}

	err = pm.InitializePlugins(configs)
	if err != nil {
		t.Errorf("InitializePlugins failed: %v", err)
	}

	if plugin.GetState() != StateInitialized {
		t.Errorf("expected plugin state %d, got %d", StateInitialized, plugin.GetState())
	}
}

func TestPluginManager_InitializePlugins_ConfigurablePlugin(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockConfigurablePlugin("configurable-plugin")
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	configs := map[string]map[string]interface{}{
		"configurable-plugin": {
			"timeout": 60,
		},
	}

	err = pm.InitializePlugins(configs)
	if err != nil {
		t.Errorf("InitializePlugins failed: %v", err)
	}
}

func TestPluginManager_InitializePlugins_ConfigValidationError(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockConfigurablePlugin("configurable-plugin")
	plugin.validateConfigError = errors.New("invalid config")
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	configs := map[string]map[string]interface{}{
		"configurable-plugin": {
			"timeout": -1,
		},
	}

	err = pm.InitializePlugins(configs)
	if err == nil {
		t.Fatal("expected config validation error")
	}
}

func TestPluginManager_InitializePlugins_LifecycleHooks(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockLifecycleHooks("lifecycle-plugin")
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	err = pm.InitializePlugins(map[string]map[string]interface{}{})
	if err != nil {
		t.Errorf("InitializePlugins failed: %v", err)
	}

	if !plugin.beforeInitCalled {
		t.Error("OnBeforeInit was not called")
	}
	if !plugin.afterInitCalled {
		t.Error("OnAfterInit was not called")
	}
}

func TestPluginManager_InitializePlugins_ServiceProvider(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	services := map[string]interface{}{
		"test-service": "test-value",
	}
	plugin := NewMockServiceProvider("service-provider", services)
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	err = pm.InitializePlugins(map[string]map[string]interface{}{})
	if err != nil {
		t.Errorf("InitializePlugins failed: %v", err)
	}

	// Check if service was registered
	service, exists := c.Get("test-service")
	if !exists {
		t.Fatal("service was not registered")
	}
	if service != "test-value" {
		t.Errorf("expected service value 'test-value', got '%v'", service)
	}
}

func TestPluginManager_InitializePlugins_EventSubscriber(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockEventSubscriber("event-subscriber")
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	err = pm.InitializePlugins(map[string]map[string]interface{}{})
	if err != nil {
		t.Errorf("InitializePlugins failed: %v", err)
	}

	// Test if event subscription works
	var eventReceived bool
	plugin.subscriptions["test.event"] = func(event *Event) error {
		eventReceived = true
		return nil
	}

	// Manually subscribe since we can't access internal eventBus state easily
	pm.eventBus.Subscribe("test.event", plugin.subscriptions["test.event"])

	ctx := context.Background()
	err = pm.eventBus.Publish(ctx, "test.event", map[string]interface{}{})
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	if !eventReceived {
		t.Error("event was not received by subscriber")
	}
}

func TestPluginManager_StartPlugins(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockPlugin("test-plugin", PriorityNormal)
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	// Initialize first
	err = pm.InitializePlugins(map[string]map[string]interface{}{})
	if err != nil {
		t.Fatalf("InitializePlugins failed: %v", err)
	}

	ctx := context.Background()
	err = pm.StartPlugins(ctx)
	if err != nil {
		t.Errorf("StartPlugins failed: %v", err)
	}

	if plugin.GetState() != StateRunning {
		t.Errorf("expected plugin state %d, got %d", StateRunning, plugin.GetState())
	}
}

func TestPluginManager_StartPlugins_WithLifecycleHooks(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockLifecycleHooks("lifecycle-plugin")
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	// Initialize first
	err = pm.InitializePlugins(map[string]map[string]interface{}{})
	if err != nil {
		t.Fatalf("InitializePlugins failed: %v", err)
	}

	ctx := context.Background()
	err = pm.StartPlugins(ctx)
	if err != nil {
		t.Errorf("StartPlugins failed: %v", err)
	}

	if !plugin.beforeStartCalled {
		t.Error("OnBeforeStart was not called")
	}
	if !plugin.afterStartCalled {
		t.Error("OnAfterStart was not called")
	}
}

func TestPluginManager_StopPlugins(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockPlugin("test-plugin", PriorityNormal)
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	// Initialize and start first
	err = pm.InitializePlugins(map[string]map[string]interface{}{})
	if err != nil {
		t.Fatalf("InitializePlugins failed: %v", err)
	}

	ctx := context.Background()
	err = pm.StartPlugins(ctx)
	if err != nil {
		t.Fatalf("StartPlugins failed: %v", err)
	}

	err = pm.StopPlugins(ctx)
	if err != nil {
		t.Errorf("StopPlugins failed: %v", err)
	}

	if plugin.GetState() != StateStopped {
		t.Errorf("expected plugin state %d, got %d", StateStopped, plugin.GetState())
	}
}

func TestPluginManager_StopPlugins_WithLifecycleHooks(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockLifecycleHooks("lifecycle-plugin")
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	// Initialize and start first
	err = pm.InitializePlugins(map[string]map[string]interface{}{})
	if err != nil {
		t.Fatalf("InitializePlugins failed: %v", err)
	}

	ctx := context.Background()
	err = pm.StartPlugins(ctx)
	if err != nil {
		t.Fatalf("StartPlugins failed: %v", err)
	}

	err = pm.StopPlugins(ctx)
	if err != nil {
		t.Errorf("StopPlugins failed: %v", err)
	}

	if !plugin.beforeStopCalled {
		t.Error("OnBeforeStop was not called")
	}
	if !plugin.afterStopCalled {
		t.Error("OnAfterStop was not called")
	}
}

func TestPluginManager_GetEventBus(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	eventBus := pm.GetEventBus()
	if eventBus == nil {
		t.Fatal("GetEventBus returned nil")
	}
	if eventBus != pm.eventBus {
		t.Fatal("GetEventBus returned different event bus")
	}
}

func TestPluginManager_GetMiddleware(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockMiddlewareProvider("middleware-plugin")
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	middleware := pm.GetMiddleware()
	if len(middleware) != 1 {
		t.Errorf("expected 1 middleware function, got %d", len(middleware))
	}
}

func TestPluginManager_HotReloadPlugin(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockHotReloadable("reloadable-plugin", true)
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	newConfig := map[string]interface{}{
		"new-setting": "new-value",
	}

	err = pm.HotReloadPlugin("reloadable-plugin", newConfig)
	if err != nil {
		t.Errorf("HotReloadPlugin failed: %v", err)
	}

	if !plugin.reloadCalled {
		t.Error("OnHotReload was not called")
	}

	if plugin.reloadedWith["new-setting"] != "new-value" {
		t.Error("plugin was not reloaded with correct config")
	}
}

func TestPluginManager_HotReloadPlugin_NotSupported(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockHotReloadable("non-reloadable-plugin", false)
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	err = pm.HotReloadPlugin("non-reloadable-plugin", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for non-reloadable plugin")
	}

	expectedError := "plugin non-reloadable-plugin does not support hot reload"
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestPluginManager_HotReloadPlugin_NotFound(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	err := pm.HotReloadPlugin("non-existent-plugin", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for non-existent plugin")
	}

	expectedError := "plugin non-existent-plugin not found"
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestPluginManager_GetSortedPlugins_ByPriority(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	// Register plugins with different priorities
	highPriorityPlugin := NewMockPlugin("high-priority", PriorityHigh)
	lowPriorityPlugin := NewMockPlugin("low-priority", PriorityLow)
	normalPriorityPlugin := NewMockPlugin("normal-priority", PriorityNormal)

	err := pm.RegisterPlugin(lowPriorityPlugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}
	err = pm.RegisterPlugin(highPriorityPlugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}
	err = pm.RegisterPlugin(normalPriorityPlugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	sortedPlugins := pm.getSortedPlugins()

	// Check order: high -> normal -> low
	if sortedPlugins[0].GetMetadata().Name != "high-priority" {
		t.Errorf("expected first plugin to be 'high-priority', got '%s'", sortedPlugins[0].GetMetadata().Name)
	}
	if sortedPlugins[1].GetMetadata().Name != "normal-priority" {
		t.Errorf("expected second plugin to be 'normal-priority', got '%s'", sortedPlugins[1].GetMetadata().Name)
	}
	if sortedPlugins[2].GetMetadata().Name != "low-priority" {
		t.Errorf("expected third plugin to be 'low-priority', got '%s'", sortedPlugins[2].GetMetadata().Name)
	}
}

func TestPluginManager_GetSortedPlugins_ByDependencies(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	// Create dependency
	dependency := NewMockPlugin("dependency", PriorityNormal)
	err := pm.RegisterPlugin(dependency)
	if err != nil {
		t.Fatalf("RegisterPlugin failed for dependency: %v", err)
	}

	// Create dependent plugin with same priority
	dependentMetadata := PluginMetadata{
		Name:         "dependent",
		Dependencies: []string{"dependency"},
		Priority:     PriorityNormal,
	}
	dependent := &MockPlugin{BasePlugin: NewBasePlugin(dependentMetadata)}
	err = pm.RegisterPlugin(dependent)
	if err != nil {
		t.Fatalf("RegisterPlugin failed for dependent: %v", err)
	}

	sortedPlugins := pm.getSortedPlugins()

	// Check that dependency comes before dependent
	var dependencyIndex, dependentIndex int
	for i, plugin := range sortedPlugins {
		name := plugin.GetMetadata().Name
		if name == "dependency" {
			dependencyIndex = i
		} else if name == "dependent" {
			dependentIndex = i
		}
	}

	if dependencyIndex >= dependentIndex {
		t.Error("dependency should come before dependent plugin")
	}
}

// Test error scenarios

func TestPluginManager_InitializePlugins_InitError(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockPlugin("error-plugin", PriorityNormal)
	plugin.initError = errors.New("init failed")
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	err = pm.InitializePlugins(map[string]map[string]interface{}{})
	if err == nil {
		t.Fatal("expected initialization error")
	}
}

func TestPluginManager_StartPlugins_StartError(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockPlugin("error-plugin", PriorityNormal)
	plugin.startError = errors.New("start failed")
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	err = pm.InitializePlugins(map[string]map[string]interface{}{})
	if err != nil {
		t.Fatalf("InitializePlugins failed: %v", err)
	}

	ctx := context.Background()
	err = pm.StartPlugins(ctx)
	if err == nil {
		t.Fatal("expected start error")
	}
}

func TestPluginManager_StopPlugins_StopError(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockPlugin("error-plugin", PriorityNormal)
	plugin.stopError = errors.New("stop failed")
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	err = pm.InitializePlugins(map[string]map[string]interface{}{})
	if err != nil {
		t.Fatalf("InitializePlugins failed: %v", err)
	}

	ctx := context.Background()
	err = pm.StartPlugins(ctx)
	if err != nil {
		t.Fatalf("StartPlugins failed: %v", err)
	}

	err = pm.StopPlugins(ctx)
	if err == nil {
		t.Fatal("expected stop error")
	}
}

func TestPluginManager_LifecycleHooks_Error(t *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockLifecycleHooks("hook-error-plugin")
	plugin.hookError = errors.New("hook failed")
	err := pm.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	// Test init hook error
	err = pm.InitializePlugins(map[string]map[string]interface{}{})
	if err == nil {
		t.Fatal("expected hook error during initialization")
	}
}

// Benchmark tests

func BenchmarkPluginManager_RegisterPlugin(b *testing.B) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plugin := NewMockPlugin(fmt.Sprintf("plugin-%d", i), PriorityNormal)
		pm.RegisterPlugin(plugin)
	}
}

func BenchmarkPluginManager_GetPlugin(b *testing.B) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	plugin := NewMockPlugin("benchmark-plugin", PriorityNormal)
	pm.RegisterPlugin(plugin)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.GetPlugin("benchmark-plugin")
	}
}

func BenchmarkEventBus_Publish(b *testing.B) {
	eventBus := NewEventBus()

	handler := func(event *Event) error {
		return nil
	}
	eventBus.Subscribe("bench.event", handler)

	ctx := context.Background()
	data := map[string]interface{}{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eventBus.Publish(ctx, "bench.event", data)
	}
}

// Test concurrent access

func TestPluginManager_ConcurrentAccess(b *testing.T) {
	c := container.NewContainer()
	pm := NewPluginManager(c)

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	// Concurrent plugin registration
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			plugin := NewMockPlugin(fmt.Sprintf("concurrent-plugin-%d", id), PriorityNormal)
			pm.RegisterPlugin(plugin)
			done <- true
		}(i)
	}

	// Wait for all registrations
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Concurrent plugin retrieval
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			pm.GetPlugin(fmt.Sprintf("concurrent-plugin-%d", id))
			done <- true
		}(i)
	}

	// Wait for all retrievals
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
