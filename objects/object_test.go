package objects

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func TestUnserializeSuccess(t *testing.T) {
	r := bufio.NewReader(bytes.NewBuffer([]byte("blob 16\n0123456789abcdef")))

	o, err := Unserialize(r)
	if err != nil {
		t.Fatalf("Unserialize failed: %s", err)
	}

	if o.Type != OTBlob {
		t.Errorf("expected type %s, got %s", OTBlob, o.Type)
	}

	if !bytes.Equal([]byte("0123456789abcdef"), o.Payload) {
		t.Errorf("Unexpected payload: (size %d) %v", len(o.Payload), o.Payload)
	}
}

func TestUnserializeSuccess0Payload(t *testing.T) {
	r := bufio.NewReader(bytes.NewBuffer([]byte("blob 0\n")))

	o, err := Unserialize(r)
	if err != nil {
		t.Fatalf("Unserialize failed: %s", err)
	}

	if o.Type != OTBlob {
		t.Errorf("expected type %s, got %s", OTBlob, o.Type)
	}

	if !bytes.Equal([]byte{}, o.Payload) {
		t.Errorf("Unexpected payload: (size %d) %v", len(o.Payload), o.Payload)
	}
}

func unserializeMustFail(b []byte, t *testing.T) {
	r := bufio.NewReader(bytes.NewBuffer(b))

	o, err := Unserialize(r)
	if err == nil {
		t.Fatalf("Expected an error, but object was successfully read: %#v", o)
	}

	_, ok := err.(UnserializeError)
	if !ok {
		t.Fatalf("Unknown error, expected an UnserializeError: %#v", err)
	}
}

func TestUnserializeInvalidEmpty(t *testing.T) {
	unserializeMustFail([]byte{}, t)
}

func TestUnserializeIncompleteHeader(t *testing.T) {
	unserializeMustFail([]byte("foo\nbar"), t)
}

func TestUnserializeInvalidNumber(t *testing.T) {
	unserializeMustFail([]byte("blob abc\nbar"), t)
}

func TestUnserializePayloadTooSmall(t *testing.T) {
	unserializeMustFail([]byte("blob 10\nbar"), t)
}

func TestSerialize(t *testing.T) {
	o := RawObject{
		Type:    OTBlob,
		Payload: []byte("foo bar\nbaz"),
	}

	buf := new(bytes.Buffer)
	if err := o.Serialize(buf); err != nil {
		t.Fatalf("Serialization failed:%s", err)
	}

	b := buf.Bytes()
	if !bytes.Equal(b, []byte("blob 11\nfoo bar\nbaz")) {
		t.Errorf("Unexpected serialization result: %v", b)
	}
}

func TestSerializeAndId(t *testing.T) {
	o := RawObject{
		Type:    OTBlob,
		Payload: []byte("foo bar\nbaz"),
	}

	buf := new(bytes.Buffer)
	id, err := o.SerializeAndId(buf, OIdAlgoDefault)

	if err != nil {
		t.Fatalf("Serialization failed:%s", err)
	}

	if !id.VerifyObject(o) {
		t.Errorf("Verification failed")
	}
}

func unserializeObj(t *testing.T, ot ObjectType, b []byte) Object {
	ro, err := Unserialize(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("Failed unserializing: %s", err)
	}

	o, err := ro.Object()
	if err != nil {
		t.Fatalf("Failed generating Object from RawObject: %s", err)
	}

	if o.Type() != ot {
		t.Fatalf("Unexpected object type, have %s, want %s", o.Type(), ot)
	}

	return o
}

func TestUnserializeBlobObj(t *testing.T) {
	blob := unserializeObj(t, OTBlob, []byte(""+
		"blob 6\n"+
		"foobar")).(*Blob)

	if string(*blob) != "foobar" {
		t.Errorf("Unexpected blob content: %v", blob)
	}
}

func TestUnserializeFileObj(t *testing.T) {
	file := unserializeObj(t, OTFile, append(
		[]byte(fmt.Sprintf("file %d\n", len(testFileSerialization))),
		testFileSerialization...,
	)).(*File)

	if !file.Equals(testFileObj) {
		t.Errorf("Unexpected file object: %v", *file)
	}
}

func TestUnserializeTreeObj(t *testing.T) {
	tree := unserializeObj(t, OTTree, append(
		[]byte(fmt.Sprintf("tree %d\n", len(testTreeSerialization))),
		testTreeSerialization...,
	)).(Tree)

	if !tree.Equals(testTreeObj) {
		t.Errorf("Unexpected tree object: %v", tree)
	}
}

func TestUnserializeSnapshotObj(t *testing.T) {
	snapshot := unserializeObj(t, OTSnapshot, append(
		[]byte(fmt.Sprintf("snapshot %d\n", len(testSnapshotSerialization))),
		testSnapshotSerialization...,
	)).(*Snapshot)

	if !snapshot.Equals(testSnapshotObj) {
		t.Errorf("Unexpected snapshot object: %v", *snapshot)
	}
}
