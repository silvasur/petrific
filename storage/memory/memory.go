package memory

import (
	"github.com/silvasur/petrific/config"
	"github.com/silvasur/petrific/objects"
	"github.com/silvasur/petrific/storage"
)

// Memory storage is an in-memory storage. It is rather useless when using petrific, it is mostly used for internal testing.
// But if you want to use the memory storage anyway, you can do that by putting a storage section with the method
// "memory" in your config file
type MemoryStorage struct {
	objects map[string][]byte
	bytype  map[objects.ObjectType][]objects.ObjectId
}

func NewMemoryStorage() storage.Storage {
	return MemoryStorage{
		objects: make(map[string][]byte),
		bytype:  make(map[objects.ObjectType][]objects.ObjectId),
	}
}

func MemoryStorageFromConfig(conf config.Config, name string) (storage.Storage, error) {
	return NewMemoryStorage(), nil
}

func copyBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (ms MemoryStorage) Get(id objects.ObjectId) ([]byte, error) {
	b, ok := ms.objects[id.String()]
	if !ok {
		return nil, storage.ObjectNotFound
	}
	return copyBytes(b), nil
}

func (ms MemoryStorage) Has(id objects.ObjectId) (bool, error) {
	_, ok := ms.objects[id.String()]
	return ok, nil
}

func (ms MemoryStorage) Set(id objects.ObjectId, typ objects.ObjectType, raw []byte) error {
	ms.objects[id.String()] = copyBytes(raw)
	ms.bytype[typ] = append(ms.bytype[typ], id)

	return nil
}

func (ms MemoryStorage) List(typ objects.ObjectType) ([]objects.ObjectId, error) {
	return ms.bytype[typ], nil
}

func (MemoryStorage) Close() error {
	return nil
}
