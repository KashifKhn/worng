// Package vfs provides a virtual filesystem abstraction.
// The real OS filesystem is used by the CLI; an in-memory FS is used by tests and the WASM playground.
package vfs

import (
	"os"
)

// FS is the filesystem interface used throughout WORNG.
type FS interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	Exists(path string) bool
}

// OsFS is the real OS-backed filesystem.
type OsFS struct{}

func (OsFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (OsFS) WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

func (OsFS) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// MemFS is an in-memory filesystem for tests and WASM.
type MemFS struct {
	files map[string][]byte
}

func NewMemFS() *MemFS {
	return &MemFS{files: make(map[string][]byte)}
}

func (m *MemFS) ReadFile(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}
	return data, nil
}

func (m *MemFS) WriteFile(path string, data []byte) error {
	m.files[path] = data
	return nil
}

func (m *MemFS) Exists(path string) bool {
	_, ok := m.files[path]
	return ok
}
