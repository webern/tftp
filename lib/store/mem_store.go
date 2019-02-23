// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package store

import (
	"github.com/webern/flog"
	"github.com/webern/tftp/lib/tfcore"
	"sync"
)

var _ Store = (*memStore)(nil)

// memStore implements the Store interface for storing and retrieving files in a memory cache.
type memStore struct {
	mx    sync.RWMutex // protects the files map
	files map[string][]byte
}

// NewMemStore creates a new Store for storing and retrieving files to/from a memory cache.
func NewMemStore() Store {
	return &memStore{
		mx:    sync.RWMutex{},
		files: make(map[string][]byte),
	}
}

// Get returns a file from the store
func (m *memStore) Get(name string) (tfcore.File, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()
	if b, ok := m.files[name]; ok {
		f := tfcore.File{}
		f.Name = name
		f.Data = make([]byte, len(b), len(b))
		copy(f.Data, b)
		return f, nil
	}

	return tfcore.File{}, flog.Raisef("the file '%s' was not found", name)
}

// Put places a file into the Store
func (m *memStore) Put(f tfcore.File) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	b := make([]byte, len(f.Data), len(f.Data))
	copy(b, f.Data)
	m.files[f.Name] = b
	return nil
}

// Terminate tells the Store it is about to be destroyed
func (m *memStore) Terminate() {
	// nothing to do
}
