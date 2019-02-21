// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package store

import "github.com/webern/tftp/lib/tfcore"

type Store interface {
	// Put stores a file in the Store (overwriting if a file by the same name exists). Set is safe for concurrent
	// goroutine access. The file is deep copied before storing.
	Put(f tfcore.File) error

	// Get returns a file from the Store or an error if it is not found. Get is safe for concurrent goroutine access.
	// The returned file is a deep copy of the stored file, you may mutate it without affecting the Store.
	Get(name string) (tfcore.File, error)
}
