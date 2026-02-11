package log

import (
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

// TestLevels tests the Levels map for correct string to zapcore.Level mapping.
func TestLevels(t *testing.T) {
	tests := []struct {
		input    string
		expected zapcore.Level
	}{
		{"", zapcore.DebugLevel},
		{"trace", zapcore.DebugLevel},
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"fatal", zapcore.FatalLevel},
	}

	for _, tt := range tests {
		if Levels[tt.input] != tt.expected {
			t.Errorf("Levels[%q] = %v, want %v", tt.input, Levels[tt.input], tt.expected)
		}
	}
}

// TestGetLogEncoderKey tests the GetLogEncoderKey function.
func TestGetLogEncoderKey(t *testing.T) {
	tests := []struct {
		defKey   string
		key      string
		expected string
	}{
		{"T", "", "T"},       // Empty key returns default
		{"T", "Time", "Time"}, // Non-empty key returns key
		{"L", "Level", "Level"},
		{"", "Custom", "Custom"},
	}

	for _, tt := range tests {
		result := GetLogEncoderKey(tt.defKey, tt.key)
		if result != tt.expected {
			t.Errorf("GetLogEncoderKey(%q, %q) = %q, want %q", tt.defKey, tt.key, result, tt.expected)
		}
	}
}

// TestNewTimeEncoder tests the NewTimeEncoder function with various formats.
func TestNewTimeEncoder(t *testing.T) {
	_ = time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)

	t.Run("empty format", func(t *testing.T) {
		encoder := NewTimeEncoder("")
		if encoder == nil {
			t.Error("Expected non-nil encoder for empty format")
		}
	})

	t.Run("seconds format", func(t *testing.T) {
		encoder := NewTimeEncoder("seconds")
		if encoder == nil {
			t.Error("Expected non-nil encoder for seconds format")
		}
	})

	t.Run("milliseconds format", func(t *testing.T) {
		encoder := NewTimeEncoder("milliseconds")
		if encoder == nil {
			t.Error("Expected non-nil encoder for milliseconds format")
		}
	})

	t.Run("nanoseconds format", func(t *testing.T) {
		encoder := NewTimeEncoder("nanoseconds")
		if encoder == nil {
			t.Error("Expected non-nil encoder for nanoseconds format")
		}
	})

	t.Run("custom format", func(t *testing.T) {
		encoder := NewTimeEncoder("2006-01-02")
		if encoder == nil {
			t.Error("Expected non-nil encoder for custom format")
		}
		// Verify it doesn't panic - the actual formatting is tested via defaultTimeFormat
	})
}

// TestDefaultTimeFormat tests the defaultTimeFormat output.
func TestDefaultTimeFormat(t *testing.T) {
	// Use local timezone to avoid timezone conversion issues
	loc, _ := time.LoadLocation("Local")
	testTime := time.Date(2024, 1, 15, 10, 30, 45, 123456789, loc)

	result := defaultTimeFormat(testTime)

	// Expected format: "2024-01-15 10:30:45.123"
	expected := "2024-01-15 10:30:45.123"

	if len(result) != 23 {
		t.Errorf("defaultTimeFormat returned %d bytes, want 23", len(result))
	}

	if string(result) != expected {
		t.Errorf("defaultTimeFormat returned %q, want %q", string(result), expected)
	}
}

// TestNewZapLog tests creating a ZapLogger with console output.
func TestNewZapLog(t *testing.T) {
	// This should not panic with valid console config
	cfg := Config{{
		Writer:     OutputConsole,
		Level:      "info",
		Formatter:  "console",
		EnableColor: false,
	}}

	logger := NewZapLog(cfg)
	if logger == nil {
		t.Error("Expected non-nil ZapLogger")
	}

	// Verify it implements Logger interface
	var _ Logger = logger
}

// TestZapLoggerMethods tests calling logging methods on ZapLogger.
func TestZapLoggerMethods(t *testing.T) {
	cfg := Config{{
		Writer:    OutputConsole,
		Level:     "debug",
		Formatter: "console",
	}}

	logger := NewZapLog(cfg)
	if logger == nil {
		t.Fatal("Failed to create ZapLogger")
	}

	// These should not panic
	logger.Infof("test info message %s", "arg")
	logger.Errorf("test error message %s", "arg")
	logger.Warnf("test warn message %s", "arg")
	logger.Debugf("test debug message %s", "arg")
}

// TestRegisterFormatEncoder tests registering a custom format encoder.
func TestRegisterFormatEncoder(t *testing.T) {
	// Register a custom encoder
	customEncoderCalled := false
	RegisterFormatEncoder("custom_test", func(cfg zapcore.EncoderConfig) zapcore.Encoder {
		customEncoderCalled = true
		return zapcore.NewConsoleEncoder(cfg)
	})

	// Verify it was registered
	if _, ok := formatEncoders["custom_test"]; !ok {
		t.Error("Custom encoder was not registered")
	}

	// Use it
	cfg := Config{{
		Writer:    OutputConsole,
		Level:     "info",
		Formatter: "custom_test",
	}}

	_ = NewZapLog(cfg)

	if !customEncoderCalled {
		t.Error("Custom encoder was not called")
	}

	// Clean up
	delete(formatEncoders, "custom_test")
}
