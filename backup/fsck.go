package backup

import (
	"code.laria.me/petrific/logging"
	"code.laria.me/petrific/objects"
	"code.laria.me/petrific/storage"
	"fmt"
	"runtime"
	"strings"
	"sync"
)

type FsckProblemType int

const (
	FsckStorageError FsckProblemType = iota
	FsckDeserializationError
	FsckUnexpectedBlobSize
)

type FsckProblem struct {
	Id                 objects.ObjectId
	Ancestors          []AncestorInfo
	ProblemType        FsckProblemType
	Err                error
	WantSize, HaveSize int
}

func (problem FsckProblem) String() string {
	desc := ""

	switch problem.ProblemType {
	case FsckStorageError:
		desc = fmt.Sprintf("Failed retrieving object from storage: %s", problem.Err)
	case FsckDeserializationError:
		desc = fmt.Sprintf("Object could not be deserialized: %s", problem.Err)
	case FsckUnexpectedBlobSize:
		desc = fmt.Sprintf("Unexpected blob size: have %d, want %d", problem.HaveSize, problem.WantSize)
	}

	ancestors := make([]string, len(problem.Ancestors))
	for i, a := range problem.Ancestors {
		ancestors[i] = a.String()
	}

	return fmt.Sprintf("%s. Object %s (path: %s)", desc, problem.Id, strings.Join(ancestors, " / "))
}

type AncestorInfo struct {
	Id   objects.ObjectId
	Type objects.ObjectType
	Name string
}

func (a AncestorInfo) String() string {
	if a.Name == "" {
		return string(a.Type) + " " + a.Id.String()
	} else {
		return fmt.Sprintf("%s of %s %s", a.Name, a.Type, a.Id)
	}
}

type queueElement struct {
	Id        objects.ObjectId
	Ancestors []AncestorInfo
	Extra     interface{}
}

type fsckProcess struct {
	st       storage.Storage
	blobs    bool
	problems chan<- FsckProblem
	wait     *sync.WaitGroup
	queue    chan queueElement
	seen     map[string]struct{}
	seenLock sync.Locker
	log      *logging.Log
}

func (fsck fsckProcess) onlyUnseen(elems []queueElement) []queueElement {
	fsck.seenLock.Lock()
	defer fsck.seenLock.Unlock()

	newElems := make([]queueElement, 0, len(elems))
	for _, elem := range elems {
		id := elem.Id.String()
		_, ok := fsck.seen[id]
		if !ok {
			newElems = append(newElems, elem)

			fsck.seen[id] = struct{}{}
		}
		fsck.log.Debug().Printf("seen %s? %t", id, ok)
	}

	return newElems
}

func (fsck fsckProcess) enqueue(elems []queueElement) {
	fsck.log.Debug().Printf("enqueueing %d elements", len(elems))

	elems = fsck.onlyUnseen(elems)

	fsck.wait.Add(len(elems))
	go func() {
		for _, elem := range elems {
			fsck.log.Debug().Printf("enqueueing %v", elem)
			fsck.queue <- elem
		}
	}()
}

func (fsck fsckProcess) handle(elem queueElement) {
	defer fsck.wait.Done()

	rawobj, err := storage.GetObject(fsck.st, elem.Id)
	if err != nil {
		fsck.problems <- FsckProblem{
			Id:          elem.Id,
			Ancestors:   elem.Ancestors,
			ProblemType: FsckStorageError,
			Err:         err,
		}

		return
	}

	obj, err := rawobj.Object()
	if err != nil {
		fsck.problems <- FsckProblem{
			Id:          elem.Id,
			Ancestors:   elem.Ancestors,
			ProblemType: FsckDeserializationError,
			Err:         err,
		}

		return
	}

	switch obj.Type() {
	case objects.OTBlob:
		fsck.handleBlob(elem, obj.(*objects.Blob))
	case objects.OTFile:
		fsck.handleFile(elem, obj.(*objects.File))
	case objects.OTTree:
		fsck.handleTree(elem, obj.(objects.Tree))
	case objects.OTSnapshot:
		fsck.handleSnapshot(elem, obj.(*objects.Snapshot))
	}
}

func (fsck fsckProcess) handleBlob(elem queueElement, obj *objects.Blob) {
	if elem.Extra == nil {
		return
	}

	want, ok := elem.Extra.(int)
	if !ok {
		return
	}
	have := len(*obj)

	if have != want {
		fsck.problems <- FsckProblem{
			Id:          elem.Id,
			Ancestors:   elem.Ancestors,
			ProblemType: FsckUnexpectedBlobSize,
			WantSize:    want,
			HaveSize:    have,
		}
	}
}

func (fsck fsckProcess) handleFile(elem queueElement, obj *objects.File) {
	if !fsck.blobs {
		return
	}

	enqueue := make([]queueElement, 0, len(*obj))

	for _, fragment := range *obj {
		enqueue = append(enqueue, queueElement{
			Id: fragment.Blob,
			Ancestors: append(elem.Ancestors, AncestorInfo{
				Id:   elem.Id,
				Type: objects.OTFile,
			}),
			Extra: int(fragment.Size),
		})
	}

	fsck.enqueue(enqueue)
}

func (fsck fsckProcess) handleTree(elem queueElement, obj objects.Tree) {
	ancestors := func(name string) []AncestorInfo {
		return append(elem.Ancestors, AncestorInfo{
			Id:   elem.Id,
			Type: objects.OTTree,
			Name: name,
		})
	}

	enqueue := make([]queueElement, 0, len(obj))

	for name, entry := range obj {
		switch entry.Type() {
		case objects.TETDir:
			enqueue = append(enqueue, queueElement{
				Id:        entry.(objects.TreeEntryDir).Ref,
				Ancestors: ancestors(name),
			})
		case objects.TETFile:
			enqueue = append(enqueue, queueElement{
				Id:        entry.(objects.TreeEntryFile).Ref,
				Ancestors: ancestors(name),
			})
		}
	}

	fsck.enqueue(enqueue)
}

func (fsck fsckProcess) handleSnapshot(elem queueElement, obj *objects.Snapshot) {
	fsck.enqueue([]queueElement{
		{Id: obj.Tree},
	})
}

func (fsck fsckProcess) worker(i int) {
	for elem := range fsck.queue {
		fsck.handle(elem)
	}
	fsck.log.Debug().Printf("stopping worker %d", i)
}

// Fsck checks the consistency of objects in a storage
func Fsck(
	st storage.Storage,
	start *objects.ObjectId,
	blobs bool,
	problems chan<- FsckProblem,
	log *logging.Log,
) error {
	proc := fsckProcess{
		st:       st,
		blobs:    blobs,
		problems: problems,
		wait:     new(sync.WaitGroup),
		queue:    make(chan queueElement),
		seen:     make(map[string]struct{}),
		seenLock: new(sync.Mutex),
		log:      log,
	}

	enqueue := []queueElement{}

	if start == nil {
		types := []objects.ObjectType{
			objects.OTFile,
			objects.OTTree,
			objects.OTSnapshot,
		}
		for _, t := range types {
			ids, err := st.List(t)
			if err != nil {
				return err
			}

			for _, id := range ids {
				enqueue = append(enqueue, queueElement{Id: id})
			}
		}
	} else {
		enqueue = []queueElement{
			{Id: *start},
		}
	}

	if len(enqueue) == 0 {
		return nil
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		log.Debug().Printf("starting worker %d", i)
		go proc.worker(i)
	}

	proc.enqueue(enqueue)

	proc.wait.Wait()
	close(proc.queue)
	return nil
}
