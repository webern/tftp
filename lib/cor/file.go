// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package cor

// File represents a file that will be transferred by TFTP
type File struct {
	Name string
	Data []byte
}
