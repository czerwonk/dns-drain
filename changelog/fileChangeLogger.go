package changelog

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type FileChangeLogger struct {
	mutex  sync.Mutex
	file   *os.File
	writer *bufio.Writer
}

func NewFileChangeLogger(filePath string) (*FileChangeLogger, error) {
	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	w := bufio.NewWriter(f)
	return &FileChangeLogger{file: f, writer: w}, nil
}

func (l *FileChangeLogger) LogChange(provider string, c DnsChange) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	_, err := fmt.Fprintf(l.writer, "%s\t%s\t%s\t%s\t%s\t%s\n", c.Action, provider, c.Zone, c.Record, c.RecordType, c.Value)
	return err
}

func (l *FileChangeLogger) Close() error {
	err := l.writer.Flush()
	if err != nil {
		return err
	}

	return l.file.Close()
}
