package objects

import (
	"bufio"
	"bytes"
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
	equalContent(TreeEntry) bool
	toProperties() properties
}

func compareTreeEntries(a, b TreeEntry) bool {
	return a.Type() == b.Type() && a.equalContent(b)
}

type TreeEntryFile ObjectId

func (tef TreeEntryFile) Type() TreeEntryType {
	return TETFile
}

func (tef TreeEntryFile) toProperties() properties {
	return properties{"ref": ObjectId(tef).String()}
}

func (a TreeEntryFile) equalContent(_b TreeEntry) bool {
	b, ok := _b.(TreeEntryFile)
	return ok && ObjectId(b).Equals(ObjectId(a))
}

type TreeEntryDir ObjectId

func (ted TreeEntryDir) Type() TreeEntryType {
	return TETDir
}

func (ted TreeEntryDir) toProperties() properties {
	return properties{"ref": ObjectId(ted).String()}
}

func (a TreeEntryDir) equalContent(_b TreeEntry) bool {
	b, ok := _b.(TreeEntryDir)
	return ok && ObjectId(b).Equals(ObjectId(a))
}

type TreeEntrySymlink string

func (tes TreeEntrySymlink) Type() TreeEntryType {
	return TETSymlink
}

func (tes TreeEntrySymlink) toProperties() properties {
	return properties{"target": string(tes)}
}

func (a TreeEntrySymlink) equalContent(_b TreeEntry) bool {
	b, ok := _b.(TreeEntrySymlink)
	return ok && b == a
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

func (t Tree) FromPayload(payload []byte) error {
	sc := bufio.NewScanner(bytes.NewReader(payload))

	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}

		props := make(properties)
		if err := props.UnmarshalText(line); err != nil {
			return nil
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
			entry = TreeEntryFile(ref)
		case TETDir:
			ref, err := getObjectIdFromProps(props, "ref")
			if err != nil {
				return err
			}
			entry = TreeEntryDir(ref)
		case TETSymlink:
			target, ok := props["target"]
			if !ok {
				return errors.New("Missing key: target")
			}
			entry = TreeEntrySymlink(target)
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
