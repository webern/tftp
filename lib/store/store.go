// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package store

import "github.com/webern/tftp/lib/tfcore"

type Store interface {

	// Chan returns a channel that can be used to store and retrieve files to and from the Store
	RequestChan() chan<- Request

	// Terminate destroys the memory store, closes its request channel, and leaves it in an unusable state.
	Terminate()
}

type Response struct {
	Err  error
	File *tfcore.File
}

type RequestType int

const (
	ReqPut  RequestType = iota // Request to Put a file into the Store
	ReqGet                     // Request to Get a file from the Store
	ReqTerm                    // Request the termination of the Store and it's goroutine
)

type Request struct {
	Type         RequestType
	File         *tfcore.File
	ResponseChan chan<- Response
}

//// PutResponse is used to signal the completion of an atomic operation in which a File is stored in the Store
//type PutResponse struct {
//	// Err represents an error condition in which the file was not stored.
//	Err error
//}
//
//// PutRequest represents a request to store a file in the Store. A response will be sent on ResponseChan if not nil.
//type PutRequest struct {
//	// File represents the File that we are requesting be stored in the Store.
//	tfcore.File
//
//	// ResponseChan will be used to signal the completion of the put operation. It is optional and can be left nil.
//	// The caller supplies ResponseChan if the caller wishes to wait for successful completion of the operation.
//	ResponseChan chan PutResponse
//}
//
//// GetRequest represents a request sent to the Store to retrieve a file by name.
//type GetRequest struct {
//	// Name represents the file name of the file that we wish to retrieve from the Store.
//	Name string
//
//	// ResponseChan must be supplied by the caller, this is how the data will be returned to the caller from the store.
//	ResponseChan chan GetResponse
//}
//
//// GetResponse represents a response sent back from the Store for a given request.
//type GetResponse struct {
//	// Err represents any error that occurred while trying to retrieve the file. For example, file not found.
//	Err error
//
//	// File represents a deep copy of the retrieved file, unless Err is not nil.
//	File tfcore.File
//}
