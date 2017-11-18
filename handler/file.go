package handler

import (
	"errors"
	"os"
	"time"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// FileHandler is logging file struct
type FileHandler struct {
	fd       *os.File    // File descriptor of opened file
	path     string      // Path to file in filesystem
	mode     os.FileMode // File permission mode
	size     uint64      // Size in bytes
	etime    int64       // Expiration time in seconds
	duration int64       // Expiration interval
}

// Error types
var (
	ErrFileHandlerIsNil = errors.New("file handler is nil")
	ErrOutputNotSet     = errors.New("output file is not set")
)

// Global is a global file handler
var Global = &FileHandler{}

// ////////////////////////////////////////////////////////////////////////////////// //

// New creates new FileHandler struct
func New(path string, mode os.FileMode) (*FileHandler, error) {
	h := &FileHandler{
		path: path,
		mode: mode,
		size: 0,
	}

	err := h.Set(path, mode, 0)

	if err != nil {
		return nil, err
	}

	return h, nil
}

// Path returns path
func Path() string {
	return Global.Path()
}

// Size returns calculated file size
func Size() uint64 {
	return Global.Size()
}

// ExpirationTime returns modified time of the file
//func ExpirationTime() int64 {
//	return Global.ExpirationTime()
//}

// Set sets initial parameters for logging
func Set(path string, perms os.FileMode, duration int64) error {
	return Global.Set(path, perms, duration)
}

// Close closes logging file
func Close() error {
	return Global.Close()
}

// Write writes data to logging file
func Write(p []byte) (n int, err error) {
	return Global.Write(p)
}

//
func Truncate() error {
	return Global.Truncate()
}

// Reopen tries to reopen log file (useful for rotating)
func Reopen() error {
	return Global.Reopen()
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Set sets initial parameters for logging
func (h *FileHandler) Set(path string, perms os.FileMode, duration int64) error {
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, perms)

	if err == nil {
		h.fd, h.path, h.mode, h.size, h.etime, h.duration = fp, path, perms, 0, 0, duration

		if duration > 0 {
			h.etime = (int64(time.Now().Unix()/duration) + 1) * duration
		}

		stat, err := fp.Stat()

		if err == nil {
			h.size = uint64(stat.Size())
		}
	}

	return err
}

// Path
func (h *FileHandler) Path() string {
	return h.path
}

// Size returns calculated file size
func (h *FileHandler) Size() uint64 {
	return h.size
}

// ExpirationTime returns modified time of the file
func (h *FileHandler) ExpirationTime() int64 {
	return h.etime
}

// Close closes logging file
func (h *FileHandler) Close() error {
	return h.fd.Close()
}

// Write writes data to logging file
func (h *FileHandler) Write(p []byte) (int, error) {
	if h == nil {
		return 0, ErrFileHandlerIsNil
	}

	if h.fd == nil {
		return 0, ErrOutputNotSet
	}

	n, err := h.fd.Write(p)

	if err == nil {
		h.size += uint64(len(p))
	}

	return n, err
}

//
func (h *FileHandler) Truncate() error {
	if h == nil {
		return ErrFileHandlerIsNil
	}

	if h.fd == nil {
		return ErrOutputNotSet
	}

	// TODO: truncate file
	return os.Remove(h.Path())
}

// Reopen tries to reopen log file (useful for rotating)
func (h *FileHandler) Reopen() error {
	if h == nil {
		return ErrFileHandlerIsNil
	}

	if h.fd == nil {
		return ErrOutputNotSet
	}

	h.fd.Close()

	return h.Set(h.path, h.mode, h.duration)
}
