package plugin

var plugins = make(map[string]map[string]Factory) // type => name => factory

// Factory is the interface for plugin factory abstraction.
// Custom Plugins need to implement this interface to be registered as a plugin with certain type.
type Factory interface {
	// Type returns type of the plugin, i.e. selector, log, config, tracing.
	Type() string
	// Setup loads plugin by configuration.
	// The data structure of the configuration of the plugin needs to be defined in advance。
	Setup(name string, dec Decoder) error
}

// Decoder is the interface used to decode plugin configuration.
type Decoder interface {
	Decode(v interface{}) error // the input param is the custom configuration of the plugin
}

func Register(name string, f Factory) {
	factories, ok := plugins[f.Type()]
	if !ok {
		factories = make(map[string]Factory)
		plugins[f.Type()] = factories
	}
	factories[name] = f
}

func Get(typ string, name string) Factory {
	return plugins[typ][name]
}
