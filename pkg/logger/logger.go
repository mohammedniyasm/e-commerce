package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type DailyWriter struct {
	mu          sync.Mutex
	currentDate string
	file        *os.File
	logDir      string
}

func NewDailyWriter(logDir string) (*DailyWriter, error) {
	writer := &DailyWriter{
		logDir: logDir,
	}
	if err := writer.rotate(); err != nil {
		return nil, err
	}
	return writer, nil
}
func (w *DailyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	if today != w.currentDate {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}
func (w *DailyWriter) rotate() error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
		w.file = nil
	}
	if err := os.MkdirAll(w.logDir, 0755); err != nil {
		return err
	}
	today := time.Now().Format("2006-01-02")
	filePath := filepath.Join(w.logDir, fmt.Sprintf("%s.log", today))
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	w.file = file
	w.currentDate = today
	return nil
}
func (w *DailyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}
func New() (*slog.Logger, io.Closer, error) {
	writer, err := NewDailyWriter("logs")
	if err != nil {
		return nil, nil, err
	}
	logger := slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	return logger, writer, nil
}
