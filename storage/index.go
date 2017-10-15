package storage

import (
	"bufio"
	"fmt"
	"github.com/silvasur/petrific/objects"
	"io"
	"strings"
)

type Index map[objects.ObjectType]map[string]struct{}

func NewIndex() Index {
	idx := make(Index)
	idx.Init()
	return idx
}

func (idx Index) Init() {
	for _, t := range objects.AllObjectTypes {
		idx[t] = make(map[string]struct{})
	}
}

func (idx Index) Set(id objects.ObjectId, typ objects.ObjectType) {
	idx[typ][id.String()] = struct{}{}
}

func (idx Index) List(typ objects.ObjectType) []objects.ObjectId {
	ids := make([]objects.ObjectId, 0, len(idx[typ]))
	for id := range idx[typ] {
		ids = append(ids, objects.MustParseObjectId(id))
	}

	return ids
}

func (idx Index) Save(w io.Writer) error {
	for t, objs := range idx {
		for id := range objs {
			if _, err := fmt.Fprintf(w, "%s %s\n", t, id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (idx Index) Load(r io.Reader) error {
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		line := scan.Text()

		parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
		if len(parts) == 2 {
			id, err := objects.ParseObjectId(parts[1])
			if err != nil {
				return err
			}

			typ := objects.ObjectType(parts[0])

			if _, ok := idx[typ]; !ok {
				return fmt.Errorf("Failed loading index: Unknown ObjectType %s", typ)
			}

			idx[typ][id.String()] = struct{}{}
		}
	}
	return scan.Err()
}

func (a Index) Combine(b Index) {
	for t, objs := range b {
		for id := range objs {
			a[t][id] = struct{}{}
		}
	}
}
