package backup

import (
	"code.laria.me/petrific/cache"
	"code.laria.me/petrific/fs"
	"code.laria.me/petrific/logging"
	"code.laria.me/petrific/objects"
	"code.laria.me/petrific/storage/memory"
	"testing"
	"time"
)

func TestCacheMTime(t *testing.T) {
	c := cache.NewFileCache("") // location doesn't matter here, just use empty string
	st := memory.NewMemoryStorage()
	filesys := fs.NewMemoryFSRoot("/foo")

	file, err := filesys.CreateChildFile("bar", false)
	if err != nil {
		t.Fatal(err)
	}

	want := file.ModTime()
	if _, err := WriteDir(st, "/foo", filesys, c, logging.NewNopLog()); err != nil {
		t.Fatal(err)
	}

	have, _, ok := c.PathUpdated("/foo/bar")
	if !ok {
		t.Fatal("cache doesn't know anything about /foo/bar")
	}

	if !have.Equal(want) {
		t.Errorf("Unexpected cache time for /foo/bar (want=%s, have=%s)", want, have)
	}
}

func TestCacheRetrieve(t *testing.T) {
	c := cache.NewFileCache("") // location doesn't matter here, just use empty string
	st := memory.NewMemoryStorage()
	filesys := fs.NewMemoryFSRoot("/foo")

	if err := st.Set(objid_emptyfile, objects.OTFile, obj_emptyfile); err != nil {
		t.Fatalf("could not set empty file object: %s", err)
	}

	file, err := filesys.CreateChildFile("bar", false)
	if err != nil {
		t.Fatal(err)
	}
	mfile := file.(*fs.MemfsFile)

	mtime := file.ModTime().Add(1 * time.Hour)
	c.SetPathUpdated("/foo/bar", mtime, objid_emptyfile)

	if _, err := WriteDir(st, "/foo", filesys, c, logging.NewNopLog()); err != nil {
		t.Fatal(err)
	}

	if mfile.HasBeenRead {
		t.Error("/foo/bar has been read by WriteDir")
	}
}
