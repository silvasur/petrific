package objects

import (
	"bytes"
	"code.laria.me/petrific/acl"
	"testing"
)

var (
	testTreeObj = Tree{
		"foo": TreeEntryFile{
			TreeEntryBase: TreeEntryBase{
				acl:   acl.ACLFromUnixPerms(0644),
				user:  "user1",
				group: "group1",
			},
			Ref: genId(0x11),
		},
		"bar": TreeEntryDir{
			TreeEntryBase: TreeEntryBase{
				acl:   acl.ACLFromUnixPerms(0755),
				user:  "user2",
				group: "group2",
			},
			Ref: genId(0x22),
		},
		"baz": TreeEntrySymlink{
			TreeEntryBase: TreeEntryBase{
				acl:   acl.ACLFromUnixPerms(0644),
				user:  "user3",
				group: "group3",
			},
			Target: "/föö&bär/💾",
		}, // Test special chars and unicode
		"😃": TreeEntryFile{
			TreeEntryBase: TreeEntryBase{
				acl: acl.ACL{
					User: acl.QualifiedPerms{
						"":      acl.PermR | acl.PermW,
						"user1": acl.PermR | acl.PermW,
					},
					Group: acl.QualifiedPerms{
						"": acl.PermR,
					},
					Other: acl.QualifiedPerms{
						"": acl.PermR,
					},
					Mask: acl.QualifiedPerms{
						"": acl.PermR | acl.PermW,
					},
				},
			},
			Ref: genId(0x33),
		},
	}

	testTreeSerialization = []byte("" +
		"acl=u::rw-,g::r--,o::r--&group=group1&name=foo&ref=sha3-256:1111111111111111111111111111111111111111111111111111111111111111&type=file&user=user1\n" +
		"acl=u::rw-,g::r--,o::r--&group=group3&name=baz&target=%2ff%c3%b6%c3%b6%26b%c3%a4r%2f%f0%9f%92%be&type=symlink&user=user3\n" +
		"acl=u::rw-,u:user1:rw-,g::r--,o::r--,m::rw-&name=%f0%9f%98%83&ref=sha3-256:3333333333333333333333333333333333333333333333333333333333333333&type=file\n" +
		"acl=u::rwx,g::r-x,o::r-x&group=group2&name=bar&ref=sha3-256:2222222222222222222222222222222222222222222222222222222222222222&type=dir&user=user2\n")
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
