package changelog

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

type ChangesJson struct {
	Changes []ChangeJson `json:"changes"`
}

type ChangeJson struct {
	Provider   string `json:"provider"`
	Action     string `json:"action"`
	Zone       string `json:"zone"`
	Record     string `json:"record"`
	RecordType string `json:"recordType"`
	Value      string `json:"value"`
}

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
	j := l.getJson()
	b, err := json.Marshal(j)

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

func (l *FileChangeLogger) getJson() ChangesJson {
	changes := make([]ChangeJson, 0)

	for _, x := range l.changes {
		c := ChangeJson{Provider: x.Provider, Action: x.Action, Zone: x.Zone, Record: x.Record, RecordType: x.RecordType, Value: x.Value}
		changes = append(changes, c)
	}

	return ChangesJson{Changes: changes}
}

func (l *FileChangeLogger) Close() error {
	return l.file.Close()
}
