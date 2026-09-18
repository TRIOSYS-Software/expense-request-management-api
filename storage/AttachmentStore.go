package storage

import (
	"os"
	"path/filepath"
)

type AttachmentStore interface {
	Remove(name string)
}

// DiskStore removes files from a base directory.
type DiskStore struct {
	dir string
}

func NewDiskStore(dir string) *DiskStore {
	return &DiskStore{dir: dir}
}

// Remove deletes one stored file if it exists.
func (s *DiskStore) Remove(name string) {
	if name == "" || filepath.Base(name) != name {
		return
	}
	path := filepath.Join(s.dir, name)
	if _, err := os.Stat(path); err != nil {
		return
	}
	_ = os.Remove(path)
}
