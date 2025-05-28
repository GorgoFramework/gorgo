package gorgo

type AppContext interface {
	AddPlugin(p Plugin)
}

type Plugin interface {
	Name() string
	Configure(cfg map[string]interface{}) error
	Init(app AppContext) error
	Shutdown() error
}
