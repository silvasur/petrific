package registry

import (
	"errors"
	"fmt"
	"github.com/silvasur/petrific/config"
	"github.com/silvasur/petrific/storage"
	"github.com/silvasur/petrific/storage/cloud"
	"github.com/silvasur/petrific/storage/local"
	"github.com/silvasur/petrific/storage/memory"
)

// List af all available storage types
func getStorageTypes() map[string]storage.CreateStorageFromConfig {
	return map[string]storage.CreateStorageFromConfig{
		"local":           local.LocalStorageFromConfig,
		"memory":          memory.MemoryStorageFromConfig,
		"filter":          filterStorageFromConfig,
		"openstack-swift": cloud.SwiftStorageCreator(),
	}
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

	st, ok := getStorageTypes()[method]
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
