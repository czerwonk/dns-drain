// SPDX-FileCopyrightText: (c) 2016 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package changelog

import (
	"encoding/json"
	"os"
)

type FileChangeLog struct {
	filename string
}

func NewFileChangeLog(file string) *FileChangeLog {
	return &FileChangeLog{filename: file}
}

func (l *FileChangeLog) GetChanges() (*DnsChangeSet, error) {
	b, err := os.ReadFile(l.filename)
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
