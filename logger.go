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

// LogLevel represents the severity level of log messages.
type LogLevel int

// Constants representing different log levels.
// These values are used to define the severity of log messages.
const (
	DEBUG   LogLevel = iota // DEBUG represents detailed logs, typically used for troubleshooting.
	INFO                    // INFO represents general operational messages to track the system state.
	SUCCESS                 // SUCCESS represents successful operations or significant positive outcomes.
	WARNING                 // WARNING represents warnings about potential issues or non-critical problems.
	ERROR                   // ERROR represents serious issues that may cause malfunctions but can be handled.
	FATAL                   // FATAL represents critical errors causing the program to terminate.
)

// Log is a wrapper around the standard log.Logger that is used to output log messages.
// It provides additional functionality (e.g., log levels) in the context of application logging.
type Log struct {
	logger *log.Logger // The internal logger that handles actual log message writing.
}

// defaultLogger is a global instance of Log that writes log messages to the standard output (os.Stdout).
// It uses a default logger initialized with no prefix and no flags.
var defaultLogger = &Log{
	logger: log.New(os.Stdout, "", 0), // log.New creates a new logger that writes to os.Stdout with no prefix or flags.
}

// NewLogger creates and returns a new instance of Log. The logger is initialized to write to standard output.
func NewLogger() *Log {
	return &Log{
		logger: log.New(os.Stdout, "", 0), // Initializes the logger to write to os.Stdout with no prefix or flags.
	}
}

// TODO: also use file logger for the log and logf functions

// Package-level functions using default logger

// Debug logs a debug level message using the default logger.
func Debug(v ...interface{}) { defaultLogger.Debug(v...) }

// Debugf logs a formatted debug level message using the default logger.
func Debugf(format string, v ...interface{}) { defaultLogger.Debugf(format, v...) }

// Info logs an info level message using the default logger.
func Info(v ...interface{}) { defaultLogger.Info(v...) }

// Infof logs a formatted info level message using the default logger.
func Infof(format string, v ...interface{}) { defaultLogger.Infof(format, v...) }

// Success logs a success level message using the default logger.
func Success(v ...interface{}) { defaultLogger.Success(v...) }

func Successf(format string, v ...interface{}) { defaultLogger.Successf(format, v...) }

// Warn logs a warning level message using the default logger.
func Warn(v ...interface{}) { defaultLogger.Warn(v...) }

// Warnf logs a formatted warning level message using the default logger.
func Warnf(format string, v ...interface{}) { defaultLogger.Warnf(format, v...) }

// Error logs an error level message using the default logger.
func Error(v ...interface{}) { defaultLogger.Error(v...) }

// Errorf logs a formatted error level message using the default logger.
func Errorf(format string, v ...interface{}) { defaultLogger.Errorf(format, v...) }

// Fatal logs a fatal level message using the default logger and exits the application.
func Fatal(v ...interface{}) { defaultLogger.Fatal(v...) }

// Fatalf logs a formatted fatal level message using the default logger and exits the application.
func Fatalf(format string, v ...interface{}) { defaultLogger.Fatalf(format, v...) }

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

	if IsFileLoggingEnabled(){
		fileLogger.Printf("%s [%s] %s",
			timestamp,
			prefix,
			message,
		)
	}
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

	if IsFileLoggingEnabled(){
		fileLogger.Printf("%s [%s] %s",
			timestamp,
			prefix,
			message,
		)
	}
}

func (l *Log) Debug(v ...interface{}) {
	l.log(DEBUG, Gray, "DEBUG", v...)
}

func (l *Log) Debugf(format string, v ...interface{}) {
	l.logf(DEBUG, Gray, "DEBUG", format, v...)
}

func (l *Log) Info(v ...interface{}) {
	l.log(INFO, Blue, "INFO", v...)
}

func (l *Log) Infof(format string, v ...interface{}) {
	l.logf(INFO, Blue, "INFO", format, v...)
}

func (l *Log) Success(v ...interface{}) {
	l.log(SUCCESS, Green, "SUCCESS", v...)
}

func (l *Log) Successf(format string, v ...interface{}) {
	l.logf(SUCCESS, Green, "SUCCESS", format, v...)
}

func (l *Log) Warn(v ...interface{}) {
	l.log(WARNING, Yellow, "WARNING", v...)
}

func (l *Log) Warnf(format string, v ...interface{}) {
	l.logf(WARNING, Yellow, "WARNING", format, v...)
}

func (l *Log) Error(v ...interface{}) {
	l.log(ERROR, Red, "ERROR", v...)
}

func (l *Log) Errorf(format string, v ...interface{}) {
	l.logf(ERROR, Red, "ERROR", format, v...)
}

func (l *Log) Fatal(v ...interface{}) {
	l.log(FATAL, Purple, "FATAL", v...)
	os.Exit(1)
}

func (l *Log) Fatalf(format string, v ...interface{}) {
	l.logf(FATAL, Purple, "FATAL", format, v...)
	os.Exit(1)
}

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

// DefaultLoggerConfig returns a LoggerConfig with default settings
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
			Warnf("failed to initialise file logger: %v", err)
		}
	}

	return func(c *Context) {
		// Start timer
		start := time.Now()
		path := c.GetURLPath()
		raw := c.GetRawQuery()

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
		methodColor := GetMethodColor(c.GetMethod())

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

func IsFileLoggingEnabled() bool {
	return defaultLogger != nil && fileLogger != nil
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
