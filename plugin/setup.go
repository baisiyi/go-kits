package plugin

import "gopkg.in/yaml.v3"

// Config is the configuration of all plugins. plugin type => { plugin name => plugin config }
type Config map[string]map[string]yaml.Node
