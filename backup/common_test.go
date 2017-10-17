package backup

import (
	"code.laria.me/petrific/objects"
)

// Test tree
var (
	objid_emptyfile = objects.MustParseObjectId("sha3-256:4a10682307d5b5dc072d1b862497296640176109347b149aad38cd640000491b")
	obj_emptyfile   = []byte("file 0\n")

	objid_fooblob = objects.MustParseObjectId("sha3-256:ba632076629ff33238850c870fcb51e4b7b67b3d9dcb66314adbcf1770a5fea7")
	obj_fooblob   = []byte("blob 3\nfoo")
	objid_foofile = objects.MustParseObjectId("sha3-256:fa50ca1fc739852528ecc149b424a8ccbdf84b73c8718cde4525f2a410d79244")
	obj_foofile   = []byte("file 86\nblob=sha3-256:ba632076629ff33238850c870fcb51e4b7b67b3d9dcb66314adbcf1770a5fea7&size=3\n")

	objid_emptytree = objects.MustParseObjectId("sha3-256:1dc6fae780ae4a1e823a5b8e26266356a2e1d22e5904b0652dcff6e3c0e72067")
	obj_emptytree   = []byte("tree 0\n")

	objid_subtree = objects.MustParseObjectId("sha3-256:f1716a1b0cad23b6faab9712243402b8f8e7919c377fc5d5d87bd465cef056d7")
	obj_subtree   = []byte("tree 239\n" +
		"acl=u::rw-,g::r--,o::r--&name=a&ref=sha3-256:4a10682307d5b5dc072d1b862497296640176109347b149aad38cd640000491b&type=file\n" +
		"acl=u::rwx,g::r-x,o::r-x&name=b&ref=sha3-256:1dc6fae780ae4a1e823a5b8e26266356a2e1d22e5904b0652dcff6e3c0e72067&type=dir\n")

	objid_testtree = objects.MustParseObjectId("sha3-256:09e881f57befa1eacec744e3857a36f0d9d5dd1fa72ba96564b467a3d7d0c0d5")
	obj_testtree   = []byte("tree 423\n" +
		"acl=u::rw-,g::r--,o::r--&name=baz&target=foo&type=symlink\n" +
		"acl=u::rw-,g::r--,o::r--&name=foo&ref=sha3-256:fa50ca1fc739852528ecc149b424a8ccbdf84b73c8718cde4525f2a410d79244&type=file\n" +
		"acl=u::rwx,g::r-x,o::r-x&name=bar&ref=sha3-256:4a10682307d5b5dc072d1b862497296640176109347b149aad38cd640000491b&type=file\n" +
		"acl=u::rwx,g::r-x,o::r-x&name=sub&ref=sha3-256:f1716a1b0cad23b6faab9712243402b8f8e7919c377fc5d5d87bd465cef056d7&type=dir\n")
)

// Large file
var (
	content_largefile = make([]byte, 2*BlobChunkSize+100)

	objid_largefile_blob0 = objects.MustParseObjectId("sha3-256:7287cbb09bdd8a0d96a6f6297413cd9d09a2763814636245a5a44120e6351be3")
	obj_largefile_blob0   = append([]byte("blob 16777216\n"), make([]byte, BlobChunkSize)...)

	objid_largefile_blob1 = objects.MustParseObjectId("sha3-256:ddf124464f7b80e95f4a9c704f79e7037ff5d731648ba6b40c769893b428128c")
	obj_largefile_blob1   = append([]byte("blob 100\n"), make([]byte, 100)...)

	objid_largefile = objects.MustParseObjectId("sha3-256:ab7907ee6b45b343422a0354de500bcf99f5ff69fe8125be84e43d421803c34e")
	obj_largefile   = []byte("file 274\n" +
		"blob=sha3-256:7287cbb09bdd8a0d96a6f6297413cd9d09a2763814636245a5a44120e6351be3&size=16777216\n" +
		"blob=sha3-256:7287cbb09bdd8a0d96a6f6297413cd9d09a2763814636245a5a44120e6351be3&size=16777216\n" +
		"blob=sha3-256:ddf124464f7b80e95f4a9c704f79e7037ff5d731648ba6b40c769893b428128c&size=100\n")
)
