package cloud

import (
	"code.laria.me/petrific/config"
	"code.laria.me/petrific/storage"
	"github.com/ncw/swift"
	"time"
)

type SwiftCloudStorage struct {
	con       *swift.Connection
	container string
}

type SwiftConfig struct {
	// Mandatory options
	Container string `toml:"container"`
	UserName  string `toml:"user_name"`
	ApiKey    string `toml:"api_key"`
	AuthURL   string `toml:"auth_url"`

	// Optional... options
	Domain         string `toml:"domain,omitempty"`
	DomainID       string `toml:"domain_id,omitempty"`
	UserID         string `toml:"user_id,omitempty"`
	Retries        int    `toml:"retries,omitempty"`
	Region         string `toml:"region,omitempty"`
	AuthVersion    int    `toml:"auth_version,omitempty"`
	Tenant         string `toml:"tenant,omitempty"`
	TenantID       string `toml:"tenant_id,omitempty"`
	TenantDomain   string `toml:"tenant_domain,omitempty"`
	TenantDomainID string `toml:"tenant_domain_id,omitempty"`
	TrustID        string `toml:"trust_id,omitempty"`
	ConnectTimeout string `toml:"connect_timeout,omitempty"`
	Timeout        string `toml:"timeout,omitempty"`
}

func SwiftStorageCreator() storage.CreateStorageFromConfig {
	return cloudStorageCreator(func(conf config.Config, name string) (CloudStorage, error) {
		var storage_conf SwiftConfig

		if err := conf.GetStorageConfData(name, &storage_conf); err != nil {
			return nil, err
		}

		storage := SwiftCloudStorage{}
		storage.con = new(swift.Connection)

		storage.container = storage_conf.Container
		storage.con.UserName = storage_conf.UserName
		storage.con.ApiKey = storage_conf.ApiKey
		storage.con.AuthUrl = storage_conf.AuthURL

		storage.con.Domain = storage_conf.Domain
		storage.con.DomainId = storage_conf.DomainID
		storage.con.UserId = storage_conf.UserID
		storage.con.Retries = storage_conf.Retries
		storage.con.Region = storage_conf.Region
		storage.con.AuthVersion = storage_conf.AuthVersion
		storage.con.Tenant = storage_conf.Tenant
		storage.con.TenantId = storage_conf.TenantID
		storage.con.TenantDomain = storage_conf.TenantDomain
		storage.con.TenantDomainId = storage_conf.TenantDomainID
		storage.con.TrustId = storage_conf.TrustID

		if storage_conf.ConnectTimeout != "" {
			d, err := time.ParseDuration(storage_conf.ConnectTimeout)
			if err != nil {
				return nil, err
			}
			storage.con.ConnectTimeout = d
		}

		if storage_conf.Timeout != "" {
			d, err := time.ParseDuration(storage_conf.Timeout)
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
