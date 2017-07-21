package backup

import (
	"code.laria.me/petrific/cache"
	"code.laria.me/petrific/fs"
	"code.laria.me/petrific/objects"
	"code.laria.me/petrific/storage"
	"io"
	"time"
)

func WriteDir(
	store storage.Storage,
	abspath string,
	d fs.Dir,
	pcache cache.Cache,
) (objects.ObjectId, error) {
	children, err := d.Readdir()
	if err != nil {
		return objects.ObjectId{}, err
	}

	infos := make(objects.Tree)
	for _, c := range children {
		var info objects.TreeEntry

		switch c.Type() {
		case fs.FFile:
			mtime, file_id, ok := pcache.PathUpdated(abspath)
			if !ok || mtime.Before(c.ModTime()) {
				// According to cache the file was changed

				rwc, err := c.(fs.RegularFile).Open()
				if err != nil {
					return objects.ObjectId{}, err
				}

				file_id, err = WriteFile(store, rwc)
				rwc.Close()
				if err != nil {
					return objects.ObjectId{}, err
				}

				pcache.SetPathUpdated(abspath+"/"+c.Name(), c.ModTime(), file_id)
			}

			info = objects.NewTreeEntryFile(file_id, c.Executable())
		case fs.FDir:
			subtree_id, err := WriteDir(store, abspath+"/"+c.Name(), c.(fs.Dir), pcache)
			if err != nil {
				return objects.ObjectId{}, err
			}

			info = objects.NewTreeEntryDir(subtree_id, c.Executable())
		case fs.FSymlink:
			target, err := c.(fs.Symlink).Readlink()
			if err != nil {
				return objects.ObjectId{}, err
			}

			info = objects.NewTreeEntrySymlink(target, c.Executable())
		}

		if info != nil {
			infos[c.Name()] = info
		}
	}

	return storage.SetObject(store, objects.ToRawObject(infos))
}

const BlobChunkSize = 16 * 1024 * 1024 // 16MB

func WriteFile(store storage.Storage, r io.Reader) (objects.ObjectId, error) {
	// Right now we will create the file as fixed chunks of size BlobChunkSize
	// It Would be more efficient to use a dynamic/content aware chunking method but this can be added later.
	// The way files are serialized allows for any chunk size and addition of more properties in the future while staying compatible

	fragments := make(objects.File, 0)

	read_buf := make([]byte, BlobChunkSize)
	for {
		n, err := io.ReadFull(r, read_buf)
		if err == nil || err == io.ErrUnexpectedEOF {
			content := objects.Blob(read_buf[:n])
			blob_id, err := storage.SetObject(store, objects.ToRawObject(&content))
			if err != nil {
				return objects.ObjectId{}, err
			}

			fragments = append(fragments, objects.FileFragment{Blob: blob_id, Size: uint64(n)})
		} else if err == io.EOF {
			break
		} else {
			return objects.ObjectId{}, err
		}
	}

	return storage.SetObject(store, objects.ToRawObject(&fragments))
}

func CreateSnapshot(
	store storage.Storage,
	tree_id objects.ObjectId,
	date time.Time,
	archive string,
	comment string,
) (objects.ObjectId, error) {
	return storage.SetObject(store, objects.ToRawObject(&objects.Snapshot{
		Tree:    tree_id,
		Date:    date,
		Archive: archive,
		Comment: comment,
	}))
}
