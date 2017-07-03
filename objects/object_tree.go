package objects

import (
	"bufio"
	"bytes"
	"code.laria.me/petrific/acl"
	"errors"
	"fmt"
	"sort"
)

type TreeEntryType string

const (
	TETFile    TreeEntryType = "file"
	TETDir     TreeEntryType = "dir"
	TETSymlink TreeEntryType = "symlink"
)

var treeEntryTypes = map[TreeEntryType]struct{}{
	TETFile:    {},
	TETDir:     {},
	TETSymlink: {},
}

type TreeEntry interface {
	Type() TreeEntryType
	ACL() acl.ACL
	User() string
	Group() string
	equalContent(TreeEntry) bool
	toProperties() properties
}

func compareTreeEntries(a, b TreeEntry) bool {
	return a.Type() == b.Type() && a.equalContent(b)
}

type TreeEntryBase struct {
	acl         acl.ACL
	user, group string
}

func baseFromExec(exec bool) (base TreeEntryBase) {
	if exec {
		base.acl = acl.ACLFromUnixPerms(0755)
	} else {
		base.acl = acl.ACLFromUnixPerms(0644)
	}
	return
}

func (teb TreeEntryBase) ACL() acl.ACL {
	return teb.acl
}

func (teb TreeEntryBase) User() string {
	return teb.user
}

func (teb TreeEntryBase) Group() string {
	return teb.group
}

func (teb TreeEntryBase) toProperties() properties {
	props := properties{"acl": teb.acl.String()}
	if teb.user != "" {
		props["user"] = teb.user
	}
	if teb.group != "" {
		props["group"] = teb.group
	}
	return props
}

func (a TreeEntryBase) equalContent(b TreeEntryBase) bool {
	return a.acl.Equals(b.acl) && a.user == b.user && a.group == b.group
}

type TreeEntryFile struct {
	TreeEntryBase
	Ref ObjectId
}

func NewTreeEntryFile(ref ObjectId, exec bool) TreeEntryFile {
	return TreeEntryFile{
		TreeEntryBase: baseFromExec(exec),
		Ref:           ref,
	}
}

func (tef TreeEntryFile) Type() TreeEntryType {
	return TETFile
}

func (tef TreeEntryFile) toProperties() properties {
	props := tef.TreeEntryBase.toProperties()
	props["ref"] = tef.Ref.String()
	return props
}

func (a TreeEntryFile) equalContent(_b TreeEntry) bool {
	b, ok := _b.(TreeEntryFile)
	return ok && a.TreeEntryBase.equalContent(b.TreeEntryBase) && a.Ref.Equals(b.Ref)
}

type TreeEntryDir struct {
	TreeEntryBase
	Ref ObjectId
}

func NewTreeEntryDir(ref ObjectId, exec bool) TreeEntryDir {
	return TreeEntryDir{
		TreeEntryBase: baseFromExec(exec),
		Ref:           ref,
	}
}

func (ted TreeEntryDir) Type() TreeEntryType {
	return TETDir
}

func (ted TreeEntryDir) toProperties() properties {
	props := ted.TreeEntryBase.toProperties()
	props["ref"] = ted.Ref.String()
	return props
}

func (a TreeEntryDir) equalContent(_b TreeEntry) bool {
	b, ok := _b.(TreeEntryDir)
	return ok && a.TreeEntryBase.equalContent(b.TreeEntryBase) && a.Ref.Equals(b.Ref)
}

type TreeEntrySymlink struct {
	TreeEntryBase
	Target string
}

func NewTreeEntrySymlink(target string, exec bool) TreeEntrySymlink {
	return TreeEntrySymlink{
		TreeEntryBase: baseFromExec(exec),
		Target:        target,
	}
}

func (tes TreeEntrySymlink) Type() TreeEntryType {
	return TETSymlink
}

func (tes TreeEntrySymlink) toProperties() properties {
	props := tes.TreeEntryBase.toProperties()
	props["target"] = tes.Target
	return props
}

func (a TreeEntrySymlink) equalContent(_b TreeEntry) bool {
	b, ok := _b.(TreeEntrySymlink)
	return ok && a.TreeEntryBase.equalContent(b.TreeEntryBase) && a.Target == b.Target
}

type Tree map[string]TreeEntry

func (t Tree) Type() ObjectType {
	return OTTree
}

func (t Tree) Payload() (out []byte) {
	lines := []string{}
	for name, entry := range t {
		props := entry.toProperties()
		props["type"] = string(entry.Type())
		props["name"] = name

		line, err := props.MarshalText()
		if err != nil {
			panic(err)
		}

		lines = append(lines, string(line)+"\n")
	}

	sort.Strings(lines)

	for _, line := range lines {
		out = append(out, line...)
	}

	return
}

func getObjectIdFromProps(p properties, key string) (ObjectId, error) {
	raw, ok := p[key]
	if !ok {
		return ObjectId{}, fmt.Errorf("Missing key: %s", key)
	}

	oid, err := ParseObjectId(raw)
	return oid, err
}

func defaultFileTreeEntryBase(_acl *acl.ACL, props properties) (base TreeEntryBase) {
	base.user = props["user"]
	base.group = props["group"]
	if _acl == nil {
		base.acl = acl.ACLFromUnixPerms(0664)
	} else {
		base.acl = *_acl
	}
	return
}

func defaultDirTreeEntryBase(_acl *acl.ACL, props properties) (base TreeEntryBase) {
	base.user = props["user"]
	base.group = props["group"]
	if _acl == nil {
		base.acl = acl.ACLFromUnixPerms(0775)
	} else {
		base.acl = *_acl
	}
	return
}

func (t Tree) FromPayload(payload []byte) error {
	sc := bufio.NewScanner(bytes.NewReader(payload))

	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}

		props := make(properties)
		if err := props.UnmarshalText(line); err != nil {
			return err
		}

		var _acl *acl.ACL
		if acl_string, ok := props["acl"]; ok {
			acltmp, err := acl.ParseACL(acl_string)
			if err != nil {
				return err
			}
			_acl = &acltmp
		}

		entry_type, ok := props["type"]
		if !ok {
			return errors.New("Missing property: type")
		}

		var entry TreeEntry
		switch TreeEntryType(entry_type) {
		case TETFile:
			ref, err := getObjectIdFromProps(props, "ref")
			if err != nil {
				return err
			}
			entry = TreeEntryFile{
				TreeEntryBase: defaultFileTreeEntryBase(_acl, props),
				Ref:           ref,
			}
		case TETDir:
			ref, err := getObjectIdFromProps(props, "ref")
			if err != nil {
				return err
			}
			entry = TreeEntryDir{
				TreeEntryBase: defaultDirTreeEntryBase(_acl, props),
				Ref:           ref,
			}
		case TETSymlink:
			target, ok := props["target"]
			if !ok {
				return errors.New("Missing key: target")
			}
			entry = TreeEntrySymlink{
				TreeEntryBase: defaultFileTreeEntryBase(_acl, props),
				Target:        target,
			}
		default:
			// TODO: Or should we just ignore this entry? There might be more types in the future...
			return fmt.Errorf("Unknown tree entry type: %s", entry_type)
		}

		name, ok := props["name"]
		if !ok {
			return errors.New("Missing property: name")
		}

		t[name] = entry
	}

	return sc.Err()
}

func (a Tree) Equals(b Tree) bool {
	for k, va := range a {
		vb, ok := b[k]
		if !ok || va.Type() != vb.Type() || !va.equalContent(vb) {
			return false
		}
	}

	for k := range b {
		_, ok := a[k]
		if !ok {
			return false
		}
	}

	return true
}
