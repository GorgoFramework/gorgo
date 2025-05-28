package web

import (
	"gorgo/pkg/gorgo"

	"github.com/valyala/fasthttp"
)

type WebConfig struct {
	Port string `toml:"port"`
}

type WebPlugin struct {
	server *fasthttp.Server
	app    *gorgo.App
	cfg    WebConfig
}
