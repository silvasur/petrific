package cache

import (
	"bufio"
	"code.laria.me/petrific/objects"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type Cache interface {
	PathUpdated(path string) (mtime time.Time, id objects.ObjectId, ok bool)
	SetPathUpdated(path string, mtime time.Time, id objects.ObjectId)
	Close() error
}

type NopCache struct{}

func (NopCache) PathUpdated(_ string) (_ time.Time, _ objects.ObjectId, ok bool) {
	ok = false
	return
}

func (NopCache) SetPathUpdated(_ string, _ time.Time, _ objects.ObjectId) {}

func (NopCache) Close() error { return nil }

type fileCacheEntry struct {
	mtime time.Time
	id    objects.ObjectId
}

type FileCache struct {
	cache    map[string]fileCacheEntry
	location string
}

func (fc FileCache) PathUpdated(path string) (time.Time, objects.ObjectId, bool) {
	entry, ok := fc.cache[path]
	return entry.mtime, entry.id, ok
}

func (fc FileCache) SetPathUpdated(path string, mtime time.Time, id objects.ObjectId) {
	fc.cache[path] = fileCacheEntry{mtime, id}
}

func NewFileCache(location string) FileCache {
	return FileCache{make(map[string]fileCacheEntry), location}
}

func escapeName(name string) string {
	name = strings.Replace(name, "\\", "\\\\", -1)
	name = strings.Replace(name, "\n", "\\n", -1)
	return name
}

func unescapeName(name string) string {
	name = strings.Replace(name, "\\n", "\n", -1)
	name = strings.Replace(name, "\\\\", "\\", -1)
	return name
}

func (fc FileCache) dump(w io.Writer) error {
	for path, entry := range fc.cache {
		if _, err := fmt.Fprintf(
			w,
			"%s %d %d %s\n",
			entry.id,
			entry.mtime.Unix(),
			entry.mtime.Nanosecond(),
			escapeName(path),
		); err != nil {
			return err
		}
	}

	return nil
}

func (fc FileCache) load(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 4)
		if len(parts) != 4 {
			return fmt.Errorf("Could not load FileCache: Expected 4 entries, got %d", len(parts))
		}

		id, err := objects.ParseObjectId(parts[0])
		if err != nil {
			return err
		}

		sec, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return err
		}

		nsec, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return err
		}

		fc.cache[unescapeName(parts[3])] = fileCacheEntry{time.Unix(sec, nsec), id}
	}

	return scanner.Err()
}

func (fc FileCache) Load() error {
	f, err := os.Open(fc.location)
	switch {
	case os.IsNotExist(err):
		return nil
	case err != nil:
		return err
	default:
	}
	defer f.Close()

	return fc.load(f)
}

func (fc FileCache) Close() error {
	f, err := os.Create(fc.location)
	if err != nil {
		return err
	}
	defer f.Close()

	return fc.dump(f)
}
