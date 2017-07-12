package backup

import (
	"code.laria.me/petrific/objects"
)

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
