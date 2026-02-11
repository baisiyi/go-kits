package plugin

import (
	"testing"
)

// mockFactory is a mock implementation of Factory for testing.
type mockFactory struct {
	typ   string
	setupCalled bool
}

func (m *mockFactory) Type() string {
	return m.typ
}

func (m *mockFactory) Setup(name string, dec Decoder) error {
	m.setupCalled = true
	return nil
}

// TestRegisterAndGet tests that a registered factory can be retrieved.
func TestRegisterAndGet(t *testing.T) {
	// Clear the global plugins map before test.
	plugins = make(map[string]map[string]Factory)

	factory := &mockFactory{typ: "test"}
	Register("test_plugin", factory)

	retrieved := Get("test", "test_plugin")
	if retrieved == nil {
		t.Error("Expected to get factory, got nil")
	}
	if retrieved != factory {
		t.Error("Retrieved factory is not the same as registered")
	}
}

// TestGetNotFound tests that Get returns nil for unregistered plugins.
func TestGetNotFound(t *testing.T) {
	// Clear the global plugins map before test.
	plugins = make(map[string]map[string]Factory)

	retrieved := Get("nonexistent", "nonexistent")
	if retrieved != nil {
		t.Error("Expected nil for unregistered plugin, got non-nil")
	}
}

// TestMultiplePluginsSameType tests that multiple plugins of the same type are stored separately.
func TestMultiplePluginsSameType(t *testing.T) {
	// Clear the global plugins map before test.
	plugins = make(map[string]map[string]Factory)

	factory1 := &mockFactory{typ: "log"}
	factory2 := &mockFactory{typ: "log"}

	Register("logger1", factory1)
	Register("logger2", factory2)

	retrieved1 := Get("log", "logger1")
	retrieved2 := Get("log", "logger2")

	if retrieved1 != factory1 {
		t.Error("logger1 not retrieved correctly")
	}
	if retrieved2 != factory2 {
		t.Error("logger2 not retrieved correctly")
	}
}

// TestDifferentTypes tests that different types are isolated from each other.
func TestDifferentTypes(t *testing.T) {
	// Clear the global plugins map before test.
	plugins = make(map[string]map[string]Factory)

	logFactory := &mockFactory{typ: "log"}
	configFactory := &mockFactory{typ: "config"}

	Register("logger", logFactory)
	Register("cfg_loader", configFactory)

	logRetrieved := Get("log", "logger")
	configRetrieved := Get("config", "cfg_loader")

	if logRetrieved != logFactory {
		t.Error("log plugin not retrieved correctly")
	}
	if configRetrieved != configFactory {
		t.Error("config plugin not retrieved correctly")
	}

	// Verify they are different types
	if Get("log", "cfg_loader") != nil {
		t.Error("Cross-type retrieval should return nil")
	}
	if Get("config", "logger") != nil {
		t.Error("Cross-type retrieval should return nil")
	}
}

// TestRegisterOverwrite tests that duplicate registration overwrites the previous one.
func TestRegisterOverwrite(t *testing.T) {
	// Clear the global plugins map before test.
	plugins = make(map[string]map[string]Factory)

	factory1 := &mockFactory{typ: "test"}
	factory2 := &mockFactory{typ: "test"}

	Register("plugin", factory1)
	Register("plugin", factory2)

	retrieved := Get("test", "plugin")
	if retrieved != factory2 {
		t.Error("Expected second factory after overwrite, got first factory")
	}
}
