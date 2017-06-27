package objects

import (
	"bytes"
	"testing"
)

var (
	testTreeObj = Tree{
		"foo": TreeEntryFile(genId(0x11)),
		"bar": TreeEntryDir(genId(0x22)),
		"baz": TreeEntrySymlink("/föö&bär/💾"), // Test special chars and unicode
		"😃":   TreeEntryFile(genId(0x33)),
	}

	testTreeSerialization = []byte("" +
		"name=%f0%9f%98%83&ref=sha3-256:3333333333333333333333333333333333333333333333333333333333333333&type=file\n" +
		"name=bar&ref=sha3-256:2222222222222222222222222222222222222222222222222222222222222222&type=dir\n" +
		"name=baz&target=%2ff%c3%b6%c3%b6%26b%c3%a4r%2f%f0%9f%92%be&type=symlink\n" +
		"name=foo&ref=sha3-256:1111111111111111111111111111111111111111111111111111111111111111&type=file\n")
)

func TestSerializeTree(t *testing.T) {
	have := testTreeObj.Payload()

	if !bytes.Equal(have, testTreeSerialization) {
		t.Errorf("Unexpected serialization result: %s", have)
	}
}

func TestSerializeEmptyTree(t *testing.T) {
	have := Tree{}.Payload()

	if !bytes.Equal(have, []byte{}) {
		t.Errorf("Unexpected serialization result: %s", have)
	}
}

func TestUnserializeTree(t *testing.T) {
	have := make(Tree)
	if err := have.FromPayload(testTreeSerialization); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if !have.Equals(testTreeObj) {
		t.Errorf("Unexpeced unserialization result: %v", have)
	}
}

func TestUnserializeTreeFailure(t *testing.T) {
	subtests := []struct{ name, payload string }{
		{"name missing", "ref=sha3-256:0000000000000000000000000000000000000000000000000000000000000000&type=file\n"},
		{"type missing", "name=foo\n"},
		{"unknown type", "name=baz&type=foobar\n"},
		{"file ref missing", "name=foo&type=file\n"},
		{"dir ref missing", "name=foo&type=dir\n"},
		{"symlink target missing", "name=foo&type=symlink\n"},
	}

	for _, subtest := range subtests {
		have := make(Tree)
		err := have.FromPayload([]byte(subtest.payload))
		if err == nil {
			t.Errorf("Unexpected unserialization success: %v", have)
		}
	}
}
