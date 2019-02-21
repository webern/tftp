// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package store

import (
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/tfcore"
)

var _ Store = (*memStore)(nil)

// memStore implements the Store interface for storing and retrieving files.
type memStore struct {
	ch    chan Request
	files map[string][]byte
}

// NewMemStore creates an in-memory file store which listens for requests on a goroutine. The goroutine is created by
// the NewMemStore function. To destoy the goroutine call memStore.Terminate()
func NewMemStore() Store {
	m := memStore{
		ch:    make(chan Request),
		files: make(map[string][]byte, 10),
	}

	go m.listen()
	return &m
}

func (m *memStore) RequestChan() chan<- Request {
	return m.ch
}

func (m *memStore) Terminate() {
	resp := make(chan Response)
	termReq := Request{Type: ReqTerm, File: nil, ResponseChan: resp}
	go func() { m.ch <- termReq }()
	//<-resp
	close(m.ch)
	return
}

func (m *memStore) listen() {
	select {
	case req := <-m.ch:
		{
			if req.Type == ReqPut {
				err := m.put(req.File)
				if req.ResponseChan != nil {
					go func() { req.ResponseChan <- Response{err, nil} }()
				}
			} else if req.Type == ReqGet {
				file, err := m.get(req.File)
				if req.ResponseChan != nil {
					go func() { req.ResponseChan <- Response{err, &file} }()
				}
			} else if req.Type == ReqTerm {
				m.files = nil
				if req.ResponseChan != nil {
					go func() { req.ResponseChan <- Response{nil, nil} }()
				}
				return
			}
		}
	}
}

func (m *memStore) put(file *tfcore.File) error {
	if file == nil {
		return flog.Raise("file is nil")
	}

	b := make([]byte, len(file.Data), len(file.Data))
	copy(b, file.Data)
	m.files[file.Name] = b
	return nil
}

func (m *memStore) get(file *tfcore.File) (tfcore.File, error) {
	if file == nil {
		return tfcore.File{}, flog.Raise("file is nil, you need to pass the filename using file.Name")
	}

	b, ok := m.files[file.Name]
	file.Data = make([]byte, len(b), len(b))
	copy(file.Data, b)

	if !ok {
		return tfcore.File{}, flog.Raisef("the file '%s' was not found", file.Name)
	}

	return *file, nil
}
