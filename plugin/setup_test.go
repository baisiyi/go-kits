package plugin

import (
	"errors"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

// mockFactoryWithConfig is a mock factory that can return custom config.
type mockFactoryWithConfig struct {
	typ       string
	setupFunc func(name string, dec Decoder) error
}

func (m *mockFactoryWithConfig) Type() string {
	return m.typ
}

func (m *mockFactoryWithConfig) Setup(name string, dec Decoder) error {
	if m.setupFunc != nil {
		return m.setupFunc(name, dec)
	}
	return nil
}

// mockCloserFactory is a mock factory that implements Closer interface.
type mockCloserFactory struct {
	mockFactoryWithConfig
	closeCalled bool
}

func (m *mockCloserFactory) Close() error {
	m.closeCalled = true
	return nil
}

// mockFinishNotifierFactory is a mock factory that implements FinishNotifier interface.
type mockFinishNotifierFactory struct {
	mockFactoryWithConfig
	onFinishCalled bool
}

func (m *mockFinishNotifierFactory) OnFinish(name string) error {
	m.onFinishCalled = true
	return nil
}

// mockDependerFactory is a mock factory that implements Depender interface.
type mockDependerFactory struct {
	mockFactoryWithConfig
	dependsOn []string
}

func (m *mockDependerFactory) DependsOn() []string {
	return m.dependsOn
}

// mockFlexDependerFactory is a mock factory that implements FlexDepender interface.
type mockFlexDependerFactory struct {
	mockFactoryWithConfig
	flexDependsOn []string
}

func (m *mockFlexDependerFactory) FlexDependsOn() []string {
	return m.flexDependsOn
}

// TestSetupClosablesBasic tests basic plugin setup.
func TestSetupClosablesBasic(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	factory := &mockFactoryWithConfig{typ: "log"}
	Register("default", factory)

	config := Config{
		"log": {
			"default": yaml.Node{},
		},
	}

	closeFunc, err := config.SetupClosables()
	if err != nil {
		t.Fatalf("SetupClosables failed: %v", err)
	}
	if closeFunc == nil {
		t.Fatal("Expected closeFunc, got nil")
	}

	// Test close function
	if err := closeFunc(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

// TestSetupClosablesNotRegistered tests error when plugin is not registered.
func TestSetupClosablesNotRegistered(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	config := Config{
		"log": {
			"default": yaml.Node{},
		},
	}

	_, err := config.SetupClosables()
	if err == nil {
		t.Fatal("Expected error for unregistered plugin")
	}
}

// TestSetupClosablesWithConfig tests that config is properly decoded.
func TestSetupClosablesWithConfig(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	var receivedConfig map[string]any
	factory := &mockFactoryWithConfig{
		typ: "log",
		setupFunc: func(name string, dec Decoder) error {
			return dec.Decode(&receivedConfig)
		},
	}
	Register("default", factory)

	yamlContent := `level: debug
output: console`
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal yaml: %v", err)
	}

	config := Config{
		"log": {
			"default": node,
		},
	}

	_, err := config.SetupClosables()
	if err != nil {
		t.Fatalf("SetupClosables failed: %v", err)
	}

	if receivedConfig["level"] != "debug" {
		t.Errorf("Expected level to be debug, got %v", receivedConfig["level"])
	}
}

// TestSetupClosablesWithCloser tests that Closer interface is properly handled.
func TestSetupClosablesWithCloser(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	factory := &mockCloserFactory{
		mockFactoryWithConfig: mockFactoryWithConfig{typ: "log"},
	}
	Register("default", factory)

	config := Config{
		"log": {
			"default": yaml.Node{},
		},
	}

	closeFunc, err := config.SetupClosables()
	if err != nil {
		t.Fatalf("SetupClosables failed: %v", err)
	}

	// Closer.Close is called when the closeFunc is invoked, not during setup
	if factory.closeCalled {
		t.Error("Close should not be called during setup")
	}

	// Test close function calls factory close
	if err := closeFunc(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if !factory.closeCalled {
		t.Error("Expected close to be called")
	}
}

// TestSetupClosablesWithFinishNotifier tests that FinishNotifier is properly called.
func TestSetupClosablesWithFinishNotifier(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	factory := &mockFinishNotifierFactory{
		mockFactoryWithConfig: mockFactoryWithConfig{typ: "log"},
	}
	Register("default", factory)

	config := Config{
		"log": {
			"default": yaml.Node{},
		},
	}

	_, err := config.SetupClosables()
	if err != nil {
		t.Fatalf("SetupClosables failed: %v", err)
	}

	if !factory.onFinishCalled {
		t.Error("Expected OnFinish to be called")
	}
}

// TestSetupClosablesWithDepender tests strong dependency between plugins.
func TestSetupClosablesWithDepender(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	var callOrder []string
	mu := sync.Mutex{}

	// factoryA depends on factoryB (log-default), so it should be setup after factoryB
	factoryA := &mockDependerFactory{
		mockFactoryWithConfig: mockFactoryWithConfig{
			typ: "log",
			setupFunc: func(name string, dec Decoder) error {
				mu.Lock()
				callOrder = append(callOrder, "A")
				mu.Unlock()
				return nil
			},
		},
		dependsOn: []string{"log-default"},
	}
	Register("dependent", factoryA) // key = "log-dependent"

	// factoryB has no dependencies
	factoryB := &mockFactoryWithConfig{
		typ: "log",
		setupFunc: func(name string, dec Decoder) error {
			mu.Lock()
			callOrder = append(callOrder, "B")
			mu.Unlock()
			return nil
		},
	}
	Register("default", factoryB) // key = "log-default"

	config := Config{
		"log": {
			"dependent": yaml.Node{}, // factoryA
			"default":   yaml.Node{}, // factoryB
		},
	}

	_, err := config.SetupClosables()
	if err != nil {
		t.Fatalf("SetupClosables failed: %v", err)
	}

	// default (factoryB) should be setup before dependent (factoryA)
	if len(callOrder) != 2 {
		t.Fatalf("Expected 2 calls, got %d", len(callOrder))
	}
	if callOrder[0] != "B" || callOrder[1] != "A" {
		t.Errorf("Expected call order [B, A], got %v", callOrder)
	}
}

// TestSetupClosablesWithFlexDepender tests weak dependency between plugins.
func TestSetupClosablesWithFlexDepender(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	var callOrder []string
	mu := sync.Mutex{}

	factoryA := &mockFlexDependerFactory{
		mockFactoryWithConfig: mockFactoryWithConfig{
			typ: "log",
			setupFunc: func(name string, dec Decoder) error {
				mu.Lock()
				callOrder = append(callOrder, "A")
				mu.Unlock()
				return nil
			},
		},
		flexDependsOn: []string{"log-missing"},
	}
	Register("default", factoryA)

	factoryB := &mockFactoryWithConfig{
		typ: "log",
		setupFunc: func(name string, dec Decoder) error {
			mu.Lock()
			callOrder = append(callOrder, "B")
			mu.Unlock()
			return nil
		},
	}
	Register("optional", factoryB)

	config := Config{
		"log": {
			"default":  yaml.Node{},
			"optional": yaml.Node{},
		},
	}

	_, err := config.SetupClosables()
	if err != nil {
		t.Fatalf("SetupClosables failed: %v", err)
	}

	// Should still work even though flex dependency doesn't exist
	if len(callOrder) != 2 {
		t.Fatalf("Expected 2 calls, got %d", len(callOrder))
	}
}

// TestSetupClosablesCycleDependence tests cycle dependency detection.
func TestSetupClosablesCycleDependence(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	factoryA := &mockDependerFactory{
		mockFactoryWithConfig: mockFactoryWithConfig{typ: "log"},
		dependsOn:            []string{"log-B"},
	}
	Register("A", factoryA)

	factoryB := &mockDependerFactory{
		mockFactoryWithConfig: mockFactoryWithConfig{typ: "log"},
		dependsOn:            []string{"log-A"},
	}
	Register("B", factoryB)

	config := Config{
		"log": {
			"A": yaml.Node{},
			"B": yaml.Node{},
		},
	}

	_, err := config.SetupClosables()
	if err == nil {
		t.Fatal("Expected error for cycle dependency")
	}
}

// TestSetupClosablesSelfDependence tests self-dependency detection.
func TestSetupClosablesSelfDependence(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	factory := &mockDependerFactory{
		mockFactoryWithConfig: mockFactoryWithConfig{typ: "log"},
		dependsOn:            []string{"log-self"},
	}
	Register("self", factory)

	config := Config{
		"log": {
			"self": yaml.Node{},
		},
	}

	_, err := config.SetupClosables()
	if err == nil {
		t.Fatal("Expected error for self dependency")
	}
	if err.Error() != "plugin not allowed to depend on itself" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestYamlNodeDecoderDecode tests YamlNodeDecoder Decode method.
func TestYamlNodeDecoderDecode(t *testing.T) {
	yamlContent := `level: debug
output: console`
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal yaml: %v", err)
	}

	decoder := &YamlNodeDecoder{Node: &node}

	type Config struct {
		Level  string `yaml:"level"`
		Output string `yaml:"output"`
	}

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if cfg.Level != "debug" {
		t.Errorf("Expected level 'debug', got '%s'", cfg.Level)
	}
	if cfg.Output != "console" {
		t.Errorf("Expected output 'console', got '%s'", cfg.Output)
	}
}

// TestYamlNodeDecoderDecodeEmptyNode tests decoding with empty node.
func TestYamlNodeDecoderDecodeEmptyNode(t *testing.T) {
	decoder := &YamlNodeDecoder{Node: nil}

	var cfg map[string]any
	err := decoder.Decode(&cfg)
	if err == nil {
		t.Fatal("Expected error for empty node")
	}
	if err.Error() != "yaml node empty" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestPluginInfoKey tests pluginInfo key generation.
func TestPluginInfoKey(t *testing.T) {
	info := pluginInfo{
		typ:  "log",
		name: "default",
	}

	key := info.key()
	if key != "log-default" {
		t.Errorf("Expected key 'log-default', got '%s'", key)
	}
}

// TestMultiplePlugins tests setup of multiple plugins.
func TestMultiplePlugins(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	setupCalls := &sync.Map{}

	factoryLog := &mockFactoryWithConfig{
		typ: "log",
		setupFunc: func(name string, dec Decoder) error {
			setupCalls.Store("log", true)
			return nil
		},
	}
	Register("default", factoryLog)

	factoryConfig := &mockFactoryWithConfig{
		typ: "config",
		setupFunc: func(name string, dec Decoder) error {
			setupCalls.Store("config", true)
			return nil
		},
	}
	Register("default", factoryConfig)

	config := Config{
		"log":    {"default": yaml.Node{}},
		"config": {"default": yaml.Node{}},
	}

	_, err := config.SetupClosables()
	if err != nil {
		t.Fatalf("SetupClosables failed: %v", err)
	}

	if _, ok := setupCalls.Load("log"); !ok {
		t.Error("log plugin not setup")
	}
	if _, ok := setupCalls.Load("config"); !ok {
		t.Error("config plugin not setup")
	}
}

// TestSetupErrorPropagation tests that setup errors are propagated.
func TestSetupErrorPropagation(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	factory := &mockFactoryWithConfig{
		typ: "log",
		setupFunc: func(name string, dec Decoder) error {
			return errors.New("setup failed")
		},
	}
	Register("default", factory)

	config := Config{
		"log": {
			"default": yaml.Node{},
		},
	}

	_, err := config.SetupClosables()
	if err == nil {
		t.Fatal("Expected error to be propagated")
	}
}

// TestEmptyConfig tests setup with empty config.
func TestEmptyConfig(t *testing.T) {
	plugins = make(map[string]map[string]Factory)

	config := Config{}

	closeFunc, err := config.SetupClosables()
	if err != nil {
		t.Fatalf("SetupClosables failed: %v", err)
	}
	if closeFunc == nil {
		t.Fatal("Expected closeFunc for empty config")
	}

	if err := closeFunc(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}
