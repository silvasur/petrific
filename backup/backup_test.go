package backup

import (
	"bytes"
	"code.laria.me/petrific/cache"
	"code.laria.me/petrific/fs"
	"code.laria.me/petrific/objects"
	"code.laria.me/petrific/storage"
	"testing"
)

func wantObject(
	t *testing.T,
	s storage.Storage,
	id_str string,
	want []byte,
) {
	id, err := objects.ParseObjectId(id_str)
	if err != nil {
		t.Errorf("Could not parse id: %s", err)
		return
	}

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
	s := storage.NewMemoryStorage()

	id, err := WriteFile(s, bytes.NewReader(make([]byte, 2*BlobChunkSize+100)))
	if err != nil {
		t.Fatalf("Unexpected error when writing file: %s", err)
	}

	if id.String() != "sha3-256:ab7907ee6b45b343422a0354de500bcf99f5ff69fe8125be84e43d421803c34e" {
		t.Errorf("Unexpected file id: %s", id)
	}

	want_large_blob := append([]byte("blob 16777216\n"), make([]byte, BlobChunkSize)...)
	want_small_blob := append([]byte("blob 100\n"), make([]byte, 100)...)
	want_file := []byte("file 274\n" +
		"blob=sha3-256:7287cbb09bdd8a0d96a6f6297413cd9d09a2763814636245a5a44120e6351be3&size=16777216\n" +
		"blob=sha3-256:7287cbb09bdd8a0d96a6f6297413cd9d09a2763814636245a5a44120e6351be3&size=16777216\n" +
		"blob=sha3-256:ddf124464f7b80e95f4a9c704f79e7037ff5d731648ba6b40c769893b428128c&size=100\n")

	wantObject(t, s, "sha3-256:ab7907ee6b45b343422a0354de500bcf99f5ff69fe8125be84e43d421803c34e", want_file)
	wantObject(t, s, "sha3-256:7287cbb09bdd8a0d96a6f6297413cd9d09a2763814636245a5a44120e6351be3", want_large_blob)
	wantObject(t, s, "sha3-256:ddf124464f7b80e95f4a9c704f79e7037ff5d731648ba6b40c769893b428128c", want_small_blob)
}

func mkfile(t *testing.T, d fs.Dir, name string, exec bool, content []byte) {
	f, err := d.CreateChildFile(name, exec)
	if err != nil {
		t.Fatalf("Could not create file %s: %s", name, err)
	}

	rwc, err := f.Open()
	if err != nil {
		t.Fatalf("Could not create file %s: %s", name, err)
	}
	defer rwc.Close()

	if _, err := rwc.Write(content); err != nil {
		t.Fatalf("Could not create file %s: %s", name, err)
	}
}

func TestWriteDir(t *testing.T) {
	s := storage.NewMemoryStorage()

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

	//4a10682307d5b5dc072d1b862497296640176109347b149aad38cd640000491b
	obj_emptyfile := []byte("file 0\n")

	//ba632076629ff33238850c870fcb51e4b7b67b3d9dcb66314adbcf1770a5fea7
	obj_fooblob := []byte("blob 3\nfoo")
	//fa50ca1fc739852528ecc149b424a8ccbdf84b73c8718cde4525f2a410d79244
	obj_foofile := []byte("file 86\nblob=sha3-256:ba632076629ff33238850c870fcb51e4b7b67b3d9dcb66314adbcf1770a5fea7&size=3\n")

	//1dc6fae780ae4a1e823a5b8e26266356a2e1d22e5904b0652dcff6e3c0e72067
	obj_emptytree := []byte("tree 0\n")

	//f1716a1b0cad23b6faab9712243402b8f8e7919c377fc5d5d87bd465cef056d7
	obj_subdir := []byte("tree 239\n" +
		"acl=u::rw-,g::r--,o::r--&name=a&ref=sha3-256:4a10682307d5b5dc072d1b862497296640176109347b149aad38cd640000491b&type=file\n" +
		"acl=u::rwx,g::r-x,o::r-x&name=b&ref=sha3-256:1dc6fae780ae4a1e823a5b8e26266356a2e1d22e5904b0652dcff6e3c0e72067&type=dir\n")

	//09e881f57befa1eacec744e3857a36f0d9d5dd1fa72ba96564b467a3d7d0c0d5
	obj_dir := []byte("tree 423\n" +
		"acl=u::rw-,g::r--,o::r--&name=baz&target=foo&type=symlink\n" +
		"acl=u::rw-,g::r--,o::r--&name=foo&ref=sha3-256:fa50ca1fc739852528ecc149b424a8ccbdf84b73c8718cde4525f2a410d79244&type=file\n" +
		"acl=u::rwx,g::r-x,o::r-x&name=bar&ref=sha3-256:4a10682307d5b5dc072d1b862497296640176109347b149aad38cd640000491b&type=file\n" +
		"acl=u::rwx,g::r-x,o::r-x&name=sub&ref=sha3-256:f1716a1b0cad23b6faab9712243402b8f8e7919c377fc5d5d87bd465cef056d7&type=dir\n")

	wantObject(t, s, "sha3-256:4a10682307d5b5dc072d1b862497296640176109347b149aad38cd640000491b", obj_emptyfile)
	wantObject(t, s, "sha3-256:ba632076629ff33238850c870fcb51e4b7b67b3d9dcb66314adbcf1770a5fea7", obj_fooblob)
	wantObject(t, s, "sha3-256:fa50ca1fc739852528ecc149b424a8ccbdf84b73c8718cde4525f2a410d79244", obj_foofile)
	wantObject(t, s, "sha3-256:1dc6fae780ae4a1e823a5b8e26266356a2e1d22e5904b0652dcff6e3c0e72067", obj_emptytree)
	wantObject(t, s, "sha3-256:f1716a1b0cad23b6faab9712243402b8f8e7919c377fc5d5d87bd465cef056d7", obj_subdir) //!
	wantObject(t, s, "sha3-256:09e881f57befa1eacec744e3857a36f0d9d5dd1fa72ba96564b467a3d7d0c0d5", obj_dir)    //!
}
