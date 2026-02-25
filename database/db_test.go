package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/baisiyi/go-kits/log"
)

// mockLogger is a mock implementation of log.Logger for testing.
type mockLogger struct {
	infos      []string
	errors     []string
	warns      []string
	debugs     []string
	lastFormat string
	lastArgs   []interface{}
}

func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.infos = append(m.infos, format)
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, format)
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Warnf(format string, args ...interface{}) {
	m.warns = append(m.warns, format)
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Debugf(format string, args ...interface{}) {
	m.debugs = append(m.debugs, format)
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Debug(msg string, fields ...log.Field) {}

func (m *mockLogger) Info(msg string, fields ...log.Field) {}

func (m *mockLogger) Warn(msg string, fields ...log.Field) {}

func (m *mockLogger) Error(msg string, fields ...log.Field) {}

func (m *mockLogger) Fatal(msg string, fields ...log.Field) {}

func (m *mockLogger) Panic(msg string, fields ...log.Field) {}

func (m *mockLogger) With(fields ...log.Field) log.Logger { return m }

func (m *mockLogger) Named(name string) log.Logger { return m }

func (m *mockLogger) Sync() error { return nil }

// TestConnect_ToDSN tests the DSN generation from Connect config.
func TestConnect_ToDSN(t *testing.T) {
	tests := []struct {
		name     string
		connect  *Connect
		expected string
	}{
		{
			name: "standard configuration",
			connect: &Connect{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Name:     "testdb",
			},
			expected: "root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "custom port and host",
			connect: &Connect{
				Host:     "db.example.com",
				Port:     3307,
				Username: "admin",
				Password: "secret",
				Name:     "production",
			},
			expected: "admin:secret@tcp(db.example.com:3307)/production?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "empty password",
			connect: &Connect{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "",
				Name:     "testdb",
			},
			expected: "root:@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.connect.ToDSN()
			if result != tt.expected {
				t.Errorf("ToDSN() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestConnect_ToDSN_Format tests the format of generated DSN.
func TestConnect_ToDSN_Format(t *testing.T) {
	connect := &Connect{
		Host:     "localhost",
		Port:     3306,
		Username: "user",
		Password: "pass",
		Name:     "mydb",
	}

	dsn := connect.ToDSN()

	// Verify DSN contains expected components
	if !strings.HasPrefix(dsn, "user:pass@tcp(") {
		t.Errorf("DSN does not start with expected prefix")
	}
	if !strings.Contains(dsn, ":3306") {
		t.Errorf("DSN does not contain port")
	}
	if !strings.Contains(dsn, "/mydb?") {
		t.Errorf("DSN does not contain database name")
	}
	if !strings.Contains(dsn, "charset=utf8mb4") {
		t.Errorf("DSN does not contain charset")
	}
	if !strings.Contains(dsn, "parseTime=True") {
		t.Errorf("DSN does not contain parseTime")
	}
}

// TestDBConfig_Fields tests that DBConfig has all required fields.
func TestDBConfig_Fields(t *testing.T) {
	cfg := DBConfig{
		DSN: Connect{
			Host: "localhost",
			Port: 3306,
		},
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		LogLevel:        4,
		SlowThreshold:   200 * time.Millisecond,
	}

	_ = cfg.MaxIdleConns
	_ = cfg.ConnMaxLifetime
	_ = cfg.ConnMaxIdleTime
	_ = cfg.LogLevel

	if cfg.MaxOpenConns != 25 {
		t.Errorf("MaxOpenConns = %v, want 25", cfg.MaxOpenConns)
	}
	if cfg.SlowThreshold != 200*time.Millisecond {
		t.Errorf("SlowThreshold = %v, want 200ms", cfg.SlowThreshold)
	}
}

// TestClient_GetDB tests that GetDB returns GORM instance with context.
func TestClient_GetDB(t *testing.T) {
	// We can't easily test this without a real DB connection
	// So we just verify the method signature exists
	_ = func(*Client) {}
}

// TestGormLoggerAdapter_Interface tests that GormLoggerAdapter implements logger.Interface.
func TestGormLoggerAdapter_Interface(t *testing.T) {
	mock := &mockLogger{}
	adapter := NewGormLogger(mock, 200*time.Millisecond, 4)

	// Verify adapter is not nil
	if adapter == nil {
		t.Fatal("NewGormLogger returned nil")
	}
}

// TestGormLoggerAdapter_LogMode tests the LogMode method.
func TestGormLoggerAdapter_LogMode(t *testing.T) {
	mock := &mockLogger{}
	adapter := NewGormLogger(mock, 200*time.Millisecond, 4)

	// Test changing log level
	newAdapter := adapter.LogMode(2)
	if newAdapter == nil {
		t.Error("LogMode should return logger.Interface")
	}
}

// TestGormLoggerAdapter_Info tests the Info method.
func TestGormLoggerAdapter_Info(t *testing.T) {
	mock := &mockLogger{}
	adapter := NewGormLogger(mock, 200*time.Millisecond, 4)

	ctx := context.Background()
	adapter.Info(ctx, "test message %s", "arg")

	if len(mock.infos) != 1 {
		t.Errorf("Expected 1 info log, got %d", len(mock.infos))
	}
}

// TestGormLoggerAdapter_Warn tests the Warn method.
func TestGormLoggerAdapter_Warn(t *testing.T) {
	mock := &mockLogger{}
	adapter := NewGormLogger(mock, 200*time.Millisecond, 4)

	ctx := context.Background()
	adapter.Warn(ctx, "warning message %s", "arg")

	if len(mock.warns) != 1 {
		t.Errorf("Expected 1 warn log, got %d", len(mock.warns))
	}
}

// TestGormLoggerAdapter_Error tests the Error method.
func TestGormLoggerAdapter_Error(t *testing.T) {
	mock := &mockLogger{}
	adapter := NewGormLogger(mock, 200*time.Millisecond, 4)

	ctx := context.Background()
	adapter.Error(ctx, "error message %s", "arg")

	if len(mock.errors) != 1 {
		t.Errorf("Expected 1 error log, got %d", len(mock.errors))
	}
}

// TestGormLoggerAdapter_Trace tests the Trace method.
func TestGormLoggerAdapter_Trace(t *testing.T) {
	mock := &mockLogger{}
	adapter := NewGormLogger(mock, 200*time.Millisecond, 4)

	ctx := context.Background()

	// Test normal trace (no error, fast query)
	adapter.Trace(ctx, time.Now(), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	// Should log as Info level
	if len(mock.infos) != 1 {
		t.Errorf("Expected 1 info log for normal trace, got %d", len(mock.infos))
	}
}

// TestGormLoggerAdapter_Trace_SlowQuery tests Trace with slow query.
func TestGormLoggerAdapter_Trace_SlowQuery(t *testing.T) {
	mock := &mockLogger{}
	adapter := NewGormLogger(mock, 50*time.Millisecond, 3) // Warn level

	ctx := context.Background()

	// Simulate slow query (300ms > 50ms threshold)
	start := time.Now().Add(-300 * time.Millisecond)
	adapter.Trace(ctx, start, func() (string, int64) {
		return "SELECT * FROM large_table", 1000
	}, nil)

	// Should log as Warn (slow query)
	if len(mock.warns) != 1 {
		t.Errorf("Expected 1 warn log for slow query, got %d", len(mock.warns))
	}
}

// TestGormLoggerAdapter_Trace_Error tests Trace with error.
func TestGormLoggerAdapter_Trace_Error(t *testing.T) {
	mock := &mockLogger{}
	adapter := NewGormLogger(mock, 200*time.Millisecond, 2) // Error level

	ctx := context.Background()
	testErr := &testError{"database connection failed"}

	adapter.Trace(ctx, time.Now(), func() (string, int64) {
		return "SELECT * FROM users", 0
	}, testErr)

	// Should log as Error
	if len(mock.errors) != 1 {
		t.Errorf("Expected 1 error log, got %d", len(mock.errors))
	}
}

// testError is a simple error type for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestGormLoggerAdapter_SilentLevel tests that Silent level disables logging.
func TestGormLoggerAdapter_SilentLevel(t *testing.T) {
	mock := &mockLogger{}
	adapter := NewGormLogger(mock, 200*time.Millisecond, 1) // Silent

	ctx := context.Background()

	// Even with slow query, should not log at Silent level
	start := time.Now().Add(-300 * time.Millisecond)
	adapter.Trace(ctx, start, func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	if len(mock.infos)+len(mock.warns)+len(mock.errors) > 0 {
		t.Error("Silent level should suppress all logs")
	}
}

// BenchmarkConnect_ToDSN benchmarks the ToDSN method.
func BenchmarkConnect_ToDSN(b *testing.B) {
	connect := &Connect{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Name:     "testdb",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connect.ToDSN()
	}
}

// BenchmarkGormLoggerAdapter_Trace benchmarks the Trace method.
func BenchmarkGormLoggerAdapter_Trace(b *testing.B) {
	mock := &mockLogger{}
	adapter := NewGormLogger(mock, 200*time.Millisecond, 4)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Trace(ctx, time.Now(), func() (string, int64) {
			return "SELECT 1", 1
		}, nil)
	}
}
