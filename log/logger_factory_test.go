package log

import (
	"testing"

	"github.com/baisiyi/go-kits/plugin"
)

// TestFactoryType tests that Factory.Type() returns "log".
func TestFactoryType(t *testing.T) {
	factory := &Factory{}
	if factory.Type() != "log" {
		t.Errorf("Factory.Type() = %q, want %q", factory.Type(), "log")
	}
}

// TestFactorySetup_NilDecoder tests that Setup returns error for nil decoder.
func TestFactorySetup_NilDecoder(t *testing.T) {
	factory := &Factory{}

	err := factory.Setup("test", nil)
	if err == nil {
		t.Error("Expected error for nil decoder")
	}
}

// TestFactoryImplementsPluginFactory tests that Factory implements plugin.Factory.
func TestFactoryImplementsPluginFactory(t *testing.T) {
	var _ plugin.Factory = &Factory{}
}

// TestDecoder_DecodeForWriterFactory tests that Decoder.Decode works correctly
// when called by Writer factories (ConsoleWriterFactory/FileWriterFactory).
// This is the intended usage pattern: factories call decoder.Decode(&cfg) with **OutputConfig.
func TestDecoder_DecodeForWriterFactory(t *testing.T) {
	// Simulate how ConsoleWriterFactory uses Decoder
	cfg := &OutputConfig{
		Writer:    OutputConsole,
		Level:     "info",
		Formatter: "console",
	}
	decoder := &Decoder{OutputConfig: cfg}

	// This is what ConsoleWriterFactory.Setup does
	var decodedCfg *OutputConfig
	err := decoder.Decode(&decodedCfg)
	if err != nil {
		t.Errorf("Decoder.Decode failed: %v", err)
	}

	if decodedCfg == nil {
		t.Fatal("decodedCfg should not be nil")
	}

	if decodedCfg.Writer != OutputConsole {
		t.Errorf("Writer = %q, want %q", decodedCfg.Writer, OutputConsole)
	}
	if decodedCfg.Level != "info" {
		t.Errorf("Level = %q, want %q", decodedCfg.Level, "info")
	}
}

// TestDecoder_DecodeWrongType tests that Decoder.Decode returns error for wrong type.
func TestDecoder_DecodeWrongType(t *testing.T) {
	decoder := &Decoder{OutputConfig: nil}

	// Pass wrong type (*Config instead of **OutputConfig)
	var wrongType Config
	err := decoder.Decode(&wrongType)
	if err == nil {
		t.Error("Expected error for wrong type")
	}
}
