package registry

import (
	"code.laria.me/petrific/config"
	"code.laria.me/petrific/storage"
	"code.laria.me/petrific/storage/cloud"
	"code.laria.me/petrific/storage/local"
	"code.laria.me/petrific/storage/memory"
	"errors"
	"fmt"
)

// List af all available storage types
var StorageTypes = map[string]storage.CreateStorageFromConfig{
	"local":           local.LocalStorageFromConfig,
	"memory":          memory.MemoryStorageFromConfig,
	"openstack-swift": cloud.SwiftStorageCreator(),
}

var notFoundErr = errors.New("Storage not found")

type unknownMethodErr string

func (method unknownMethodErr) Error() string {
	return fmt.Sprintf("Method %s unknown", string(method))
}

type storageConfErr struct {
	name string
	err  error
}

func (e storageConfErr) Error() string {
	return fmt.Sprintf("Failed setting up storage %s: %s", e.name, e.err.Error())
}

func loadStorage(conf config.Config, storageName string) (storage.Storage, error) {
	method, err := conf.GetStorageMethod(storageName)

	if err != nil {
		return nil, err
	}

	st, ok := StorageTypes[method]
	if !ok {
		return nil, unknownMethodErr(method)
	}

	s, err := st(conf, storageName)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func LoadStorage(conf config.Config, storageName string) (storage.Storage, error) {
	s, err := loadStorage(conf, storageName)
	if err != nil {
		return nil, storageConfErr{storageName, err}
	}
	return s, nil
}
