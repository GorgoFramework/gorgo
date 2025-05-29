package gorgo

import (
	"encoding/json"
	"mime/multipart"
	"sync"

	"github.com/GorgoFramework/gorgo/internal/container"
	"github.com/valyala/fasthttp"
)

type Context struct {
	fastCtx *fasthttp.RequestCtx

	container *container.Container
	plugins   map[string]Plugin
	params    map[string]string
	data      map[string]interface{} // Additional data
	mu        sync.RWMutex
}

func NewContext(ctx *fasthttp.RequestCtx, container *container.Container, plugins map[string]Plugin) *Context {
	return &Context{
		fastCtx:   ctx,
		container: container,
		plugins:   plugins,
		params:    make(map[string]string),
		data:      make(map[string]interface{}),
	}
}

func (c *Context) GetService(name string) (interface{}, bool) {
	return c.container.Get(name)
}

func (c *Context) GetPlugin(name string) (Plugin, bool) {
	plugin, ok := c.plugins[name]
	return plugin, ok
}

func (c *Context) JSON(data Map) error {
	c.fastCtx.Response.Header.SetContentType("application/json")
	return json.NewEncoder(c.fastCtx.Response.BodyWriter()).Encode(data)
}

func (c *Context) String(data string) error {
	c.fastCtx.Response.Header.SetContentType("text/plain")
	c.fastCtx.SetBodyString(data)
	return nil
}

func (c *Context) HTML(html string) error {
	c.fastCtx.Response.Header.SetContentType("text/html")
	c.fastCtx.SetBodyString(html)
	return nil
}

func (c *Context) XML(data interface{}) error {
	c.fastCtx.Response.Header.SetContentType("application/xml")
	// Simple XML serialization (can be improved)
	c.fastCtx.SetBodyString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<response></response>")
	return nil
}

func (c *Context) Status(code int) *Context {
	c.fastCtx.SetStatusCode(code)
	return c
}

func (c *Context) Header(key, value string) *Context {
	c.fastCtx.Response.Header.Set(key, value)
	return c
}

func (c *Context) Cookie(cookie *fasthttp.Cookie) *Context {
	c.fastCtx.Response.Header.SetCookie(cookie)
	return c
}

func (c *Context) GetHeader(key string) string {
	return string(c.fastCtx.Request.Header.Peek(key))
}

func (c *Context) GetCookie(key string) string {
	return string(c.fastCtx.Request.Header.Cookie(key))
}

func (c *Context) Param(key string) string {
	return c.params[key]
}

func (c *Context) ParamDefault(key, defaultValue string) string {
	if value, exists := c.params[key]; exists {
		return value
	}
	return defaultValue
}

func (c *Context) HasParam(key string) bool {
	_, exists := c.params[key]
	return exists
}

func (c *Context) Params() map[string]string {
	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range c.params {
		result[k] = v
	}
	return result
}

func (c *Context) SetParam(key, value string) {
	c.params[key] = value
}

// Methods for working with additional data
func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *Context) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.data[key]
	return value, exists
}

func (c *Context) GetString(key string) string {
	if value, exists := c.Get(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (c *Context) GetInt(key string) int {
	if value, exists := c.Get(key); exists {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return 0
}

func (c *Context) GetBool(key string) bool {
	if value, exists := c.Get(key); exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// Methods for working with query parameters
func (c *Context) Query(key string) string {
	return string(c.fastCtx.QueryArgs().Peek(key))
}

func (c *Context) QueryDefault(key, defaultValue string) string {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func (c *Context) QueryInt(key string) int {
	return c.fastCtx.QueryArgs().GetUintOrZero(key)
}

func (c *Context) QueryBool(key string) bool {
	return c.fastCtx.QueryArgs().GetBool(key)
}

// Methods for working with forms
func (c *Context) FormValue(key string) string {
	return string(c.fastCtx.FormValue(key))
}

// FormFile returns the first file for the provided form key
func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	return c.fastCtx.FormFile(key)
}

// Methods for working with request body
func (c *Context) Body() []byte {
	return c.fastCtx.Request.Body()
}

func (c *Context) BodyString() string {
	return string(c.Body())
}

func (c *Context) BindJSON(v interface{}) error {
	return json.Unmarshal(c.Body(), v)
}

// Methods for redirects
func (c *Context) Redirect(url string, statusCode int) error {
	c.fastCtx.Redirect(url, statusCode)
	return nil
}

// Methods for working with IP
func (c *Context) ClientIP() string {
	return c.fastCtx.RemoteIP().String()
}

func (c *Context) UserAgent() string {
	return string(c.fastCtx.UserAgent())
}

// Methods for working with path
func (c *Context) Path() string {
	return string(c.fastCtx.Path())
}

func (c *Context) Method() string {
	return string(c.fastCtx.Method())
}

func (c *Context) URL() string {
	return c.fastCtx.URI().String()
}

// Get base FastHTTP context (for advanced usage)
func (c *Context) FastHTTP() *fasthttp.RequestCtx {
	return c.fastCtx
}
