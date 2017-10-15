package backup

import (
	"bytes"
	"github.com/silvasur/petrific/fs"
	"github.com/silvasur/petrific/objects"
	"github.com/silvasur/petrific/storage/memory"
	"io"
	"testing"
)

func withChildOfType(t *testing.T, root fs.Dir, name string, ft fs.FileType, do func(*testing.T, fs.File)) {
	f, err := root.GetChild(name)
	if err != nil {
		t.Errorf("Could not GetChild(%s): %s", name, err)
		return
	}

	if f.Type() != ft {
		t.Errorf("Child '%s' has type %s, expected %s", name, f.Type(), ft)
		return
	}

	do(t, f)
}

func wantFileWithContent(want []byte, exec bool) func(*testing.T, fs.File) {
	return func(t *testing.T, f fs.File) {
		rf := f.(fs.RegularFile)

		if rf.Executable() != exec {
			t.Errorf("Child '%s' has executable bit %b, expected %b", f.Name(), rf.Executable(), exec)
		}

		rwc, err := rf.Open()
		if err != nil {
			t.Errorf("could not open child '%s': %s", f.Name(), err)
		}
		defer rwc.Close()

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, rwc); err != nil {
			t.Errorf("Could not read content of child '%s': %s", err)
			return
		}

		have := buf.Bytes()
		if !bytes.Equal(have, want) {
			t.Errorf("Unexpected content of child '%s': %s", f.Name(), have)
		}
	}
}

func wantDir(n int, fx func(*testing.T, fs.Dir)) func(*testing.T, fs.File) {
	return func(t *testing.T, f fs.File) {
		d := f.(fs.Dir)

		children, err := d.Readdir()
		if err != nil {
			t.Errorf("Could not Readdir() '%s': %s", f.Name(), err)
			return
		}

		if len(children) != n {
			t.Errorf("Expected '%s' to have %d children, got %d", f.Name(), n, len(children))
			return
		}

		fx(t, d)
	}
}

func TestRestoreDir(t *testing.T) {
	s := memory.NewMemoryStorage()

	s.Set(objid_emptyfile, objects.OTFile, obj_emptyfile)
	s.Set(objid_fooblob, objects.OTBlob, obj_fooblob)
	s.Set(objid_foofile, objects.OTFile, obj_foofile)
	s.Set(objid_emptytree, objects.OTTree, obj_emptytree)
	s.Set(objid_subtree, objects.OTTree, obj_subtree)
	s.Set(objid_testtree, objects.OTTree, obj_testtree)

	root := fs.NewMemoryFSRoot("")

	if err := RestoreDir(s, objid_testtree, root); err != nil {
		t.Fatalf("Unexpected error from RestoreDir(): %s", err)
	}

	wantDir(4, func(t *testing.T, root fs.Dir) {
		withChildOfType(t, root, "foo", fs.FFile, wantFileWithContent([]byte("foo"), false))
		withChildOfType(t, root, "bar", fs.FFile, wantFileWithContent([]byte{}, true))
		withChildOfType(t, root, "baz", fs.FSymlink, func(t *testing.T, f fs.File) {
			target, err := f.(fs.Symlink).Readlink()
			if err != nil {
				t.Errorf("Could not Readlink() child 'baz': %s", err)
				return
			}

			if target != "foo" {
				t.Errorf("Unexpected target for baz: %s", target)
			}
		})

		withChildOfType(t, root, "sub", fs.FDir, wantDir(2, func(t *testing.T, d fs.Dir) {
			withChildOfType(t, d, "a", fs.FFile, wantFileWithContent([]byte{}, false))
			withChildOfType(t, d, "b", fs.FDir, wantDir(0, func(t *testing.T, d fs.Dir) {}))
		}))
	})(t, root)
}

func TestRestoreLargeFile(t *testing.T) {
	s := memory.NewMemoryStorage()
	s.Set(objid_largefile_blob0, objects.OTBlob, obj_largefile_blob0)
	s.Set(objid_largefile_blob1, objects.OTBlob, obj_largefile_blob1)
	s.Set(objid_largefile, objects.OTFile, obj_largefile)

	buf := new(bytes.Buffer)

	if err := RestoreFile(s, objid_largefile, buf); err != nil {
		t.Fatalf("Unexpected error while restoring file: %s", err)
	}

	have := buf.Bytes()
	if !bytes.Equal(have, content_largefile) {
		t.Errorf("Unexpected restoration result: %s", have)
	}
}
