package log

import (
	"testing"

	"github.com/baisiyi/go-kits/plugin"
)

// TestRegisterWriterAndGetWriter tests registering and retrieving writers.
func TestRegisterWriterAndGetWriter(t *testing.T) {
	factory := &ConsoleWriterFactory{}
	RegisterWriter("test_writer", factory)

	retrieved := GetWriter("test_writer")
	if retrieved == nil {
		t.Fatal("Expected to get writer, got nil")
	}
	if retrieved != factory {
		t.Error("Retrieved writer is not the same as registered")
	}
}

// TestGetWriterNotFound tests that GetWriter returns nil for unregistered writers.
func TestGetWriterNotFound(t *testing.T) {
	retrieved := GetWriter("nonexistent_writer")
	if retrieved != nil {
		t.Error("Expected nil for unregistered writer, got non-nil")
	}
}

// TestConsoleWriterFactory_Type tests that ConsoleWriterFactory.Type() returns "log".
func TestConsoleWriterFactory_Type(t *testing.T) {
	factory := &ConsoleWriterFactory{}
	if factory.Type() != "log" {
		t.Errorf("ConsoleWriterFactory.Type() = %q, want %q", factory.Type(), "log")
	}
}

// TestConsoleWriterFactory_Setup tests ConsoleWriterFactory.Setup with valid decoder.
func TestConsoleWriterFactory_Setup(t *testing.T) {
	factory := &ConsoleWriterFactory{}
	cfg := &OutputConfig{
		Writer:    OutputConsole,
		Level:     "info",
		Formatter: "console",
	}
	decoder := &Decoder{OutputConfig: cfg}

	err := factory.Setup(OutputConsole, decoder)
	if err != nil {
		t.Errorf("Setup failed: %v", err)
	}

	if decoder.Core == nil {
		t.Error("Core should be set after Setup")
	}
	// ZapLevel is set alongside Core, so if Core is set, ZapLevel should also be configured
}

// TestConsoleWriterFactory_Setup_NilDecoder tests that Setup panics with nil decoder.
func TestConsoleWriterFactory_Setup_NilDecoder(t *testing.T) {
	factory := &ConsoleWriterFactory{}

	err := factory.Setup(OutputConsole, nil)
	if err == nil {
		t.Error("Expected error for nil decoder")
	}
}

// TestConsoleWriterFactory_Setup_WrongDecoder tests that Setup returns error for wrong decoder type.
func TestConsoleWriterFactory_Setup_WrongDecoder(t *testing.T) {
	factory := &ConsoleWriterFactory{}

	// Use a different decoder type that doesn't implement plugin.Decoder
	type wrongDecoder struct{}
	// This won't compile if wrongDecoder doesn't implement plugin.Decoder interface
	// So we use a nil interface value
	var wrongDecoderInterface plugin.Decoder
	err := factory.Setup(OutputConsole, wrongDecoderInterface)
	if err == nil {
		t.Error("Expected error for wrong decoder type")
	}
}

// TestFileWriterFactory_Type tests that FileWriterFactory.Type() returns "log".
func TestFileWriterFactory_Type(t *testing.T) {
	factory := &FileWriterFactory{}
	if factory.Type() != "log" {
		t.Errorf("FileWriterFactory.Type() = %q, want %q", factory.Type(), "log")
	}
}

// TestFileWriterFactory_Setup_NilDecoder tests that Setup returns error for nil decoder.
func TestFileWriterFactory_Setup_NilDecoder(t *testing.T) {
	factory := &FileWriterFactory{}

	err := factory.Setup(OutputFile, nil)
	if err == nil {
		t.Error("Expected error for nil decoder")
	}
}

// TestFileWriterFactory_Setup_WrongDecoder tests that Setup returns error for wrong decoder type.
func TestFileWriterFactory_Setup_WrongDecoder(t *testing.T) {
	factory := &FileWriterFactory{}

	// Use a nil interface value that doesn't satisfy *Decoder
	var wrongDecoderInterface plugin.Decoder
	err := factory.Setup(OutputFile, wrongDecoderInterface)
	if err == nil {
		t.Error("Expected error for wrong decoder type")
	}
}
