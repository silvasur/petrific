package registry

import (
	"code.laria.me/petrific/storage"
	"code.laria.me/petrific/storage/cloud"
)

// List af all available storage types
var StorageTypes = map[string]storage.CreateStorageFromConfig{
	"local":           storage.LocalStorageFromConfig,
	"memory":          storage.MemoryStorageFromConfig,
	"openstack-swift": cloud.SwiftStorageCreator(),
}
