package log

import (
	"sync"
	"testing"
)

// mockLogger is a mock implementation of Logger for testing.
type mockLogger struct {
	mu           sync.Mutex
	infofCalled  bool
	errorfCalled bool
	warnfCalled  bool
	debugfCalled bool
	fatalfCalled bool
	panicfCalled bool
	infoCalled   bool
	lastMsg      string
	lastFields   []Field
	lastFormat   string
	lastArgs     []interface{}
}

func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infofCalled = true
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Errorf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorfCalled = true
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Warnf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnfCalled = true
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Debugf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debugfCalled = true
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Fatalf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fatalfCalled = true
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Panicf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.panicfCalled = true
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Debug(msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debugfCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockLogger) Info(msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infoCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockLogger) Warn(msg string, fields ...Field) {}

func (m *mockLogger) Error(msg string, fields ...Field) {}

func (m *mockLogger) Fatal(msg string, fields ...Field) {}

func (m *mockLogger) Panic(msg string, fields ...Field) {}

func (m *mockLogger) With(fields ...Field) Logger { return m }

func (m *mockLogger) Named(name string) Logger { return m }

func (m *mockLogger) Sync() error { return nil }

// TestInfof tests the Infof convenience function.
func TestInfof(t *testing.T) {
	mock := &mockLogger{}
	oldLogger := defaultLogger
	SetDefault(mock)
	defer SetDefault(oldLogger)

	Infof("test %s", "message")
	if !mock.infofCalled {
		t.Error("Infof was not called on mock logger")
	}
}

// TestErrorf tests the Errorf convenience function.
func TestErrorf(t *testing.T) {
	mock := &mockLogger{}
	oldLogger := defaultLogger
	SetDefault(mock)
	defer SetDefault(oldLogger)

	Errorf("error %s", "message")
	if !mock.errorfCalled {
		t.Error("Errorf was not called on mock logger")
	}
}

// TestWarnf tests the Warnf convenience function.
func TestWarnf(t *testing.T) {
	mock := &mockLogger{}
	oldLogger := defaultLogger
	SetDefault(mock)
	defer SetDefault(oldLogger)

	Warnf("warn %s", "message")
	if !mock.warnfCalled {
		t.Error("Warnf was not called on mock logger")
	}
}

// TestDebugf tests the Debugf convenience function.
func TestDebugf(t *testing.T) {
	mock := &mockLogger{}
	oldLogger := defaultLogger
	SetDefault(mock)
	defer SetDefault(oldLogger)

	Debugf("debug %s", "message")
	if !mock.debugfCalled {
		t.Error("Debugf was not called on mock logger")
	}
}

// TestInfo tests the Info convenience function.
func TestInfo(t *testing.T) {
	mock := &mockLogger{}
	oldLogger := defaultLogger
	SetDefault(mock)
	defer SetDefault(oldLogger)

	Info("test message", String("key", "value"))
	if !mock.infoCalled {
		t.Error("Info was not called on mock logger")
	}
}

// TestGetDefaultLogger tests that the default logger is not nil.
func TestGetDefaultLogger(t *testing.T) {
	logger := GetDefaultLogger()
	if logger == nil {
		t.Error("Default logger should not be nil")
	}
}
