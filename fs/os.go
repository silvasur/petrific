package fs

import (
	"io"
	"os"
	"time"
)

func openOSFile(path string) (osFile, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return osFile{}, err
	}

	return osFile{
		fullpath: path,
		fi:       fi,
	}, nil
}

type osFile struct {
	fullpath string
	fi       os.FileInfo
}

func (f osFile) Type() FileType {
	m := f.fi.Mode()
	if m.IsDir() {
		return FDir
	}
	if m.IsRegular() {
		return FFile
	}
	if m&os.ModeSymlink != 0 {
		return FSymlink
	}
	return "unknown"
}

func (f osFile) Name() string {
	return f.fi.Name()
}

func (f osFile) Executable() bool {
	return f.fi.Mode()&0100 != 0 // x bit set for user?
}

func (f osFile) ModTime() time.Time {
	return f.fi.ModTime()
}

func (f osFile) Delete() error {
	return os.RemoveAll(f.fullpath)
}

func (f osFile) Open() (io.ReadWriteCloser, error) {
	fh, err := os.Open(f.fullpath)
	if err != nil {
		return nil, err
	}
	return fh, nil
}

func (f osFile) Readdir() (list []File, err error) {
	fh, err := os.Open(f.fullpath)
	if err != nil {
		return
	}
	defer fh.Close()

	infos, err := fh.Readdir(-1)
	if err != nil {
		return
	}

	for _, fi := range infos {
		if fi.Name() == "." || fi.Name() == ".." {
			continue
		}

		list = append(list, osFile{
			fullpath: f.fullpath + string(os.PathSeparator) + fi.Name(),
			fi:       fi,
		})
	}

	return
}

func perms(executable bool) os.FileMode {
	if executable {
		return 0755
	} else {
		return 0644
	}
}

func (f osFile) CreateChildFile(name string, exec bool) (RegularFile, error) {
	p := f.fullpath + string(os.PathSeparator) + name

	fh, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, perms(exec))
	if err != nil {
		return nil, err
	}
	fh.Close()

	return openOSFile(p)
}

func (f osFile) CreateChildDir(name string) (Dir, error) {
	p := f.fullpath + string(os.PathSeparator) + name

	if err := os.Mkdir(p, perms(true)); err != nil {
		return nil, err
	}

	return openOSFile(p)
}

func (f osFile) CreateChildSymlink(name string, target string) (Symlink, error) {
	p := f.fullpath + string(os.PathSeparator) + name

	err := os.Symlink(target, p)
	if err != nil {
		return nil, err
	}

	return openOSFile(p)
}

func (f osFile) Readlink() (string, error) {
	return os.Readlink(f.fullpath)
}
