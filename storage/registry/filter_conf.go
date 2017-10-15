package registry

import (
	"github.com/silvasur/petrific/config"
	"github.com/silvasur/petrific/storage"
	"github.com/silvasur/petrific/storage/filter"
)

// Unlike the other storage engines, we can not define FilterStorages
// *FromConfig function in the package itself, because we need to reference the
// registry package and circular imports are not allowed

func filterStorageFromConfig(conf config.Config, name string) (storage.Storage, error) {
	var storage_conf struct {
		Base   string
		Encode []string
		Decode []string
	}

	if err := conf.GetStorageConfData(name, &storage_conf); err != nil {
		return nil, err
	}

	base, err := LoadStorage(conf, storage_conf.Base)
	if err != nil {
		return nil, err
	}

	st := filter.FilterStorage{Base: base}

	if len(storage_conf.Encode) > 0 {
		st.Encode = filter.PipeFilter(storage_conf.Encode)
	}
	if len(storage_conf.Decode) > 0 {
		st.Decode = filter.PipeFilter(storage_conf.Decode)
	}

	return st, nil
}
