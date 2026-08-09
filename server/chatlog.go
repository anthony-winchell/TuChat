package main

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatLog struct {
	mu   sync.Mutex
	file *os.File
	path string
}

func NewChatLog(roomName string) (*ChatLog, error) {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, err
	}

	path := "logs/" + roomName + ".log"

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_APPEND|os.O_RDWR,
		0644,
	)
	if err != nil {
		return nil, err
	}

	return &ChatLog{
		file: file,
		path: path,
	}, nil
}

func (l *ChatLog) Write(entry LogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if _, err := l.file.Write(append(data, '\n')); err != nil {
		return err
	}

	return nil
}

func (l *ChatLog) Read() ([]LogEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := l.file.Seek(0, 0); err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(l.file)

	var entries []LogEntry

	for {
		var entry LogEntry

		err := decoder.Decode(&entry)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (l *ChatLog) Recent(count int) ([]LogEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if count <= 0 {
		return []LogEntry{}, nil
	}

	if _, err := l.file.Seek(0, 0); err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(l.file)

	entries := make([]LogEntry, 0, count)

	for {
		var entry LogEntry

		err := decoder.Decode(&entry)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)

		if len(entries) > count {
			entries = entries[1:]
		}
	}

	return entries, nil
}

func (l *ChatLog) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.file.Close()
}

func (s *Server) getChatLog(roomName string) (*ChatLog, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if log, exists := s.chatLogs[roomName]; exists {
		return log, nil
	}

	log, err := NewChatLog(roomName)
	if err != nil {
		return nil, err
	}

	s.chatLogs[roomName] = log

	return log, nil
}

func (l *ChatLog) Rename(newPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.file.Close(); err != nil {
		return err
	}

	if err := os.Rename(l.path, newPath); err != nil {
		return err
	}

	file, err := os.OpenFile(
		newPath,
		os.O_CREATE|os.O_APPEND|os.O_RDWR,
		0644,
	)
	if err != nil {
		return err
	}

	l.file = file
	l.path = newPath

	return nil
}

func (l *ChatLog) Delete() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.file.Close(); err != nil {
		return err
	}

	return os.Remove(l.path)
}
