package backup

import (
	"bytes"
	"code.laria.me/petrific/cache"
	"code.laria.me/petrific/fs"
	"code.laria.me/petrific/objects"
	"code.laria.me/petrific/storage"
	"code.laria.me/petrific/storage/memory"
	"testing"
)

func wantObject(
	t *testing.T,
	s storage.Storage,
	id objects.ObjectId,
	want []byte,
) {
	have, err := s.Get(id)
	if err != nil {
		t.Errorf("Could not get %s: %s", id, err)
		return
	}

	if !bytes.Equal(want, have) {
		t.Errorf("Wrong result for %s: (size=%d) %#s", id, len(have), have)
	}
}

func TestWriteLargeFile(t *testing.T) {
	s := memory.NewMemoryStorage()

	id, err := WriteFile(s, bytes.NewReader(content_largefile))
	if err != nil {
		t.Fatalf("Unexpected error when writing file: %s", err)
	}

	if !id.Equals(objid_largefile) {
		t.Errorf("Unexpected file id: %s", id)
	}

	wantObject(t, s, objid_largefile, obj_largefile)
	wantObject(t, s, objid_largefile_blob0, obj_largefile_blob0)
	wantObject(t, s, objid_largefile_blob1, obj_largefile_blob1)
}

func mkfile(t *testing.T, d fs.Dir, name string, exec bool, content []byte) {
	f, err := d.CreateChildFile(name, exec)
	if err != nil {
		t.Fatalf("Could not create file %s: %s", name, err)
	}

	wc, err := f.OpenWritable()
	if err != nil {
		t.Fatalf("Could not create file %s: %s", name, err)
	}
	defer wc.Close()

	if _, err := wc.Write(content); err != nil {
		t.Fatalf("Could not create file %s: %s", name, err)
	}
}

func TestWriteDir(t *testing.T) {
	s := memory.NewMemoryStorage()

	root := fs.NewMemoryFSRoot("root")

	mkfile(t, root, "foo", false, []byte("foo"))
	mkfile(t, root, "bar", true, []byte(""))
	if _, err := root.CreateChildSymlink("baz", "foo"); err != nil {
		t.Fatalf("Failed creating symlink baz: %s", err)
	}
	d, err := root.CreateChildDir("sub")
	if err != nil {
		t.Fatalf("Failed creating dir: %s", err)
	}
	mkfile(t, d, "a", false, []byte(""))
	if _, err = d.CreateChildDir("b"); err != nil {
		t.Fatalf("Failed creating dir: %s", err)
	}

	id, err := WriteDir(s, "", root, cache.NopCache{})
	if err != nil {
		t.Fatalf("Could not WriteDir: %s", err)
	}

	if id.String() != "sha3-256:09e881f57befa1eacec744e3857a36f0d9d5dd1fa72ba96564b467a3d7d0c0d5" {
		t.Errorf("Unexpected dir id: %s", id)
	}

	wantObject(t, s, objid_emptyfile, obj_emptyfile)
	wantObject(t, s, objid_fooblob, obj_fooblob)
	wantObject(t, s, objid_foofile, obj_foofile)
	wantObject(t, s, objid_emptytree, obj_emptytree)
	wantObject(t, s, objid_subtree, obj_subtree)
	wantObject(t, s, objid_testtree, obj_testtree)
}
