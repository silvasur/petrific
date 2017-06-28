package objects

import (
	"bytes"
	"testing"
	"time"
)

var (
	testSnapshotObj = Snapshot{
		Comment:   "foo\nbar\nbaz!",
		Container: "foo",
		Date:      time.Date(2017, 07, 01, 21, 40, 00, 0, time.FixedZone("", 2*60*60)),
		Tree:      genId(0xff),
	}

	testSnapshotSerialization = []byte("" +
		"container foo\n" +
		"date 2017-07-01T21:40:00+02:00\n" +
		"tree sha3-256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff\n" +
		"\n" +
		"foo\n" +
		"bar\n" +
		"baz!")
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
		{"missing tree", "container foo\ndate 2017-07-01T22:02:00+02:00\n"},
		{"missing container", "date 2017-07-01T22:02:00+02:00\ntree sha3-256:0000000000000000000000000000000000000000000000000000000000000000\n"},
		{"missing date", "container foo\ntree sha3-256:0000000000000000000000000000000000000000000000000000000000000000\n"},
		{"invalid date", "container foo\ndate foobar\ntree sha3-256:0000000000000000000000000000000000000000000000000000000000000000\n"},
	}

	for _, subtest := range subtests {
		have := Snapshot{}
		err := have.FromPayload([]byte(subtest.payload))
		if err == nil {
			t.Errorf("Unexpected unserialization success: %v", have)
		}
	}
}
