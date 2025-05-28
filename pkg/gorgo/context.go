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
}

func NewContext(ctx *fasthttp.RequestCtx, container *container.Container, plugins map[string]Plugin) *Context {
	return &Context{
		RequestCtx: ctx,
		container:  container,
		plugins:    plugins,
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
