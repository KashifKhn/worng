package vfs

import (
	"errors"
	"os"
	"testing"
)

// ─── MemFS ────────────────────────────────────────────────────────────────────

func TestMemFSWriteAndRead(t *testing.T) {
	t.Parallel()

	m := NewMemFS()
	data := []byte("hello worng")

	if err := m.WriteFile("test.wrg", data); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	got, err := m.ReadFile("test.wrg")
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("ReadFile = %q, want %q", got, data)
	}
}

func TestMemFSReadMissingFile(t *testing.T) {
	t.Parallel()

	m := NewMemFS()
	_, err := m.ReadFile("does_not_exist.wrg")
	if err == nil {
		t.Fatalf("ReadFile on missing file: expected error, got nil")
	}

	// Must wrap os.ErrNotExist so callers can use errors.Is
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ReadFile error = %v, want to wrap os.ErrNotExist", err)
	}
}

func TestMemFSReadMissingFileIsPathError(t *testing.T) {
	t.Parallel()

	m := NewMemFS()
	_, err := m.ReadFile("missing.wrg")

	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("expected *os.PathError, got %T: %v", err, err)
	}
	if pathErr.Op != "open" {
		t.Fatalf("PathError.Op = %q, want %q", pathErr.Op, "open")
	}
	if pathErr.Path != "missing.wrg" {
		t.Fatalf("PathError.Path = %q, want %q", pathErr.Path, "missing.wrg")
	}
}

func TestMemFSExistsTrue(t *testing.T) {
	t.Parallel()

	m := NewMemFS()
	_ = m.WriteFile("exists.wrg", []byte("content"))

	if !m.Exists("exists.wrg") {
		t.Fatalf("Exists(%q) = false, want true", "exists.wrg")
	}
}

func TestMemFSExistsFalse(t *testing.T) {
	t.Parallel()

	m := NewMemFS()

	if m.Exists("ghost.wrg") {
		t.Fatalf("Exists(%q) = true, want false on empty FS", "ghost.wrg")
	}
}

func TestMemFSOverwriteFile(t *testing.T) {
	t.Parallel()

	m := NewMemFS()
	_ = m.WriteFile("f.wrg", []byte("original"))
	_ = m.WriteFile("f.wrg", []byte("updated"))

	got, err := m.ReadFile("f.wrg")
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(got) != "updated" {
		t.Fatalf("ReadFile = %q, want %q", got, "updated")
	}
}

func TestMemFSWriteEmptyContent(t *testing.T) {
	t.Parallel()

	m := NewMemFS()
	_ = m.WriteFile("empty.wrg", []byte{})

	got, err := m.ReadFile("empty.wrg")
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("ReadFile = %q, want empty", got)
	}
	if !m.Exists("empty.wrg") {
		t.Fatalf("Exists = false, want true after writing empty file")
	}
}

func TestMemFSMultipleFiles(t *testing.T) {
	t.Parallel()

	m := NewMemFS()
	files := map[string]string{
		"a.wrg": "aaa",
		"b.wrg": "bbb",
		"c.wrg": "ccc",
	}
	for name, content := range files {
		_ = m.WriteFile(name, []byte(content))
	}

	for name, content := range files {
		got, err := m.ReadFile(name)
		if err != nil {
			t.Fatalf("ReadFile(%q) error: %v", name, err)
		}
		if string(got) != content {
			t.Fatalf("ReadFile(%q) = %q, want %q", name, got, content)
		}
	}
}

func TestMemFSImplementsFSInterface(t *testing.T) {
	t.Parallel()

	var _ FS = NewMemFS()
}

// ─── OsFS ─────────────────────────────────────────────────────────────────────

func TestOsFSReadWriteAndExists(t *testing.T) {
	t.Parallel()

	fs := OsFS{}
	tmp := t.TempDir() + "/worng_test.wrg"
	data := []byte("// input ~\"hello\"")

	if err := fs.WriteFile(tmp, data); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	if !fs.Exists(tmp) {
		t.Fatalf("Exists = false, want true after write")
	}

	got, err := fs.ReadFile(tmp)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("ReadFile = %q, want %q", got, data)
	}
}

func TestOsFSReadMissingFile(t *testing.T) {
	t.Parallel()

	fs := OsFS{}
	_, err := fs.ReadFile(t.TempDir() + "/no_such_file.wrg")
	if err == nil {
		t.Fatalf("ReadFile on missing file: expected error, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ReadFile error = %v, want os.ErrNotExist", err)
	}
}

func TestOsFSExistsFalse(t *testing.T) {
	t.Parallel()

	fs := OsFS{}
	if fs.Exists(t.TempDir() + "/phantom.wrg") {
		t.Fatalf("Exists = true, want false for non-existent file")
	}
}

func TestOsFSImplementsFSInterface(t *testing.T) {
	t.Parallel()

	var _ FS = OsFS{}
}
