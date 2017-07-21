package objects

import (
	"bytes"
	"testing"
	"time"
)

var (
	testSnapshotObj = Snapshot{
		Archive: "foo",
		Comment: "foo\nbar\nbaz!",
		Date:    time.Date(2017, 07, 01, 21, 40, 00, 0, time.FixedZone("", 2*60*60)),
		Signed:  true,
		Tree:    genId(0xff),
	}

	testSnapshotSerialization = []byte("" +
		"== BEGIN SNAPSHOT ==\n" +
		"archive foo\n" +
		"date 2017-07-01T21:40:00+02:00\n" +
		"signed yes\n" +
		"tree sha3-256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff\n" +
		"\n" +
		"foo\n" +
		"bar\n" +
		"baz!\n" +
		"== END SNAPSHOT ==\n")
)

func TestSerializeSnapshot(t *testing.T) {
	have := testSnapshotObj.Payload()

	if !bytes.Equal(have, testSnapshotSerialization) {
		t.Errorf("Unexpected serialization result: %s", have)
	}
}

func TestUnserializeSnapshot(t *testing.T) {
	have := Snapshot{}
	if err := have.FromPayload(testSnapshotSerialization); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if !have.Equals(testSnapshotObj) {
		t.Errorf("Unexpeced unserialization result: %v", have)
	}
}

func TestUnserializeSnapshotFailure(t *testing.T) {
	subtests := []struct{ name, payload string }{
		{"empty", ""},
		{"missing tree", snapshot_start_line + "archive foo\ndate 2017-07-01T22:02:00+02:00\n" + snapshot_end_line},
		{"missing archive", snapshot_start_line + "date 2017-07-01T22:02:00+02:00\ntree sha3-256:0000000000000000000000000000000000000000000000000000000000000000\n" + snapshot_end_line},
		{"missing date", snapshot_start_line + "archive foo\ntree sha3-256:0000000000000000000000000000000000000000000000000000000000000000\n" + snapshot_end_line},
		{"invalid date", snapshot_start_line + "archive foo\ndate foobar\ntree sha3-256:0000000000000000000000000000000000000000000000000000000000000000\n" + snapshot_end_line},
		{"End marker missing", snapshot_start_line + "archive foo\ndate 2017-07-01T22:02:00+02:00\ntree sha3-256:0000000000000000000000000000000000000000000000000000000000000000\n"},
	}

	for _, subtest := range subtests {
		have := Snapshot{}
		err := have.FromPayload([]byte(subtest.payload))
		if err == nil {
			t.Errorf("Unexpected unserialization success: %v", have)
		}
	}
}
