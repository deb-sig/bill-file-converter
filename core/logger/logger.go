package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Level string

const (
	Verbose Level = "verbose"
	Info    Level = "info"
	Warning Level = "warning"
	Error   Level = "error"
)

// NormalizeLevel maps unknown levels to Info so callers do not need to guard
// every log write.
func NormalizeLevel(level Level) Level {
	switch level {
	case Verbose, Info, Warning, Error:
		return level
	default:
		return Info
	}
}

// Logger writes short process events to a line-based log file and stores long
// debug payloads as separate files. Keeping large JSON outside the line log
// avoids editor freezes while preserving enough data to reproduce parse issues.
type Logger struct {
	taskID     string
	adapterKey string
	terminal   io.Writer
	dir        string
	path       string
	file       *os.File
	mu         sync.Mutex
}

// newBase creates the logger output directory under outputDir/logger and opens the
// line-based process log.
func newBase(outputDir string, taskID string, adapterKey string, terminal io.Writer) (*Logger, error) {
	dir := filepath.Join(outputDir, "logger")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "bill_file_converter.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	return &Logger{
		taskID:     taskID,
		adapterKey: adapterKey,
		terminal:   terminal,
		dir:        dir,
		path:       path,
		file:       file,
	}, nil
}

// Dir returns the directory used for long logger payloads.
func (l *Logger) Dir() string {
	if l == nil {
		return ""
	}
	return l.dir
}

// Path returns the path of the line-based process log.
func (l *Logger) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

// Write appends one short process event to bill_file_converter.log.
func (l *Logger) Write(level Level, message string) string {
	if l == nil || l.file == nil {
		return ""
	}
	line := formatLine(time.Now(), l.taskID, level, message)
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.file.WriteString(line + "\n")
	return line
}

func formatLine(t time.Time, taskID string, level Level, message string) string {
	level = NormalizeLevel(level)
	return fmt.Sprintf("%s [%s] [%s] %s", t.Format(time.RFC3339), strings.ToUpper(string(level)), taskID, message)
}

// SaveText writes a long text payload, such as a raw JSON response, as its own
// file under Dir.
func (l *Logger) SaveText(name string, content string) (string, error) {
	if l == nil || content == "" {
		return "", nil
	}
	path := filepath.Join(l.dir, name)
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// Close closes the process log file.
func (l *Logger) Close() {
	if l == nil || l.file == nil {
		return
	}
	_ = l.file.Close()
}
