package piper

import (
	"errors"
	"os"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// FileHandler is logging file struct
type FileHandler struct {
	file *os.File
	path string
	mode os.FileMode
}

// Error types
var (
	ErrFileHandlerIsNil = errors.New("File handler is nil")
	ErrOutputNotSet     = errors.New("Output file is not set")
)

// Global is a global file handler
var Global = &FileHandler{}

// ////////////////////////////////////////////////////////////////////////////////// //

// New creates new FileHandler struct
func New(path string, mode os.FileMode) (*FileHandler, error) {
	handler := &FileHandler{
		path: path,
		mode: mode,
	}

	err := handler.Set(path, mode)

	if err != nil {
		return nil, err
	}

	return handler, nil
}

// Set sets initial parameters for logging
func Set(path string, perms os.FileMode) error {
	return Global.Set(path, perms)
}

// Close closes logging file
func Close() error {
	return Global.Close()
}

// Write writes data to logging file
func Write(p []byte) (n int, err error) {
	return Global.Write(p)
}

// Reopen tries to reopen log file (useful for rotating)
func Reopen() error {
	return Global.Reopen()
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Set sets initial parameters for logging
func (h *FileHandler) Set(path string, perms os.FileMode) error {
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, perms)

	if err == nil {
		h.file, h.path, h.mode = fp, path, perms
	}

	return err
}

// Close closes logging file
func (h *FileHandler) Close() error {
	return h.file.Close()
}

// Write writes data to logging file
func (h *FileHandler) Write(p []byte) (n int, err error) {
	return h.file.Write(p)
}

// Reopen tries to reopen log file (useful for rotating)
func (h *FileHandler) Reopen() error {
	if h == nil {
		return ErrFileHandlerIsNil
	}

	if h.file == nil {
		return ErrOutputNotSet
	}

	h.file.Close()

	return h.Set(h.path, h.mode)
}
