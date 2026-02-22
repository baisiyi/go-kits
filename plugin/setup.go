package plugin

import (
	"errors"

	"gopkg.in/yaml.v3"
)

// Config is the configuration of all plugins. plugin type => { plugin name => plugin config }
type Config map[string]map[string]yaml.Node

// YamlNodeDecoder is a decoder for a yaml.Node of the yaml config file.
type YamlNodeDecoder struct {
	Node *yaml.Node
}

// Decode decodes a yaml.Node of the yaml config file.
func (d *YamlNodeDecoder) Decode(cfg interface{}) error {
	if d.Node == nil {
		return errors.New("yaml node empty")
	}
	return d.Node.Decode(cfg)
}
