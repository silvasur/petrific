package acl

import (
	"os"
	"testing"
)

func checkPerm(t *testing.T, cat rune, perms QualifiedPerms, k string, want Perm) {
	have, ok := perms[k]
	if ok {
		if have != want {
			t.Errorf("Perm %c:%s:%s != %c:%s:%s", cat, k, have, cat, k, want)
		}
	} else {
		t.Errorf("Perm %c:%s: not set", cat, k)
	}
}

func TestFromUnix(t *testing.T) {
	acl := ACLFromUnixPerms(0752)
	if len(acl.User) != 1 {
		t.Errorf("Expected exactly 1 user perm, got %d", len(acl.User))
	}
	if len(acl.Group) != 1 {
		t.Errorf("Expected exactly 1 group perm, got %d", len(acl.Group))
	}
	if len(acl.Other) != 1 {
		t.Errorf("Expected exactly 1 other perm, got %d", len(acl.Other))
	}
	if len(acl.Mask) != 0 {
		t.Errorf("Expected exactly 0 mask perm, got %d", len(acl.Mask))
	}

	checkPerm(t, 'u', acl.User, "", PermR|PermW|PermX)
	checkPerm(t, 'g', acl.Group, "", PermR|PermX)
	checkPerm(t, 'o', acl.Other, "", PermW)
}

func TestToUnix(t *testing.T) {
	acl := ACL{}
	acl.Init()

	acl.User[""] = PermR | PermW | PermX
	acl.Group[""] = PermR | PermX
	acl.Other[""] = PermR

	want := os.FileMode(0754)
	have := acl.ToUnixPerms()
	if have != want {
		t.Errorf("Unexpected ToUnixPerms result: have 0%o, want 0%o", have, want)
	}
}

func TestStringify(t *testing.T) {
	acl := ACL{}
	acl.Init()

	acl.User[""] = PermR | PermW | PermX
	acl.User["foo"] = PermR | PermW
	acl.User["bar"] = PermR | PermW
	acl.Group[""] = PermR | PermX
	acl.Group["baz"] = PermR | PermW | PermX
	acl.Other[""] = 0
	acl.Mask[""] = PermX

	want := "u::rwx,u:bar:rw-,u:foo:rw-,g::r-x,g:baz:rwx,o::---,m::--x"
	have := acl.String()
	if have != want {
		t.Errorf("Unexpected String result: have %s, want %s", have, want)
	}
}

func TestParsing(t *testing.T) {
	acl, err := ParseACL("u::rwx,u:bar:rw-,u:foo:rw-,g::r-x,g:baz:rwx,o::---,m::--x")
	if err != nil {
		t.Fatalf("unexpected parsing error: %s", err)
	}

	if len(acl.User) == 3 {
		checkPerm(t, 'u', acl.User, "", PermR|PermW|PermX)
		checkPerm(t, 'u', acl.User, "foo", PermR|PermW)
		checkPerm(t, 'u', acl.User, "bar", PermR|PermW)
	} else {
		t.Errorf("Expeced 3 user entries, got %d", len(acl.User))
	}

	if len(acl.Group) == 2 {
		checkPerm(t, 'u', acl.Group, "", PermR|PermX)
		checkPerm(t, 'u', acl.Group, "baz", PermR|PermW|PermX)
	} else {
		t.Errorf("Expeced 2 group entries, got %d", len(acl.Group))
	}

	if len(acl.Other) == 1 {
		checkPerm(t, 'u', acl.Other, "", 0)
	} else {
		t.Errorf("Expeced 1 other entries, got %d", len(acl.Other))
	}

	if len(acl.Mask) == 1 {
		checkPerm(t, 'u', acl.Mask, "", PermX)
	} else {
		t.Errorf("Expeced 1 mask entries, got %d", len(acl.Mask))
	}
}
