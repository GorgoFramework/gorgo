package monitoring

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/GorgoFramework/gorgo/internal/container"
	"github.com/GorgoFramework/gorgo/pkg/gorgo"
)

type MonitoringPlugin struct {
	gorgo.BasePlugin
	stats    *Stats
	config   MonitoringConfig
	stopChan chan struct{}
}

type MonitoringConfig struct {
	Enabled        bool `toml:"enabled"`
	ReportInterval int  `toml:"report_interval"` // in seconds
	LogRequests    bool `toml:"log_requests"`
}

type Stats struct {
	mu               sync.RWMutex
	TotalRequests    int64
	SuccessRequests  int64
	ErrorRequests    int64
	NotFoundRequests int64
	StartTime        time.Time
	LastRequestTime  time.Time
	ResponseTimes    []time.Duration
}

func NewMonitoringPlugin() *MonitoringPlugin {
	metadata := gorgo.PluginMetadata{
		Name:        "monitoring",
		Version:     "1.0.0",
		Description: "Application monitoring and metrics collection plugin",
		Author:      "Gorgo Framework",
		Priority:    gorgo.PriorityLow, // Low priority so it loads last
		Tags:        []string{"monitoring", "metrics", "stats"},
	}

	return &MonitoringPlugin{
		BasePlugin: gorgo.NewBasePlugin(metadata),
		stats:      &Stats{StartTime: time.Now()},
		stopChan:   make(chan struct{}),
	}
}

// ConfigurablePlugin implementation
func (p *MonitoringPlugin) ValidateConfig(config map[string]interface{}) error {
	return nil // Configuration is optional
}

func (p *MonitoringPlugin) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":         true,
		"report_interval": 60,
		"log_requests":    true,
	}
}

// ServiceProvider implementation
func (p *MonitoringPlugin) GetServices() map[string]interface{} {
	return map[string]interface{}{
		"monitoring": p,
		"stats":      p.stats,
	}
}

// EventSubscriber implementation
func (p *MonitoringPlugin) GetEventSubscriptions() map[string]gorgo.EventHandler {
	return map[string]gorgo.EventHandler{
		"request.incoming":  p.onRequestIncoming,
		"request.completed": p.onRequestCompleted,
		"request.error":     p.onRequestError,
		"request.not_found": p.onRequestNotFound,
		"app.starting":      p.onAppStarting,
		"app.stopping":      p.onAppStopping,
		"server.started":    p.onServerStarted,
		"plugin.started":    p.onPluginStarted,
		"plugin.stopped":    p.onPluginStopped,
	}
}

func (p *MonitoringPlugin) onRequestIncoming(event *gorgo.Event) error {
	if !p.config.Enabled {
		return nil
	}

	p.stats.mu.Lock()
	p.stats.TotalRequests++
	p.stats.LastRequestTime = time.Now()
	p.stats.mu.Unlock()

	if p.config.LogRequests {
		method := event.Data["method"]
		path := event.Data["path"]
		ip := event.Data["ip"]
		log.Printf("Request: %s %s from %s", method, path, ip)
	}

	return nil
}

func (p *MonitoringPlugin) onRequestCompleted(event *gorgo.Event) error {
	if !p.config.Enabled {
		return nil
	}

	p.stats.mu.Lock()
	p.stats.SuccessRequests++
	p.stats.mu.Unlock()

	return nil
}

func (p *MonitoringPlugin) onRequestError(event *gorgo.Event) error {
	if !p.config.Enabled {
		return nil
	}

	p.stats.mu.Lock()
	p.stats.ErrorRequests++
	p.stats.mu.Unlock()

	method := event.Data["method"]
	path := event.Data["path"]
	errorMsg := event.Data["error"]
	log.Printf("Error: %s %s - %s", method, path, errorMsg)

	return nil
}

func (p *MonitoringPlugin) onRequestNotFound(event *gorgo.Event) error {
	if !p.config.Enabled {
		return nil
	}

	p.stats.mu.Lock()
	p.stats.NotFoundRequests++
	p.stats.mu.Unlock()

	return nil
}

func (p *MonitoringPlugin) onAppStarting(event *gorgo.Event) error {
	log.Println("Monitoring: Application is starting...")
	return nil
}

func (p *MonitoringPlugin) onAppStopping(event *gorgo.Event) error {
	log.Println("Monitoring: Application is stopping...")
	p.printFinalStats()
	return nil
}

func (p *MonitoringPlugin) onServerStarted(event *gorgo.Event) error {
	address := event.Data["address"]
	log.Printf("Monitoring: Server started on %s", address)
	return nil
}

func (p *MonitoringPlugin) onPluginStarted(event *gorgo.Event) error {
	pluginName := event.Data["plugin"]
	log.Printf("Monitoring: Plugin '%s' started", pluginName)
	return nil
}

func (p *MonitoringPlugin) onPluginStopped(event *gorgo.Event) error {
	pluginName := event.Data["plugin"]
	log.Printf("Monitoring: Plugin '%s' stopped", pluginName)
	return nil
}

// MiddlewareProvider implementation
func (p *MonitoringPlugin) GetMiddleware() []gorgo.MiddlewareFunc {
	if !p.config.Enabled {
		return nil
	}

	return []gorgo.MiddlewareFunc{
		p.responseTimeMiddleware(),
	}
}

func (p *MonitoringPlugin) responseTimeMiddleware() gorgo.MiddlewareFunc {
	return func(next gorgo.HandlerFunc) gorgo.HandlerFunc {
		return func(ctx *gorgo.Context) error {
			start := time.Now()

			err := next(ctx)

			duration := time.Since(start)
			p.stats.mu.Lock()
			p.stats.ResponseTimes = append(p.stats.ResponseTimes, duration)
			// Limit response times array size
			if len(p.stats.ResponseTimes) > 1000 {
				p.stats.ResponseTimes = p.stats.ResponseTimes[1:]
			}
			p.stats.mu.Unlock()

			// Add response time header
			ctx.Header("X-Response-Time", duration.String())

			return err
		}
	}
}

// Main plugin methods
func (p *MonitoringPlugin) Initialize(container *container.Container, config map[string]interface{}) error {
	p.config = MonitoringConfig{
		Enabled:        getBoolConfig(config, "enabled", true),
		ReportInterval: getIntConfig(config, "report_interval", 60),
		LogRequests:    getBoolConfig(config, "log_requests", true),
	}

	log.Printf("Monitoring Plugin: Initialized with report interval %d seconds", p.config.ReportInterval)
	return p.BasePlugin.Initialize(container, config)
}

func (p *MonitoringPlugin) Start(ctx context.Context) error {
	if p.config.Enabled {
		// Start periodic reporting
		go p.startPeriodicReporting()
		log.Println("Monitoring Plugin: Started periodic reporting")
	}

	return p.BasePlugin.Start(ctx)
}

func (p *MonitoringPlugin) Stop(ctx context.Context) error {
	close(p.stopChan)
	return p.BasePlugin.Stop(ctx)
}

// Additional methods
func (p *MonitoringPlugin) startPeriodicReporting() {
	ticker := time.NewTicker(time.Duration(p.config.ReportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.printStats()
		case <-p.stopChan:
			return
		}
	}
}

func (p *MonitoringPlugin) printStats() {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	uptime := time.Since(p.stats.StartTime)
	avgResponseTime := p.calculateAverageResponseTime()

	log.Printf(`
=== Monitoring Report ===
Uptime: %v
Total Requests: %d
Success Requests: %d
Error Requests: %d
Not Found Requests: %d
Average Response Time: %v
Last Request: %v ago
========================`,
		uptime,
		p.stats.TotalRequests,
		p.stats.SuccessRequests,
		p.stats.ErrorRequests,
		p.stats.NotFoundRequests,
		avgResponseTime,
		time.Since(p.stats.LastRequestTime),
	)
}

func (p *MonitoringPlugin) printFinalStats() {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	uptime := time.Since(p.stats.StartTime)
	avgResponseTime := p.calculateAverageResponseTime()

	log.Printf(`
=== Final Monitoring Report ===
Total Uptime: %v
Total Requests: %d
Success Rate: %.2f%%
Error Rate: %.2f%%
Not Found Rate: %.2f%%
Average Response Time: %v
===============================`,
		uptime,
		p.stats.TotalRequests,
		float64(p.stats.SuccessRequests)/float64(p.stats.TotalRequests)*100,
		float64(p.stats.ErrorRequests)/float64(p.stats.TotalRequests)*100,
		float64(p.stats.NotFoundRequests)/float64(p.stats.TotalRequests)*100,
		avgResponseTime,
	)
}

func (p *MonitoringPlugin) calculateAverageResponseTime() time.Duration {
	if len(p.stats.ResponseTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, rt := range p.stats.ResponseTimes {
		total += rt
	}

	return total / time.Duration(len(p.stats.ResponseTimes))
}

func (p *MonitoringPlugin) GetStats() *Stats {
	return p.stats
}

// Middleware for creating metrics endpoint
func (p *MonitoringPlugin) MetricsEndpointMiddleware(path string) gorgo.MiddlewareFunc {
	return func(next gorgo.HandlerFunc) gorgo.HandlerFunc {
		return func(ctx *gorgo.Context) error {
			if ctx.Path() == path {
				return p.handleMetricsEndpoint(ctx)
			}
			return next(ctx)
		}
	}
}

func (p *MonitoringPlugin) handleMetricsEndpoint(ctx *gorgo.Context) error {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	uptime := time.Since(p.stats.StartTime)
	avgResponseTime := p.calculateAverageResponseTime()

	metrics := gorgo.Map{
		"uptime_seconds":           uptime.Seconds(),
		"total_requests":           p.stats.TotalRequests,
		"success_requests":         p.stats.SuccessRequests,
		"error_requests":           p.stats.ErrorRequests,
		"not_found_requests":       p.stats.NotFoundRequests,
		"average_response_time_ms": avgResponseTime.Milliseconds(),
	}

	return ctx.JSON(metrics)
}

// Helper functions
func getBoolConfig(config map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := config[key].(bool); ok {
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
