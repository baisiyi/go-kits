package rollwriter

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestWithMaxAge tests the WithMaxAge option function.
func TestWithMaxAge(t *testing.T) {
	opts := &Options{}
	WithMaxAge(30)(opts)

	expected := 30 * 24 * time.Hour
	if opts.maxAge != expected {
		t.Errorf("maxAge = %v, want %v", opts.maxAge, expected)
	}
}

// TestWithRotationAge tests the WithRotationAge option function.
func TestWithRotationAge(t *testing.T) {
	opts := &Options{}
	WithRotationAge(12)(opts)

	expected := 12 * time.Hour
	if opts.rotationAge != expected {
		t.Errorf("rotationAge = %v, want %v", opts.rotationAge, expected)
	}
}

// TestWithRotationSize tests the WithRotationSize option function.
func TestWithRotationSize(t *testing.T) {
	opts := &Options{}
	WithRotationSize(50)(opts)

	if opts.rotationSize != 50 {
		t.Errorf("rotationSize = %v, want %v", opts.rotationSize, 50)
	}
}

// TestWithRotationCount tests the WithRotationCount option function.
func TestWithRotationCount(t *testing.T) {
	opts := &Options{}
	WithRotationCount(20)(opts)

	if opts.rotationCount != 20 {
		t.Errorf("rotationCount = %v, want %v", opts.rotationCount, 20)
	}
}

// TestWithTimeFormat tests the WithTimeFormat option function.
func TestWithTimeFormat(t *testing.T) {
	opts := &Options{}
	WithTimeFormat(".%Y-%m-%d")(opts)

	if opts.timeFormat != ".%Y-%m-%d" {
		t.Errorf("timeFormat = %v, want %v", opts.timeFormat, ".%Y-%m-%d")
	}
}

// TestWithMaxAgeDuration tests the WithMaxAgeDuration option function.
func TestWithMaxAgeDuration(t *testing.T) {
	opts := &Options{}
	WithMaxAgeDuration(168 * time.Hour)(opts)

	expected := 168 * time.Hour
	if opts.maxAge != expected {
		t.Errorf("maxAge = %v, want %v", opts.maxAge, expected)
	}
}

// TestWithRotationAgeDuration tests the WithRotationAgeDuration option function.
func TestWithRotationAgeDuration(t *testing.T) {
	opts := &Options{}
	WithRotationAgeDuration(6 * time.Hour)(opts)

	expected := 6 * time.Hour
	if opts.rotationAge != expected {
		t.Errorf("rotationAge = %v, want %v", opts.rotationAge, expected)
	}
}

// TestNewRollWriter tests creating a roll writer with a temporary directory.
func TestNewRollWriter(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	// Use only WithMaxAge (without RotationCount) to avoid conflict
	writer, err := NewRollWriter(filePath, WithMaxAge(7))
	if err != nil {
		t.Fatalf("NewRollWriter failed: %v", err)
	}
	defer writer.Sync()

	// Write some data
	n, err := writer.Write([]byte("test message"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 12 {
		t.Errorf("Write returned %d, want 12", n)
	}

	// Sync should not error
	if err := writer.Sync(); err != nil {
		t.Errorf("Sync failed: %v", err)
	}
}

// TestNewRollWriterWithOptions tests creating a roll writer with all options.
func TestNewRollWriterWithOptions(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_options.log")

	// Use RotationCount instead of MaxAge to avoid conflict
	// Also set maxAgeDuration to 0 to override the default
	writer, err := NewRollWriter(filePath,
		WithTimeFormat(".%Y%m%d"),
		WithMaxAgeDuration(0), // 禁用默认的 MaxAge
		WithRotationCount(5),
		WithRotationAge(12),
		WithRotationSize(10*1024*1024), // 10MB
	)
	if err != nil {
		t.Fatalf("NewRollWriter failed: %v", err)
	}
	defer writer.Sync()

	// Write some data
	testData := []byte("test message with options")
	n, err := writer.Write(testData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Write returned %d, want %d", n, len(testData))
	}

	// Verify the file was created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(files) == 0 {
		t.Error("Expected at least one log file to be created")
	}
}
