package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu   sync.Mutex
	file *os.File
}

func New(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	return &Logger{file: f}, nil
}

func (l *Logger) Close() error {
	return l.file.Close()
}

func (l *Logger) Log(level, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	ts := time.Now().UTC().Format(time.RFC3339)
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.file, "%s\t%s\t%s\n", ts, level, msg)
}

func (l *Logger) Info(format string, args ...any)  { l.Log("INFO", format, args...) }
func (l *Logger) Error(format string, args ...any) { l.Log("ERROR", format, args...) }
func (l *Logger) Status(url string, code int) {
	l.Log("STATUS", "url=%s status=%d", url, code)
}
