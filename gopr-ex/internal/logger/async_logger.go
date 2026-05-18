package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AsyncLogger struct {
	events chan string
	file   *os.File
	wg     sync.WaitGroup
	once   sync.Once
}

func NewAsyncLogger(logFile string, buffer int) (*AsyncLogger, error) {
	if logFile == "" {
		logFile = "log.txt"
	}
	if dir := filepath.Dir(logFile); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create log dir: %w", err)
		}
	}
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	if buffer <= 0 {
		buffer = 100
	}
	l := &AsyncLogger{
		events: make(chan string, buffer),
		file:   file,
	}
	l.wg.Add(1)
	go l.worker()
	return l, nil
}

func (l *AsyncLogger) LogEvent(event string) {
	select {
	case l.events <- event:
	default:
		log.Printf("event log channel is full, dropped event: %s", event)
	}
}

func (l *AsyncLogger) Stop(ctx context.Context) error {
	l.once.Do(func() {
		close(l.events)
	})
	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return l.file.Close()
	case <-ctx.Done():
		_ = l.file.Close()
		return ctx.Err()
	}
}

func (l *AsyncLogger) worker() {
	defer l.wg.Done()
	for event := range l.events {
		select {
		case <-time.After(time.Second):
		}
		line := fmt.Sprintf("%s %s\n", time.Now().UTC().Format(time.RFC3339), event)
		log.Print(stringsTrimNewline(line))
		if _, err := l.file.WriteString(line); err != nil {
			log.Printf("write event log: %v", err)
		}
	}
}

func stringsTrimNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
