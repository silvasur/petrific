package storage

import (
	"code.laria.me/petrific/config"
	"code.laria.me/petrific/objects"
)

type MemoryStorage struct {
	objects map[string][]byte
	bytype  map[objects.ObjectType][]objects.ObjectId
}

func NewMemoryStorage() Storage {
	return MemoryStorage{
		objects: make(map[string][]byte),
		bytype:  make(map[objects.ObjectType][]objects.ObjectId),
	}
}

func MemoryStorageFromConfig(conf config.Config, name string) (Storage, error) {
	return NewMemoryStorage(), nil
}

func (ms MemoryStorage) Get(id objects.ObjectId) ([]byte, error) {
	b, ok := ms.objects[id.String()]
	if !ok {
		return nil, ObjectNotFound
	}
	return b, nil
}

func (ms MemoryStorage) Has(id objects.ObjectId) (bool, error) {
	_, ok := ms.objects[id.String()]
	return ok, nil
}

func (ms MemoryStorage) Set(id objects.ObjectId, typ objects.ObjectType, raw []byte) error {
	ms.objects[id.String()] = raw
	ms.bytype[typ] = append(ms.bytype[typ], id)

	return nil
}

func (ms MemoryStorage) List(typ objects.ObjectType) ([]objects.ObjectId, error) {
	return ms.bytype[typ], nil
}

func (MemoryStorage) Close() error {
	return nil
}