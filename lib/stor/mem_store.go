// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package stor

import (
	"sync"

	"github.com/webern/flog"
	"github.com/webern/tftp/lib/cor"
)

var _ Store = (*memStore)(nil)

// memStore implements the Store interface for storing and retrieving files in a memory cache.
type memStore struct {
	mx         sync.RWMutex      // protects all data fields
	files      map[string][]byte // stores the files
	terminated bool              // when true, all functions return an error
}

// NewMemStore creates a new Store for storing and retrieving files to/from a memory cache.
func NewMemStore() Store {
	return &memStore{
		mx:         sync.RWMutex{},
		files:      make(map[string][]byte),
		terminated: false,
	}
}

// Get returns a file from the store
func (m *memStore) Get(name string) (cor.File, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if m.terminated {
		return cor.File{}, flog.Raise("the memStore has been terminated")
	}

	if b, ok := m.files[name]; ok {
		f := cor.File{}
		f.Name = name
		f.Data = make([]byte, len(b), len(b))
		copy(f.Data, b)
		return f, nil
	}

	return cor.File{}, flog.Raisef("the file '%s' was not found", name)
}

// Put places a file into the Store
func (m *memStore) Put(f cor.File) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	if m.terminated {
		return flog.Raise("the memStore has been terminated")
	}

	b := make([]byte, len(f.Data), len(f.Data))
	copy(b, f.Data)
	m.files[f.Name] = b
	return nil
}

// Terminate tells the Store it is about to be destroyed
func (m *memStore) Terminate() {
	m.mx.Lock()
	defer m.mx.Unlock()
	defer flog.Trace("terminated")
	m.terminated = true
}
