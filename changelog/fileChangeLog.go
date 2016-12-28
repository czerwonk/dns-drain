package changelog

import (
	"encoding/json"
	"io/ioutil"
)

type FileChangeLog struct {
	filename string
}

func NewFileChangeLog(file string) *FileChangeLog {
	return &FileChangeLog{filename: file}
}

func (l *FileChangeLog) GetChanges() (*DnsChangeSet, error) {
	b, err := ioutil.ReadFile(l.filename)
	if err != nil {
		return nil, err
	}

	c := &DnsChangeSet{}
	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
