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

// saveAndReplaceDefaultLogger saves the current default logger and replaces it for testing.
func saveAndReplaceDefaultLogger(t *testing.T, newLogger Logger) (restore func()) {
	t.Helper()
	oldLogger := DefaultLogger
	oldLoggers := loggers

	loggers = make(map[string]Logger)
	DefaultLogger = newLogger
	loggers[defaultLoggerName] = newLogger

	return func() {
		DefaultLogger = oldLogger
		loggers = oldLoggers
	}
}

// TestInfof tests the Infof convenience function.
func TestInfof(t *testing.T) {
	mock := &mockLogger{}
	restore := saveAndReplaceDefaultLogger(t, mock)
	defer restore()

	Infof("test %s", "message")
	if !mock.infofCalled {
		t.Error("Infof was not called on mock logger")
	}
}

// TestErrorf tests the Errorf convenience function.
func TestErrorf(t *testing.T) {
	mock := &mockLogger{}
	restore := saveAndReplaceDefaultLogger(t, mock)
	defer restore()

	Errorf("error %s", "message")
	if !mock.errorfCalled {
		t.Error("Errorf was not called on mock logger")
	}
}

// TestWarnf tests the Warnf convenience function.
func TestWarnf(t *testing.T) {
	mock := &mockLogger{}
	restore := saveAndReplaceDefaultLogger(t, mock)
	defer restore()

	Warnf("warn %s", "message")
	if !mock.warnfCalled {
		t.Error("Warnf was not called on mock logger")
	}
}

// TestDebugf tests the Debugf convenience function.
func TestDebugf(t *testing.T) {
	mock := &mockLogger{}
	restore := saveAndReplaceDefaultLogger(t, mock)
	defer restore()

	Debugf("debug %s", "message")
	if !mock.debugfCalled {
		t.Error("Debugf was not called on mock logger")
	}
}

// TestRegisterNilPanic tests that registering nil logger panics.
func TestRegisterNilPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when registering nil logger")
		}
	}()
	// Note: This test may affect other tests because it modifies global state.
	// We use a unique name to minimize conflicts.
	Register("nil_test", nil)
}

// TestRegisterDuplicate tests that duplicate registration (non-default) panics.
func TestRegisterDuplicate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when registering duplicate logger")
		}
	}()

	mock1 := &mockLogger{}
	mock2 := &mockLogger{}

	Register("duplicate_test", mock1)
	Register("duplicate_test", mock2)
}

// TestGet tests retrieving registered and unregistered loggers.
func TestGet(t *testing.T) {
	mock := &mockLogger{}
	Register("test_get", mock)

	retrieved := Get("test_get")
	if retrieved != mock {
		t.Error("Did not retrieve the registered logger")
	}

	notFound := Get("not_registered")
	if notFound != nil {
		t.Error("Expected nil for unregistered logger")
	}
}

// TestGetDefaultLogger tests that the default logger is not nil.
func TestGetDefaultLogger(t *testing.T) {
	logger := GetDefaultLogger()
	if logger == nil {
		t.Error("Default logger should not be nil")
	}
}
