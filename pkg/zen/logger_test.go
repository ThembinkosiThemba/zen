package zen

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	// Create test directory
	testDir := t.TempDir()

	// Backup and restore original logger
	originalOutput := log.Writer()
	defer log.SetOutput(originalOutput)

	// Capture console output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	tests := []struct {
		name           string
		config         LoggerConfig
		method         string
		path           string
		query          string
		expectedStatus int
		checkLog       func(t *testing.T, log string)
	}{
		{
			name:           "Console logging only",
			config:         LoggerConfig{},
			method:         http.MethodGet,
			path:           "/test",
			expectedStatus: http.StatusOK,
			checkLog: func(t *testing.T, log string) {
				assert.Contains(t, log, "GET")
				assert.Contains(t, log, "/test")
				assert.Contains(t, log, "200")
			},
		},
		{
			name: "File logging enabled",
			config: LoggerConfig{
				LogToFile:   true,
				LogFilePath: filepath.Join(testDir, "test.log"),
			},
			method:         http.MethodPost,
			path:           "/test",
			query:          "key=value",
			expectedStatus: http.StatusOK,
			checkLog: func(t *testing.T, log string) {
				// Check console output
				assert.Contains(t, log, "POST")
				assert.Contains(t, log, "/test?key=value")

				// Check file output
				content, err := os.ReadFile(filepath.Join(testDir, "test.log"))
				require.NoError(t, err)
				fileLog := string(content)
				assert.Contains(t, fileLog, "POST")
				assert.Contains(t, fileLog, "/test?key=value")
				assert.NotContains(t, fileLog, "\x1b[") // No color codes
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear buffer
			buf.Reset()

			// Create test request
			w := httptest.NewRecorder()
			url := tt.path
			if tt.query != "" {
				url += "?" + tt.query
			}
			r := httptest.NewRequest(tt.method, url, nil)

			// Create context and run middleware
			c := NewContext(w, r)
			handler := Logger(tt.config)

			if strings.Contains(tt.path, "error") {
				c.Status(http.StatusInternalServerError)
			}

			handler(c)

			// Wait for file operations to complete
			time.Sleep(100 * time.Millisecond)

			// Verify logs
			logOutput := buf.String()
			tt.checkLog(t, logOutput)
		})
	}
}

func TestLoggerCleanup(t *testing.T) {
	testDir := t.TempDir()
	logPath := filepath.Join(testDir, "cleanup.log")

	config := LoggerConfig{
		LogToFile:   true,
		LogFilePath: logPath,
	}

	// Initialize logger
	handler := Logger(config)

	// Create and process a test request
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	c := NewContext(w, r)
	handler(c)

	// Verify file exists
	_, err := os.Stat(logPath)
 	require.NoError(t, err, "Log file should exist")

	// Test cleanup
	err = Close()
	require.NoError(t, err, "Cleanup should succeed")

	// Verify logger is reset
	assert.Nil(t, fileLogger, "File logger should be nil after cleanup")
	assert.Nil(t, logFile, "Log file should be nil after cleanup")
}

// TODO: test failing
func TestLoggerSignalHandling(t *testing.T) {
	testDir := t.TempDir()
	logPath := filepath.Join(testDir, "signal.log")

	config := LoggerConfig{
		LogToFile:   true,
		LogFilePath: logPath,
	}

	// Initialize logger
	handler := Logger(config)

	// Create and process a test request
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	c := NewContext(w, r)
	handler(c)

	// Wait for file operations
	time.Sleep(100 * time.Millisecond)

	// Verify file exists
	_, err := os.Stat(logPath)
	require.NoError(t, err, "Log file should exist")

	// Simulate SIGTERM
	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	// Send signal in goroutine to avoid blocking
	go func() {
		time.Sleep(100 * time.Millisecond)
		p.Signal(syscall.SIGTERM)
	}()

	// Wait briefly to allow signal handling
	time.Sleep(200 * time.Millisecond)

	// Cleanup
	err = Close()
	require.NoError(t, err)
}

func TestLoggerWithCustomIP(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Real-IP", "1.2.3.4")

	c := NewContext(w, r)
	handler := Logger()
	handler(c)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "1.2.3.4")
}

func TestLoggerWithSkipPaths(t *testing.T) {
	testDir := t.TempDir()
	logPath := filepath.Join(testDir, "skip.log")

	config := LoggerConfig{
		LogToFile:   true,
		LogFilePath: logPath,
		SkipPaths:   []string{"/skip"},
	}

	tests := []struct {
		path      string
		shouldLog bool
	}{
		{"/skip", false},
		{"/log", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			// Create test request
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tt.path, nil)
			c := NewContext(w, r)

			handler := Logger(config)
			handler(c)

			// Wait for file operations
			time.Sleep(100 * time.Millisecond)

			// Check log file
			content, err := os.ReadFile(logPath)
			require.NoError(t, err)
			logContent := string(content)

			if tt.shouldLog {
				assert.Contains(t, logContent, tt.path)
			} else {
				assert.NotContains(t, logContent, tt.path)
			}
		})
	}
}

func TestDefaultLoggerConfig(t *testing.T) {
	config := DefaultLoggerConfig()
	assert.False(t, config.LogToFile)
	assert.Equal(t, "logs/zen.log", config.LogFilePath)
}
