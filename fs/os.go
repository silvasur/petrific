package fs

import (
	"io"
	"os"
	"strings"
	"time"
)

func pathJoin(parts ...string) string {
	return strings.Join(parts, string(os.PathSeparator))
}

func OpenOSFile(path string) (osFile, error) {
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

func (f osFile) Open() (io.ReadCloser, error) {
	fh, err := os.Open(f.fullpath)
	if err != nil {
		return nil, err
	}
	return fh, nil
}

func (f osFile) OpenWritable() (io.WriteCloser, error) {
	fh, err := os.Create(f.fullpath)
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
			fullpath: pathJoin(f.fullpath, fi.Name()),
			fi:       fi,
		})
	}

	return
}

func (f osFile) GetChild(name string) (File, error) {
	path := pathJoin(f.fullpath, name)
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	return osFile{path, fi}, nil
}

func perms(executable bool) os.FileMode {
	if executable {
		return 0755
	} else {
		return 0644
	}
}

func (f osFile) CreateChildFile(name string, exec bool) (RegularFile, error) {
	p := pathJoin(f.fullpath, name)

	fh, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, perms(exec))
	if err != nil {
		return nil, err
	}
	fh.Close()

	return OpenOSFile(p)
}

func (f osFile) CreateChildDir(name string) (Dir, error) {
	p := pathJoin(f.fullpath, name)

	if err := os.Mkdir(p, perms(true)); err != nil {
		return nil, err
	}

	return OpenOSFile(p)
}

func (f osFile) CreateChildSymlink(name string, target string) (Symlink, error) {
	p := pathJoin(f.fullpath, name)

	err := os.Symlink(target, p)
	if err != nil {
		return nil, err
	}

	return OpenOSFile(p)
}

func (f osFile) RenameChild(oldname, newname string) error {
	return os.Rename(pathJoin(f.fullpath, oldname), pathJoin(f.fullpath, newname))
}

func (f osFile) Readlink() (string, error) {
	return os.Readlink(f.fullpath)
}
