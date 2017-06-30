package acl

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type Perm uint

const (
	PermR = 4
	PermW = 2
	PermX = 1
)

func (p Perm) String() string {
	s := []byte{'r', 'w', 'x'}
	if p&PermR == 0 {
		s[0] = '-'
	}
	if p&PermW == 0 {
		s[1] = '-'
	}
	if p&PermX == 0 {
		s[2] = '-'
	}
	return string(s)
}

type QualifiedPerms map[string]Perm

func (a QualifiedPerms) Equals(b QualifiedPerms) bool {
	if len(a) != len(b) {
		return false
	}

	for k, va := range a {
		if vb, ok := b[k]; !ok || vb != va {
			return false
		}
	}

	return true
}

type ACL struct {
	User, Group, Other, Mask QualifiedPerms
}

func (acl *ACL) Init() {
	acl.User = make(QualifiedPerms)
	acl.Group = make(QualifiedPerms)
	acl.Other = make(QualifiedPerms)
	acl.Mask = make(QualifiedPerms)
}

func ACLFromUnixPerms(perms os.FileMode) (acl ACL) {
	perms &= os.ModePerm

	acl.User = QualifiedPerms{"": Perm(7 & (perms >> 6))}
	acl.Group = QualifiedPerms{"": Perm(7 & (perms >> 3))}
	acl.Other = QualifiedPerms{"": Perm(7 & perms)}

	return acl
}

func (acl ACL) ToUnixPerms() (m os.FileMode) {
	if p, ok := acl.User[""]; ok {
		m |= os.FileMode(p << 6)
	}
	if p, ok := acl.Group[""]; ok {
		m |= os.FileMode(p << 3)
	}
	if p, ok := acl.Other[""]; ok {
		m |= os.FileMode(p)
	}
	return
}

func stringifyAclCategory(parts []string, t string, perms QualifiedPerms) []string {
	keys := make([]string, 0, len(perms))
	for k := range perms {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, t+":"+k+":"+perms[k].String())
	}
	return parts
}

func (acl ACL) String() string {
	parts := []string{}
	parts = stringifyAclCategory(parts, "u", acl.User)
	parts = stringifyAclCategory(parts, "g", acl.Group)
	parts = stringifyAclCategory(parts, "o", acl.Other)
	parts = stringifyAclCategory(parts, "m", acl.Mask)

	return strings.Join(parts, ",")
}

// ParseACL parses a POSIX ACL. Only the short text form is supported
func ParseACL(s string) (ACL, error) {
	acl := ACL{}
	acl.Init()

	entries := strings.Split(s, ",")
	for i, entry := range entries {
		entry = strings.TrimSpace(entry)

		parts := strings.Split(entry, ":")
		if len(parts) != 3 {
			return acl, fmt.Errorf("invalid acl: Entry #%d has more or less than 3 ':' separated parts", i+1)
		}

		var perms *QualifiedPerms = nil

		switch parts[0] {
		case "u", "user":
			perms = &acl.User
		case "g", "group":
			perms = &acl.Group
		case "o", "other":
			perms = &acl.Other
		case "m", "mask":
			perms = &acl.Mask
		default:
			return acl, fmt.Errorf("invalid acl: Entry #%d has unknown tag \"%s\"", i+1, parts[0])
		}

		perm := Perm(0)
		if strings.Contains(parts[2], "r") {
			perm |= PermR
		}
		if strings.Contains(parts[2], "w") {
			perm |= PermW
		}
		if strings.Contains(parts[2], "x") {
			perm |= PermX
		}
		(*perms)[parts[1]] = perm
	}

	return acl, nil
}

func (a ACL) Equals(b ACL) bool {
	return a.User.Equals(b.User) &&
		a.Group.Equals(b.Group) &&
		a.Other.Equals(b.Other) &&
		a.Mask.Equals(b.Mask)
}
