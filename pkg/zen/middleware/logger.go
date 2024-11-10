package middleware

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
)

// TODO: custom flag for writing to a log file. All logs

// logger configuration for the zen framework
type LoggerConfig struct {
	// skip logging for specific paths
	SkipPaths []string

	// Custom log format function
	Formatter func(*zen.Context, time.Duration) string

	// Enable/diable file logging
	LogToFile bool

	// file path for logging to file
	LogFilePath string
}

var (
	fileLogger *log.Logger
	logFile    *os.File
	loggerMu   sync.Once
)

// DefaultLoggerConfig reutrns a LoggerConfig with default settings
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		LogToFile:   false,
		LogFilePath: "logs/zen.log",
	}
}

// Logger middleware logs the incoming HTTP request details
func Logger(config ...LoggerConfig) zen.HandlerFunc {
	cfg := DefaultLoggerConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Initialize file logger if enabled
	if cfg.LogToFile {
		if err := initFileLogger(cfg); err != nil {
			log.Printf("Failed to initialize file logger: %v", err)
		}
	}

	return func(c *zen.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// for path in SkipPaths
		for _, skipPath := range cfg.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		if raw != "" {
			path = path + "?" + raw
		}

		// Get status code color
		statusColor := zen.ColorForStatus(c.Writer.Status())
		methodColor := zen.GetMethodColor(c.Request.Method)

		consoleLog := fmt.Sprintf("%s %3d %s| %13v | %15s | %-7s %s %s\n",
			statusColor, c.Writer.Status(), reset,
			latency,
			c.ClientIP(),
			methodColor+c.Request.Method+reset,
			gray, path,
		)

		fileLog := fmt.Sprintf("%d | %13v | %15s | %-7s %s\n",
			c.Writer.Status(),
			latency,
			c.ClientIP(),
			c.Request.Method,
			path,
		)

		// Console output
		log.Print(consoleLog)

		// File output
		if cfg.LogToFile && fileLogger != nil {
			fileLogger.Print(fileLog)
		}
	}
}

// initFileLogger initializes the file logger with proper error handling
func initFileLogger(cfg LoggerConfig) error {
	if cfg.LogFilePath == "" {
		cfg.LogFilePath = "logs/zen.log" // Fallback if path is empty
	}

	// Ensure path is absolute
	if !filepath.IsAbs(cfg.LogFilePath) {
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %v", err)
		}
		cfg.LogFilePath = filepath.Join(cwd, cfg.LogFilePath)
	}

	var initErr error
	loggerMu.Do(func() {
		// Create directory structure
		dir := filepath.Dir(cfg.LogFilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create log directory: %v", err)
			return
		}

		// Open log file
		file, err := os.OpenFile(cfg.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %v", err)
			return
		}

		logFile = file
		fileLogger = log.New(file, "", log.LstdFlags)
		log.Printf("Initialized file logger at: %s", cfg.LogFilePath)
	
		setupCleanup()
	})

	return initErr
}

// Close gracefully closes the log file if it exists
func Close() error {
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %v", err)
		}
		logFile = nil
		fileLogger = nil
	}
	return nil
}

// setupCleanup sets up signal handling for graceful shutdown
func setupCleanup() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		if err := Close(); err != nil {
			log.Printf("Error closing logger: %v", err)
		}
		os.Exit(0)
	}()
}
