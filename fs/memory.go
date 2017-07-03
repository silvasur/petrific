package fs

import (
	"bytes"
	"errors"
	"io"
	"time"
)

type memfsBase struct {
	parent *memfsDir
	name   string
	exec   bool
	mtime  time.Time
}

func (b memfsBase) Name() string       { return b.name }
func (b memfsBase) Executable() bool   { return b.exec }
func (b memfsBase) ModTime() time.Time { return b.mtime }

func (b memfsBase) Delete() error {
	if b.parent == nil {
		return errors.New("Root entry can not be deleted")
	}
	b.parent.deleteChild(b.name)
	return nil
}

type memfsFile struct {
	memfsBase
	content *bytes.Buffer
}

func (memfsFile) Type() FileType { return FFile }

func (f memfsFile) Open() (io.ReadWriteCloser, error) {
	return f, nil
}

func (f memfsFile) Read(p []byte) (int, error) {
	return f.content.Read(p)
}

func (f memfsFile) Write(p []byte) (int, error) {
	return f.content.Write(p)
}

func (memfsBase) Close() error {
	return nil
}

type memfsDir struct {
	memfsBase
	children map[string]File
}

func (memfsDir) Type() FileType { return FDir }

func (d memfsDir) Readdir() ([]File, error) {
	l := make([]File, 0, len(d.children))

	for _, f := range d.children {
		l = append(l, f)
	}

	return l, nil
}

func (d memfsDir) createChildBase(name string, exec bool) memfsBase {
	return memfsBase{
		parent: &d,
		name:   name,
		exec:   exec,
		mtime:  time.Now(),
	}
}

func (d memfsDir) CreateChildFile(name string, exec bool) (RegularFile, error) {
	child := memfsFile{
		memfsBase: d.createChildBase(name, exec),
		content:   new(bytes.Buffer),
	}
	d.children[name] = child
	return child, nil
}

func (d memfsDir) CreateChildDir(name string) (Dir, error) {
	child := memfsDir{
		memfsBase: d.createChildBase(name, true),
		children:  make(map[string]File),
	}
	d.children[name] = child
	return child, nil
}

func (d memfsDir) CreateChildSymlink(name string, target string) (Symlink, error) {
	child := memfsSymlink{
		memfsBase: d.createChildBase(name, false),
		target:    target,
	}
	d.children[name] = child
	return child, nil
}

func (d *memfsDir) deleteChild(name string) {
	delete(d.children, name)
}

func NewMemoryFSRoot(name string) Dir {
	return memfsDir{
		memfsBase: memfsBase{
			parent: nil,
			name:   name,
			exec:   true,
			mtime:  time.Now(),
		},
		children: make(map[string]File),
	}
}

type memfsSymlink struct {
	memfsBase
	target string
}

func (memfsSymlink) Type() FileType { return FSymlink }

func (s memfsSymlink) Readlink() (string, error) {
	return s.target, nil
}
