package objects

import (
	"testing"
)

func TestParseValidId(t *testing.T) {
	have, err := ParseObjectId("sha3-256:000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	want := ObjectId{
		Algo: OIdAlgoSHA3_256,
		Sum: []byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		},
	}

	if !have.Equals(want) {
		t.Errorf("unexpected result, want: '%s', have: '%s'", want, have)
	}
}

func TestParseMalformedIds(t *testing.T) {
	malformed := []string{
		"",                // Empty not permitted
		"sha3-256",        // Missing :
		"sha3-256:",       // Missing hex sum
		":abcdef",         // Missing algo
		"foobar:abcdef",   // Basic format ok, but unknown algo
		"sha3-256:foobar", // Not hexadecimal
		"sha3-256:abcdef", // sum length mismatch
	}

	for _, s := range malformed {
		oid, err := ParseObjectId(s)
		if err == nil {
			t.Errorf("'%s' resulted in valid id (%s), expected error", s, oid)
		}
	}
}

func TestParseSerialize(t *testing.T) {
	want := ObjectId{
		Algo: OIdAlgoSHA3_256,
		Sum: []byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		},
	}

	have, err := ParseObjectId(want.String())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if !have.Equals(want) {
		t.Errorf("unexpected result, want: '%s', have: '%s'", want, have)
	}
}

func TestSerializeParse(t *testing.T) {
	want := "sha3-256:000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

	oid, err := ParseObjectId(want)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	have := oid.String()

	if have != want {
		t.Errorf("unexpected result, want: '%s', have: '%s'", want, have)
	}
}
