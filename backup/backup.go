package backup

import (
	"code.laria.me/petrific/cache"
	"code.laria.me/petrific/fs"
	"code.laria.me/petrific/objects"
	"code.laria.me/petrific/storage"
	"io"
	"runtime"
	"sort"
	"time"
)

type filesLast []fs.File

func (fl filesLast) Len() int           { return len(fl) }
func (fl filesLast) Less(i, j int) bool { return fl[i].Type() != fs.FFile && fl[j].Type() == fs.FFile }
func (fl filesLast) Swap(i, j int)      { fl[i], fl[j] = fl[j], fl[i] }

type writeFileTaskResult struct {
	file    fs.RegularFile
	file_id objects.ObjectId
	err     error
}

type writeFileTask struct {
	file   fs.RegularFile
	result *chan writeFileTaskResult
}

func (task writeFileTask) process(store storage.Storage) {
	var result writeFileTaskResult
	result.file = task.file

	rwc, err := task.file.Open()
	if err != nil {
		result.err = err
		*(task.result) <- result
		return
	}
	defer rwc.Close()

	result.file_id, result.err = WriteFile(store, rwc)

	*(task.result) <- result
}

type writeDirProcess struct {
	queue chan writeFileTask
	store storage.Storage
}

func (proc writeDirProcess) worker() {
	for task := range proc.queue {
		task.process(proc.store)
	}
}

func (proc writeDirProcess) enqueue(file fs.RegularFile, result *chan writeFileTaskResult) {
	go func() {
		proc.queue <- writeFileTask{file, result}
	}()
}

func (proc writeDirProcess) stop() {
	close(proc.queue)
}

func WriteDir(
	store storage.Storage,
	abspath string,
	d fs.Dir,
	pcache cache.Cache,
) (objects.ObjectId, error) {
	proc := writeDirProcess{
		make(chan writeFileTask),
		store,
	}
	defer proc.stop()

	for i := 0; i < runtime.NumCPU(); i++ {
		go proc.worker()
	}

	return proc.writeDir(abspath, d, pcache)
}

func (proc writeDirProcess) writeDir(
	abspath string,
	d fs.Dir,
	pcache cache.Cache,
) (objects.ObjectId, error) {
	_children, err := d.Readdir()
	if err != nil {
		return objects.ObjectId{}, err
	}

	file_results := make(chan writeFileTaskResult)
	wait_for_files := 0

	// fs.FFile entries must be processed at the end, since they are processed concurrently and might otherwise lock us up
	children := filesLast(_children)
	sort.Sort(children)

	infos := make(objects.Tree)
	for _, c := range children {
		var info objects.TreeEntry = nil

		switch c.Type() {
		case fs.FFile:
			mtime, file_id, ok := pcache.PathUpdated(abspath + "/" + c.Name())
			if !ok || mtime.Before(c.ModTime()) {
				// According to cache the file was changed

				proc.enqueue(c.(fs.RegularFile), &file_results)
				wait_for_files++
			} else {
				info = objects.NewTreeEntryFile(file_id, c.Executable())
			}
		case fs.FDir:
			subtree_id, err := proc.writeDir(abspath+"/"+c.Name(), c.(fs.Dir), pcache)
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

	for ; wait_for_files > 0; wait_for_files-- {
		result := <-file_results
		if result.err != nil {
			err = result.err
			wait_for_files--
			break
		}

		pcache.SetPathUpdated(abspath+"/"+result.file.Name(), result.file.ModTime(), result.file_id)
		infos[result.file.Name()] = objects.NewTreeEntryFile(result.file_id, result.file.Executable())
	}

	for ; wait_for_files > 0; wait_for_files-- {
		// Drain remaining results in case of error
		<-file_results
	}

	if err != nil {
		return objects.ObjectId{}, err
	}

	return storage.SetObject(proc.store, objects.ToRawObject(infos))
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
