package plugin

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	// SetupTimeout is the timeout for initialization of each plugin.
	SetupTimeout = 3 * time.Second

	// MaxPluginSize is the max number of plugins.
	MaxPluginSize = 1000
)

// Config is the configuration of all plugins. plugin type => { plugin name => plugin config }
type Config map[string]map[string]yaml.Node

// SetupClosables loads plugins and returns a function to close them in reverse order.
func (c Config) SetupClosables() (close func() error, err error) {
	plugins, status, err := c.loadPlugins()
	if err != nil {
		return nil, err
	}

	pluginInfos, closes, err := c.setupPlugins(plugins, status)
	if err != nil {
		return nil, err
	}

	if err := c.onFinish(pluginInfos); err != nil {
		return nil, err
	}

	return func() error {
		for i := len(closes) - 1; i >= 0; i-- {
			if err := closes[i](); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

func (c Config) loadPlugins() (chan pluginInfo, map[string]bool, error) {
	var (
		plugins = make(chan pluginInfo, MaxPluginSize)
		status  = make(map[string]bool)
	)
	for typ, factories := range c {
		for name, cfg := range factories {
			factory := Get(typ, name)
			if factory == nil {
				return nil, nil, fmt.Errorf("plugin %s:%s no registered or imported, do not configure", typ, name)
			}
			p := pluginInfo{
				factory: factory,
				typ:     typ,
				name:    name,
				cfg:     cfg,
			}
			select {
			case plugins <- p:
			default:
				return nil, nil, fmt.Errorf("plugin number exceed max limit:%d", len(plugins))
			}
			status[p.key()] = false
		}
	}
	return plugins, status, nil
}

func (c Config) setupPlugins(plugins chan pluginInfo, status map[string]bool) ([]pluginInfo, []func() error, error) {
	var (
		result []pluginInfo
		closes []func() error
		num    = len(plugins)
	)
	for num > 0 {
		for i := 0; i < num; i++ {
			p := <-plugins
			if deps, err := p.hasDependence(status); err != nil {
				return nil, nil, err
			} else if deps {
				plugins <- p
				continue
			}
			if err := p.setup(); err != nil {
				return nil, nil, err
			}
			if closer, ok := p.asCloser(); ok {
				closes = append(closes, closer.Close)
			}
			status[p.key()] = true
			result = append(result, p)
		}
		if len(plugins) == num {
			return nil, nil, fmt.Errorf("cycle depends, not plugin is setup")
		}
		num = len(plugins)
	}
	return result, closes, nil
}

func (c Config) onFinish(plugins []pluginInfo) error {
	for _, p := range plugins {
		if err := p.onFinish(); err != nil {
			return err
		}
	}
	return nil
}

// pluginInfo is the information of a plugin.
type pluginInfo struct {
	factory Factory
	typ     string
	name    string
	cfg     yaml.Node
}

// hasDependence decides if any other plugins that this plugin depends on haven't been initialized.
func (p *pluginInfo) hasDependence(status map[string]bool) (bool, error) {
	deps, ok := p.factory.(Depender)
	if ok {
		hasDeps, err := p.checkDependence(status, deps.DependsOn(), false)
		if err != nil {
			return false, err
		}
		if hasDeps {
			return true, nil
		}
	}
	fd, ok := p.factory.(FlexDepender)
	if ok {
		return p.checkDependence(status, fd.FlexDependsOn(), true)
	}
	return false, nil
}

// Depender is the interface for "Strong Dependence".
type Depender interface {
	DependsOn() []string
}

// FlexDepender is the interface for "Weak Dependence".
type FlexDepender interface {
	FlexDependsOn() []string
}

func (p *pluginInfo) checkDependence(status map[string]bool, dependences []string, flexible bool) (bool, error) {
	for _, name := range dependences {
		if name == p.key() {
			return false, errors.New("plugin not allowed to depend on itself")
		}
		setup, ok := status[name]
		if !ok {
			if flexible {
				continue
			}
			return false, fmt.Errorf("depends plugin %s not exists", name)
		}
		if !setup {
			return true, nil
		}
	}
	return false, nil
}

func (p *pluginInfo) setup() error {
	var (
		ch  = make(chan struct{})
		err error
	)
	go func() {
		err = p.factory.Setup(p.name, &YamlNodeDecoder{Node: &p.cfg})
		close(ch)
	}()
	select {
	case <-ch:
	case <-time.After(SetupTimeout):
		return fmt.Errorf("setup plugin %s timeout", p.key())
	}
	if err != nil {
		return fmt.Errorf("setup plugin %s error: %v", p.key(), err)
	}
	return nil
}

func (p *pluginInfo) key() string {
	return p.typ + "-" + p.name
}

func (p *pluginInfo) onFinish() error {
	f, ok := p.factory.(FinishNotifier)
	if !ok {
		return nil
	}
	return f.OnFinish(p.name)
}

// FinishNotifier is the interface used to notify that all plugins' loading has been done.
type FinishNotifier interface {
	OnFinish(name string) error
}

func (p *pluginInfo) asCloser() (Closer, bool) {
	closer, ok := p.factory.(Closer)
	return closer, ok
}

// Closer is the interface used to provide a close callback of a plugin.
type Closer interface {
	Close() error
}

// YamlNodeDecoder is a decoder for a yaml.Node of the yaml config file.
type YamlNodeDecoder struct {
	Node *yaml.Node
}

// Decode decodes a yaml.Node of the yaml config file.
func (d *YamlNodeDecoder) Decode(cfg any) error {
	if d.Node == nil {
		return errors.New("yaml node empty")
	}
	return d.Node.Decode(cfg)
}
