package fs

import (
	"io"
	"time"
)

type FileType string

const (
	FFile    FileType = "file"
	FDir     FileType = "dir"
	FSymlink FileType = "symlink"
)

type File interface {
	Type() FileType // Depending on type, the File must also implement RegularFile (FFile), Dir (FDir) or Symlink (FSymlink)
	Name() string
	Executable() bool // For now we will only record the executable bit instead of all permission bits
	ModTime() time.Time
	Delete() error
}

type RegularFile interface {
	File
	Open() (io.ReadWriteCloser, error)
}

type Dir interface {
	File
	Readdir() ([]File, error)

	CreateChildFile(name string, exec bool) (RegularFile, error)
	CreateChildDir(name string) (Dir, error)
	CreateChildSymlink(name string, target string) (Symlink, error)
}

type Symlink interface {
	File
	Readlink() (string, error)
}
