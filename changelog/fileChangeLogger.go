package changelog

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

type FileChangeLogger struct {
	mutex   sync.Mutex
	file    *os.File
	changes []DnsChange
}

func NewFileChangeLogger(filePath string) (*FileChangeLogger, error) {
	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	return &FileChangeLogger{file: f, changes: make([]DnsChange, 0)}, nil
}

func (l *FileChangeLogger) LogChange(c DnsChange) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.changes = append(l.changes, c)

	return nil
}

func (l *FileChangeLogger) Flush() error {
	c := DnsChangeSet{Changes: l.changes}

	b, err := json.Marshal(c)

	if err != nil {
		return err
	}

	w := bufio.NewWriter(l.file)
	_, err = w.Write(b)
	if err != nil {
		return err
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (l *FileChangeLogger) Close() error {
	return l.file.Close()
}
