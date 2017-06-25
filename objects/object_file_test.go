package objects

import (
	"bytes"
	"testing"
)

var (
	testFileObj = File{
		FileFragment{Blob: genId(0x11), Size: 10},
		FileFragment{Blob: genId(0x22), Size: 20},
		FileFragment{Blob: genId(0x33), Size: 30},
		FileFragment{Blob: genId(0x44), Size: 40},
		FileFragment{Blob: genId(0x55), Size: 50},
	}

	testFileSerialization = []byte("" +
		"blob=sha3-256:1111111111111111111111111111111111111111111111111111111111111111&size=10\n" +
		"blob=sha3-256:2222222222222222222222222222222222222222222222222222222222222222&size=20\n" +
		"blob=sha3-256:3333333333333333333333333333333333333333333333333333333333333333&size=30\n" +
		"blob=sha3-256:4444444444444444444444444444444444444444444444444444444444444444&size=40\n" +
		"blob=sha3-256:5555555555555555555555555555555555555555555555555555555555555555&size=50\n")
)

func genId(b byte) (oid ObjectId) {
	oid.Algo = OIdAlgoSHA3_256
	oid.Sum = make([]byte, OIdAlgoSHA3_256.sumLength())
	for i := 0; i < OIdAlgoSHA3_256.sumLength(); i++ {
		oid.Sum[i] = b
	}

	return
}

func TestSerializeFile(t *testing.T) {
	have := testFileObj.Payload()

	if !bytes.Equal(have, testFileSerialization) {
		t.Errorf("Unexpected serialization result: %s", have)
	}
}

func TestSerializeEmptyFile(t *testing.T) {
	f := File{}

	have := f.Payload()
	want := []byte{}

	if !bytes.Equal(have, want) {
		t.Errorf("Unexpected serialization result: %s", have)
	}
}

func TestUnserializeFile(t *testing.T) {
	have := File{}
	err := have.FromPayload(testFileSerialization)

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if !have.Equals(testFileObj) {
		t.Errorf("Unexpeced unserialization result: %v", have)
	}
}

func TestUnserializeEmptyFile(t *testing.T) {
	have := File{}
	err := have.FromPayload([]byte{})

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if len(have) != 0 {
		t.Errorf("Unexpeced unserialization result: %v", have)
	}
}

func TestUnserializeFailure(t *testing.T) {
	subtests := []struct{ name, payload string }{
		{"missing blob", "size=100\n"},
		{"empty blob", "blob=&size=100"},
		{"invalid blob", "blob=foobar&size=100"}, // Variations of invalid IDs are tested elsewhere
		{"missing size", "blob=sha3-256:0000000000000000000000000000000000000000000000000000000000000000\n"},
		{"empty size", "blob=sha3-256:0000000000000000000000000000000000000000000000000000000000000000&size=\n"},
		{"invalid size", "blob=sha3-256:0000000000000000000000000000000000000000000000000000000000000000&size=foobar\n"},
		{"no props", "foobar\n"},
	}

	for _, subtest := range subtests {
		have := File{}
		err := have.FromPayload([]byte(subtest.payload))
		if err == nil {
			t.Errorf("Unexpected unserialization success: %v", have)
		}
	}
}
