package gorgo

import (
	"encoding/json"

	"github.com/GorgoFramework/gorgo/internal/container"
	"github.com/valyala/fasthttp"
)

type Context struct {
	*fasthttp.RequestCtx

	container *container.Container
	plugins   map[string]Plugin
	params    map[string]string
}

func NewContext(ctx *fasthttp.RequestCtx, container *container.Container, plugins map[string]Plugin) *Context {
	return &Context{
		RequestCtx: ctx,
		container:  container,
		plugins:    plugins,
		params:     make(map[string]string),
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
	c.Response.Header.SetContentType("application/json")
	return json.NewEncoder(c.Response.BodyWriter()).Encode(data)
}

func (c *Context) String(data string) error {
	c.Response.Header.SetContentType("text/plain")
	c.SetBodyString(data)
	return nil
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
