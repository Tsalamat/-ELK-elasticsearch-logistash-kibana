package app

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	mu      sync.Mutex
	file    *os.File
	service string
}

type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Service string `json:"service"`
	Message string `json:"message"`
	Method  string `json:"method,omitempty"`
	Path    string `json:"path,omitempty"`
	Status  int    `json:"status,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func NewLogger(path string, service string) (*Logger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	return &Logger{file: file, service: service}, nil
}

func (l *Logger) Close() error {
	return l.file.Close()
}

func (l *Logger) Info(message string, data any) {
	l.Write(LogEntry{Level: "info", Message: message, Data: data})
}

func (l *Logger) Error(message string, data any) {
	l.Write(LogEntry{Level: "error", Message: message, Data: data})
}

func (l *Logger) Write(entry LogEntry) {
	entry.Time = time.Now().UTC().Format(time.RFC3339Nano)
	entry.Service = l.service

	line, err := json.Marshal(entry)
	if err != nil {
		log.Printf("marshal log entry: %v", err)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := l.file.Write(append(line, '\n')); err != nil {
		log.Printf("write log file: %v", err)
	}
	log.Println(string(line))
}
