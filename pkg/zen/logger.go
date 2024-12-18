package zen

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	SUCCESS
	WARNING
	ERROR
	FATAL
)

type Log struct {
	logger *log.Logger
}

var defaultLogger = &Log{
	logger: log.New(os.Stdout, "", 0),
}

func NewLogger() *Log {
	return &Log{
		logger: log.New(os.Stdout, "", 0),
	}
}

func (l *Log) log(level LogLevel, color, prefix string, v ...interface{}) {
	if !IsDevMode() && level == DEBUG {
		return
	}

	timestamp := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprint(v...)

	l.logger.Printf("%s%s [%s] %s%s",
		color,
		timestamp,
		prefix,
		message,
		Reset,
	)
}

func (l *Log) logf(level LogLevel, color, prefix string, format string, v ...interface{}) {
	if !IsDevMode() && level == DEBUG {
		return
	}

	timestamp := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprintf(format, v...)

	l.logger.Printf("%s%s [%s] %s%s",
		color,
		timestamp,
		prefix,
		message,
		Reset,
	)
}

// Debug logs message with gray color
func (l *Log) Debug(v ...interface{}) {
	l.log(DEBUG, Gray, "DEBUG", v...)
}

// Debugf logs formatted message with gray color
func (l *Log) Debugf(format string, v ...interface{}) {
	l.logf(DEBUG, Gray, "DEBUG", format, v...)
}

// Info logs message with blue color
func (l *Log) Info(v ...interface{}) {
	l.log(INFO, Blue, "INFO", v...)
}

// Infof logs formatted message with blue color
func (l *Log) Infof(format string, v ...interface{}) {
	l.logf(INFO, Blue, "INFO", format, v...)
}

// Success logs message with green color
func (l *Log) Success(v ...interface{}) {
	l.log(SUCCESS, Green, "SUCCESS", v...)
}

// Successf logs formatted message with green color
func (l *Log) Successf(format string, v ...interface{}) {
	l.logf(SUCCESS, Green, "SUCCESS", format, v...)
}

// Warn logs message with yellow color
func (l *Log) Warn(v ...interface{}) {
	l.log(WARNING, Yellow, "WARNING", v...)
}

// Warnf logs formatted message with yellow color
func (l *Log) Warnf(format string, v ...interface{}) {
	l.logf(WARNING, Yellow, "WARNING", format, v...)
}

// Error logs message with red color
func (l *Log) Error(v ...interface{}) {
	l.log(ERROR, Red, "ERROR", v...)
}

// Errorf logs formatted message with red color
func (l *Log) Errorf(format string, v ...interface{}) {
	l.logf(ERROR, Red, "ERROR", format, v...)
}

// Fatal logs message with purple color and exits
func (l *Log) Fatal(v ...interface{}) {
	l.log(FATAL, Purple, "FATAL", v...)
	os.Exit(1)
}

// Fatalf logs formatted message with purple color and exits
func (l *Log) Fatalf(format string, v ...interface{}) {
	l.logf(FATAL, Purple, "FATAL", format, v...)
	os.Exit(1)
}

// Package-level functions using default logger
func Debug(v ...interface{})                   { defaultLogger.Debug(v...) }
func Debugf(format string, v ...interface{})   { defaultLogger.Debugf(format, v...) }
func Info(v ...interface{})                    { defaultLogger.Info(v...) }
func Infof(format string, v ...interface{})    { defaultLogger.Infof(format, v...) }
func Success(v ...interface{})                 { defaultLogger.Success(v...) }
func Successf(format string, v ...interface{}) { defaultLogger.Successf(format, v...) }
func Warn(v ...interface{})                    { defaultLogger.Warn(v...) }
func Warnf(format string, v ...interface{})    { defaultLogger.Warnf(format, v...) }
func Error(v ...interface{})                   { defaultLogger.Error(v...) }
func Errorf(format string, v ...interface{})   { defaultLogger.Errorf(format, v...) }
func Fatal(v ...interface{})                   { defaultLogger.Fatal(v...) }
func Fatalf(format string, v ...interface{})   { defaultLogger.Fatalf(format, v...) }

// logger configuration for the zen framework
type LoggerConfig struct {
	// skip logging for specific paths
	SkipPaths []string

	// Custom log format function
	Formatter func(*Context, time.Duration) string

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
func Logger(config ...LoggerConfig) HandlerFunc {
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

	return func(c *Context) {
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

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if raw != "" {
			path = path + "?" + raw
		}

		// Get status code color
		statusColor := ColorForStatus(c.Writer.Status())
		methodColor := GetMethodColor(c.Request.Method)

		consoleLog := fmt.Sprintf("%s %3d %s| %13v | %15s | %-7s %s %s\n",
			statusColor, c.Writer.Status(), Reset,
			latency,
			c.GetClientIP(),
			methodColor+c.Request.Method+Reset,
			Gray, path,
		)

		fileLog := fmt.Sprintf("%d | %13v | %15s | %-7s %s\n",
			c.Writer.Status(),
			latency,
			c.GetClientIP(),
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
		Infof("Initialized file logger at: %s", cfg.LogFilePath)
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
