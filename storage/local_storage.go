package storage

import (
	"bufio"
	"bytes"
	"code.laria.me/petrific/config"
	"code.laria.me/petrific/objects"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

func joinPath(parts ...string) string {
	return strings.Join(parts, string(os.PathSeparator))
}

func objectDir(id objects.ObjectId) string {
	return joinPath(string(id.Algo), hex.EncodeToString(id.Sum[0:1]))
}

type LocalStorage struct {
	Path  string
	index map[objects.ObjectType]map[string]struct{}
}

func LocalStorageFromConfig(conf config.Config, name string) (Storage, error) {
	var path string
	if err := conf.Storage[name].Get("path", &path); err != nil {
		return nil, err
	}

	return OpenLocalStorage(config.ExpandTilde(path))
}

func OpenLocalStorage(path string) (l LocalStorage, err error) {
	l.Path = path
	l.index = make(map[objects.ObjectType]map[string]struct{})
	for _, t := range objects.AllObjectTypes {
		l.index[t] = make(map[string]struct{})
	}

	if fi, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return l, err
		}
	} else if err != nil {
		return l, err
	} else if !fi.Mode().IsDir() {
		return l, fmt.Errorf("%s: Not a directory", path)
	}

	f, err := os.Open(joinPath(path, "index"))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return l, err
	}

	if err == nil {
		defer f.Close()

		scan := bufio.NewScanner(f)
		for scan.Scan() {
			line := scan.Text()

			parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
			if len(parts) == 2 {
				id, err := objects.ParseObjectId(parts[1])
				if err != nil {
					return l, err
				}

				l.index[objects.ObjectType(parts[0])][id.String()] = struct{}{}
			}
		}
		err = scan.Err()
	}

	return
}

func objectPath(id objects.ObjectId) string {
	return joinPath(objectDir(id), hex.EncodeToString(id.Sum[1:]))
}

func (l LocalStorage) Get(id objects.ObjectId) ([]byte, error) {
	f, err := os.Open(joinPath(l.Path, objectPath(id)))
	if os.IsNotExist(err) {
		return []byte{}, ObjectNotFound
	} else if err != nil {
		return []byte{}, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf.Bytes(), err
}

func (l LocalStorage) Has(id objects.ObjectId) (bool, error) {
	_, err := os.Stat(joinPath(l.Path, objectPath(id)))
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (l LocalStorage) Set(id objects.ObjectId, typ objects.ObjectType, raw []byte) error {
	// First, check if the directory exists
	dir := joinPath(l.Path, objectDir(id))
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	f, err := os.Create(joinPath(l.Path, objectPath(id)))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(raw)
	l.index[typ][id.String()] = struct{}{}
	return err
}

func (l LocalStorage) List(typ objects.ObjectType) ([]objects.ObjectId, error) {
	ids := make([]objects.ObjectId, 0, len(l.index[typ]))
	for id := range l.index[typ] {
		ids = append(ids, objects.MustParseObjectId(id))
	}
	return ids, nil
}

func (l LocalStorage) Close() (outerr error) {
	f, err := os.Create(joinPath(l.Path, "index"))
	if err != nil {
		return err
	}
	defer f.Close()

	for t, objs := range l.index {
		for id := range objs {
			if _, err := fmt.Fprintf(f, "%s %s\n", t, id); err != nil {
				return err
			}
		}
	}
	return nil
}
