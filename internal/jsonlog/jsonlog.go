package jsonlog

import (
	"io"
	"os"
	"sync"
)

// Define a Level type to represent the severity level for a log entry.
type Level int8

// Initialize constants which represent a specific severity level.
// We use the iota keyword as a shortcut to assign successive integer values to the constants
const (
	LevelInfo  Level = iota // Has the value 0
	LevelError              // Has the value 1
	LevelFatal              // Has the value 2
	LevelOff                // Has the value 3
)

// Return a human-friendly string for the severity level.
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Define a custom Logger type. This holds the output destination that the log entries.
// Will be written to, the minimum severity level that log entries will be written for,
// Plus a mutex for coordinating the writes.
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// Return a new Logger instance which writes log entries at or above a minimum severity
// level to a specific output destination.
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// Declare some helper methods for writing log entries at the different levels.
// Notice that these all accept a map as the second parameter which can contain any arbitrary.
// 'properties' that you want to appear in the log entry.
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1) // For entries at the FATAL level, we also terminate the application.
}

// Print is an internal methof for writting the log entry.
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	// If the severity level of the log entry is below the minimum severity for the Logger
	// Then return with no further action.
	if level < l.minLevel {
		return 0, nil
	}

	// TODO: Declare an anonymous struct holding the data for the log entry.
	return 1, nil
}
