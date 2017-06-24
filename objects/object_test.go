package objects

import (
	"bufio"
	"bytes"
	"testing"
)

func TestUnserializeSuccess(t *testing.T) {
	r := bufio.NewReader(bytes.NewBuffer([]byte("blob 16\n0123456789abcdef")))

	o, err := UnserializeObject(r)
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

	o, err := UnserializeObject(r)
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

	o, err := UnserializeObject(r)
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
	o := Object{
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
