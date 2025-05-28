package gorgo

import (
	"fmt"
	"gorgo/internal/config"
)

type App struct {
	plugins []Plugin
	config  config.Config
}

func New() *App {
	return &App{
		plugins: make([]Plugin, 0),
	}
}

func (a *App) Run() (*App, error) {
	config, err := a.loadConfig("config/app.toml")
	if err != nil {
		return nil, err
	}

	a.config = config

	for _, p := range a.plugins {
		if err := p.Configure(a.config); err != nil {
			return nil, err
		}
	}

	for _, p := range a.plugins {
		if err := p.Init(a); err != nil {
			return nil, err
		}
	}

	a.printBanner()

	return a, nil
}

func (a *App) AddPlugin(p Plugin) {
	a.plugins = append(a.plugins, p)
}

func (a *App) loadConfig(path string) (config.Config, error) {
	rawCfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}

	return rawCfg, nil
}

func (a *App) printBanner() {
	banner := `
 ██████╗  ██████╗ ██████╗  ██████╗  ██████╗ 
██╔════╝ ██╔═══██╗██╔══██╗██╔════╝ ██╔═══██╗
██║  ███╗██║   ██║██████╔╝██║  ███╗██║   ██║
██║   ██║██║   ██║██╔══██╗██║   ██║██║   ██║
╚██████╔╝╚██████╔╝██║  ██║╚██████╔╝╚██████╔╝
 ╚═════╝  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝  ╚═════╝ 

hello_world_example v0.0.1
Powered by Gorgo Framework
`
	fmt.Printf(banner)
}
