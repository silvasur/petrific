package fs

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"
)

type memfsBase struct {
	parent *MemfsDir
	name   string
	exec   bool
	mtime  time.Time
}

type memfsChild interface {
	File
	setName(string)
}

func (b memfsBase) Name() string       { return b.name }
func (b *memfsBase) setName(n string)  { b.name = n }
func (b memfsBase) Executable() bool   { return b.exec }
func (b memfsBase) ModTime() time.Time { return b.mtime }

func (b memfsBase) Delete() error {
	if b.parent == nil {
		return errors.New("Root entry can not be deleted")
	}
	b.parent.deleteChild(b.name)
	return nil
}

type MemfsFile struct {
	memfsBase
	content     *bytes.Buffer
	HasBeenRead bool
}

func (MemfsFile) Type() FileType { return FFile }

func (f *MemfsFile) Open() (io.ReadCloser, error) {
	return f, nil
}

func (f MemfsFile) OpenWritable() (io.WriteCloser, error) {
	return f, nil
}

func (f *MemfsFile) Read(p []byte) (int, error) {
	f.HasBeenRead = true
	return f.content.Read(p)
}

func (f MemfsFile) Write(p []byte) (int, error) {
	return f.content.Write(p)
}

func (MemfsFile) Close() error {
	return nil
}

type MemfsDir struct {
	memfsBase
	children map[string]memfsChild
}

func (MemfsDir) Type() FileType { return FDir }

func (d MemfsDir) Readdir() ([]File, error) {
	l := make([]File, 0, len(d.children))

	for _, f := range d.children {
		l = append(l, f)
	}

	return l, nil
}

func (d MemfsDir) GetChild(name string) (File, error) {
	c, ok := d.children[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return c, nil
}

func (d MemfsDir) createChildBase(name string, exec bool) memfsBase {
	return memfsBase{
		parent: &d,
		name:   name,
		exec:   exec,
		mtime:  time.Now(),
	}
}

func (d MemfsDir) CreateChildFile(name string, exec bool) (RegularFile, error) {
	child := MemfsFile{
		memfsBase: d.createChildBase(name, exec),
		content:   new(bytes.Buffer),
	}
	d.children[name] = &child
	return &child, nil
}

func (d MemfsDir) CreateChildDir(name string) (Dir, error) {
	child := MemfsDir{
		memfsBase: d.createChildBase(name, true),
		children:  make(map[string]memfsChild),
	}
	d.children[name] = &child
	return &child, nil
}

func (d MemfsDir) CreateChildSymlink(name string, target string) (Symlink, error) {
	child := MemfsSymlink{
		memfsBase: d.createChildBase(name, false),
		target:    target,
	}
	d.children[name] = &child
	return &child, nil
}

func (d *MemfsDir) deleteChild(name string) {
	delete(d.children, name)
}

func (d *MemfsDir) RenameChild(oldname, newname string) error {
	c, ok := d.children[oldname]
	if !ok {
		return os.ErrNotExist
	}

	c.setName(newname)

	delete(d.children, oldname)
	d.children[newname] = c

	return nil
}

func NewMemoryFSRoot(name string) Dir {
	return &MemfsDir{
		memfsBase: memfsBase{
			parent: nil,
			name:   name,
			exec:   true,
			mtime:  time.Now(),
		},
		children: make(map[string]memfsChild),
	}
}

type MemfsSymlink struct {
	memfsBase
	target string
}

func (MemfsSymlink) Type() FileType { return FSymlink }

func (s MemfsSymlink) Readlink() (string, error) {
	return s.target, nil
}
