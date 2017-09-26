package cloud

import (
	"code.laria.me/petrific/config"
	"code.laria.me/petrific/storage"
	"fmt"
	"github.com/ncw/swift"
	"time"
)

type SwiftCloudStorage struct {
	con       *swift.Connection
	container string
}

type SwiftStorageConfigMandatoryError struct {
	Key string
	Err error
}

func (e SwiftStorageConfigMandatoryError) Error() string {
	return fmt.Sprintf("Could not get mandatory key %s for swift storage: %s", e.Key, e.Err.Error())
}

func SwiftStorageCreator() storage.CreateStorageFromConfig {
	return cloudStorageCreator(func(conf config.Config, name string) (CloudStorage, error) {
		storage_conf := conf.Storage[name]
		storage := SwiftCloudStorage{}
		storage.con = new(swift.Connection)

		// Mandatory options
		if err := storage_conf.Get("container", &storage.container); err != nil {
			return nil, SwiftStorageConfigMandatoryError{"container", err}
		}
		if err := storage_conf.Get("user_name", &storage.con.UserName); err != nil {
			return nil, SwiftStorageConfigMandatoryError{"user_name", err}
		}
		if err := storage_conf.Get("api_key", &storage.con.ApiKey); err != nil {
			return nil, SwiftStorageConfigMandatoryError{"api_key", err}
		}
		if err := storage_conf.Get("auth_url", &storage.con.AuthUrl); err != nil {
			return nil, SwiftStorageConfigMandatoryError{"auth_url", err}
		}

		// Optional... options

		storage_conf.Get("domain", &storage.con.Domain)
		storage_conf.Get("domain_id", &storage.con.DomainId)
		storage_conf.Get("user_id", &storage.con.UserId)
		storage_conf.Get("retries", &storage.con.Retries)
		storage_conf.Get("region", &storage.con.Region)
		storage_conf.Get("auth_version", &storage.con.AuthVersion)
		storage_conf.Get("tenant", &storage.con.Tenant)
		storage_conf.Get("tenant_id", &storage.con.TenantId)
		storage_conf.Get("tenant_domain", &storage.con.TenantDomain)
		storage_conf.Get("tenant_domain_id", &storage.con.TenantDomainId)
		storage_conf.Get("trust_id", &storage.con.TrustId)

		var connect_timeout_str string
		if storage_conf.Get("connect_timeout", &connect_timeout_str) == nil {
			d, err := time.ParseDuration(connect_timeout_str)
			if err != nil {
				return nil, err
			}
			storage.con.ConnectTimeout = d
		}

		var timeout_str string
		if storage_conf.Get("timeout", &timeout_str) == nil {
			d, err := time.ParseDuration(timeout_str)
			if err != nil {
				return nil, err
			}
			storage.con.Timeout = d
		}

		if err := storage.con.Authenticate(); err != nil {
			return nil, err
		}

		return storage, nil
	})
}

func (scs SwiftCloudStorage) Get(key string) ([]byte, error) {
	return scs.con.ObjectGetBytes(scs.container, key)
}

func (scs SwiftCloudStorage) Has(key string) (bool, error) {
	switch _, _, err := scs.con.Object(scs.container, key); err {
	case nil:
		return true, nil
	case swift.ObjectNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (scs SwiftCloudStorage) Put(key string, content []byte) error {
	err := scs.con.ObjectPutBytes(scs.container, key, content, "application/octet-stream")

	return err
}

func (scs SwiftCloudStorage) Delete(key string) error {
	return scs.con.ObjectDelete(scs.container, key)
}

func (scs SwiftCloudStorage) List(prefix string) ([]string, error) {
	return scs.con.ObjectNamesAll(scs.container, &swift.ObjectsOpts{
		Prefix: prefix,
	})
}

func (scs SwiftCloudStorage) Close() error {
	return nil
}
