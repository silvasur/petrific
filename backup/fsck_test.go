package backup

import (
	"code.laria.me/petrific/logging"
	"code.laria.me/petrific/objects"
	"code.laria.me/petrific/storage/memory"
	"testing"
)

func TestHealthy(t *testing.T) {
	st := memory.NewMemoryStorage()
	st.Set(objid_emptyfile, objects.OTFile, obj_emptyfile)
	st.Set(objid_fooblob, objects.OTBlob, obj_fooblob)
	st.Set(objid_foofile, objects.OTFile, obj_foofile)
	st.Set(objid_emptytree, objects.OTTree, obj_emptytree)
	st.Set(objid_subtree, objects.OTTree, obj_subtree)
	st.Set(objid_testtree, objects.OTTree, obj_testtree)

	problems := make(chan FsckProblem)
	var err error
	go func() {
		err = Fsck(st, nil, true, problems, logging.NewNopLog())
		close(problems)
	}()

	for p := range problems {
		t.Errorf("Unexpected problem: %s", p)
	}

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

var (
	// snapshot with missing tree object
	obj_corrupt_snapshot_1 = []byte("" +
		"snapshot 162\n" +
		"== BEGIN SNAPSHOT ==\n" +
		"archive foo\n" +
		"date 2018-01-06T22:42:00+01:00\n" +
		"tree sha3-256:f000000000000000000000000000000000000000000000000000000000000000\n" +
		"== END SNAPSHOT ==\n")
	objid_corrupt_snapshot_1 = objects.MustParseObjectId("sha3-256:e33ad8ed4ef309099d593d249b36f2a5377dd26aeb18479695763fec514f519e")

	missing_objid_corrupt_snapshot_1 = objects.MustParseObjectId("sha3-256:f000000000000000000000000000000000000000000000000000000000000000")

	obj_corrupt_snapshot_2 = []byte("" +
		"snapshot 162\n" +
		"== BEGIN SNAPSHOT ==\n" +
		"archive foo\n" +
		"date 2018-01-06T22:45:00+01:00\n" +
		"tree sha3-256:086f877d9e0760929c0a528ca3a01a7a19c03176a132cc6f4894c69b5943d543\n" +
		"== END SNAPSHOT ==\n")
	objid_corrupt_snapshot_2 = objects.MustParseObjectId("sha3-256:d5da78d96bb1bc7bff1f7cee509dba084b54ff4b9af0ed23a6a14437765ac81f")

	obj_corrupt_tree = []byte("tree 431\n" +
		"name=invalidhash&ref=sha3-256:8888888888888888888888888888888888888888888888888888888888888888&type=file\n" +
		"name=invalidserialization&ref=sha3-256:7c3c1c331531a80d0e37a6066a6a4e4881eb897f1d76aeffd86a3bd96f0c054f&type=file\n" +
		"name=lengthmismatch&ref=sha3-256:caea41322f4e02d68a15abe3a867c9ab507674a1f70abc892a162c5b3a742349&type=file\n" +
		"name=missingobj&ref=sha3-256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff&type=file\n")
	objid_corrupt_tree = objects.MustParseObjectId("sha3-256:086f877d9e0760929c0a528ca3a01a7a19c03176a132cc6f4894c69b5943d543")

	invalid_objid = objects.MustParseObjectId("sha3-256:8888888888888888888888888888888888888888888888888888888888888888")

	missing_objid_corrupt_tree = objects.MustParseObjectId("sha3-256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	obj_corrupt_file_invalidhash   = obj_emptyfile
	objid_corrupt_file_invalidhash = objects.MustParseObjectId("sha3-256:8888888888888888888888888888888888888888888888888888888888888888")

	obj_corrupt_file_invalidserialization   = []byte("file 9\nsize=123\n")
	objid_corrupt_file_invalidserialization = objects.MustParseObjectId("sha3-256:7c3c1c331531a80d0e37a6066a6a4e4881eb897f1d76aeffd86a3bd96f0c054f")

	obj_corrupt_blob_lengthmismatch   = []byte("blob 2\nx\n")
	objid_corrupt_blob_lengthmismatch = objects.MustParseObjectId("sha3-256:c9f04ca8fb21c7abb6221060b4e2a332686d0f6be872bdeb85cdc5fe3f2743ca")

	obj_corrupt_file_lengthmismatch = []byte("" +
		"file 88\n" +
		"blob=sha3-256:c9f04ca8fb21c7abb6221060b4e2a332686d0f6be872bdeb85cdc5fe3f2743ca&size=100\n")
	objid_corrupt_file_lengthmismatch = objects.MustParseObjectId("sha3-256:caea41322f4e02d68a15abe3a867c9ab507674a1f70abc892a162c5b3a742349")
)

func TestCorrupted(t *testing.T) {
	st := memory.NewMemoryStorage()
	st.Set(objid_corrupt_snapshot_1, objects.OTSnapshot, obj_corrupt_snapshot_1)
	st.Set(objid_corrupt_snapshot_2, objects.OTSnapshot, obj_corrupt_snapshot_2)
	st.Set(objid_corrupt_tree, objects.OTTree, obj_corrupt_tree)
	st.Set(objid_corrupt_file_invalidhash, objects.OTFile, obj_corrupt_file_invalidhash)
	st.Set(objid_corrupt_file_invalidserialization, objects.OTFile, obj_corrupt_file_invalidserialization)
	st.Set(objid_corrupt_blob_lengthmismatch, objects.OTBlob, obj_corrupt_blob_lengthmismatch)
	st.Set(objid_corrupt_file_lengthmismatch, objects.OTFile, obj_corrupt_file_lengthmismatch)

	problems := make(chan FsckProblem)
	var err error
	go func() {
		err = Fsck(st, nil, true, problems, logging.NewNopLog())
		close(problems)
	}()

	var seen_snapshot_1_problem,
		seen_invalidhash,
		seen_invalidserialization,
		seen_lengthmismatch,
		seen_missing_file bool

	for p := range problems {
		if p.Id.Equals(missing_objid_corrupt_snapshot_1) && p.ProblemType == FsckStorageError {
			seen_snapshot_1_problem = true
			continue
		}

		if p.Id.Equals(invalid_objid) && p.ProblemType == FsckStorageError {
			seen_invalidhash = true
			continue
		}

		if p.Id.Equals(objid_corrupt_file_invalidserialization) && p.ProblemType == FsckDeserializationError {
			seen_invalidserialization = true
			continue
		}

		if p.Id.Equals(objid_corrupt_blob_lengthmismatch) &&
			p.ProblemType == FsckUnexpectedBlobSize &&
			p.HaveSize == 2 && p.WantSize == 100 {

			seen_lengthmismatch = true
			continue
		}

		if p.Id.Equals(missing_objid_corrupt_tree) && p.ProblemType == FsckStorageError {
			seen_missing_file = true
			continue
		}

		t.Errorf("Unexpected problem: %s", p)
	}

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !seen_snapshot_1_problem {
		t.Error("missing problem snapshot_1_problem")
	}
	if !seen_invalidhash {
		t.Error("missing problem invalidhash")
	}
	if !seen_invalidserialization {
		t.Error("missing problem invalidserialization")
	}
	if !seen_lengthmismatch {
		t.Error("missing problem lengthmismatch")
	}
	if !seen_missing_file {
		t.Error("missing problem missing_file")
	}
}
