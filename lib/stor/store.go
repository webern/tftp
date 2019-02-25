// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package stor

import "github.com/webern/tftp/lib/cor"

// Store represents a mechanism for storing and retrieving files by name
type Store interface {
	// Put stores a file in the Store (overwriting if a file by the same name exists). Set is safe for concurrent
	// goroutine access. The file is deep copied before storing.
	Put(f cor.File) error

	// Get returns a file from the Store or an error if it is not found. Get is safe for concurrent goroutine access.
	// The returned file is a deep copy of the stored file, you may mutate it without affecting the Store.
	Get(name string) (cor.File, error)

	// Terminate informs the Store that it is about to be destroyed, giving the Store time to finish any operations.
	// Terminate blocks until such operations are complete.
	Terminate()
}
