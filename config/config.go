// Package config provides methods for configuring the petrific binary.
//
// The configuration file is located in `$XDG_CONFIG_HOME/petrific/config.toml`,
// where `$XDG_CONFIG_HOME` is typically `~/.config`.
//
// The configuration file is a TOML file, its main purpose is to define the used
// storage backends, it also defines which GPG key to use for snapshot signing.
//
// Here is an example configuration file:
//
//    # This config key defines the default storage backend, as defined below ([storage.local_compressed])
//    default_storage = "local_compressed"
//
//    # This defines the location of the cache file (can speed up creating backups; can be left out)
//    cache_path = "~/.cache/petrific.cache"
//
//    [signing]
//    # Use this GPG key to sign snapshots
//    key = "0123456789ABCDEF0123456789ABCDEF01234567"
//
//    # The storage.* sections define storage backends.
//    # Every section must contain the key `method`, the other keys depend on the selected method.
//    # For more details see the documentation for ../storage
//
//    [storage.local]
//    method="local"
//    path="~/.local/share/petrific"
//
//    [storage.local_compressed]
//    method="filter"
//    base="local"
//    encode=["zlib-flate", "-compress"]
//    decode=["zlib-flate", "-uncompress"]
package config

import (
	"code.laria.me/petrific/gpg"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
	"os"
	"os/user"
	"strings"
)

type StorageConfigErr struct {
	Name string
	Err  error
}

func (sce StorageConfigErr) Error() string {
	return fmt.Sprintf("Could not get configuration for storage %s: %s", sce.Name, sce.Err.Error())
}

var (
	StorageConfigNotFound = errors.New("Not found")
)

type Config struct {
	CachePath      string `toml:"cache_path,omitempty"`
	DefaultStorage string `toml:"default_storage"`
	Signing        struct {
		Key string
	}
	Storage map[string]toml.Primitive
	meta    toml.MetaData `toml:"-"`
}

func LoadConfig(path string) (config Config, err error) {
	if path == "" {
		path, err = xdg.ConfigFile("petrific/config.toml")
		if err != nil {
			return
		}
	}

	meta, err := toml.DecodeFile(path, &config)
	config.meta = meta
	return config, err
}

func (c Config) GPGSigner() gpg.Signer {
	return gpg.Signer{Key: c.Signing.Key}
}

func (c Config) GetStorageMethod(name string) (string, error) {
	prim, ok := c.Storage[name]
	if !ok {
		return "", StorageConfigErr{name, StorageConfigNotFound}
	}

	var method_wrap struct{ Method string }
	err := c.meta.PrimitiveDecode(prim, &method_wrap)
	if err != nil {
		return "", StorageConfigErr{name, err}
	}

	return method_wrap.Method, nil
}

func (c Config) GetStorageConfData(name string, v interface{}) error {
	prim, ok := c.Storage[name]
	if !ok {
		return StorageConfigErr{name, StorageConfigNotFound}
	}

	return c.meta.PrimitiveDecode(prim, v)
}

func ExpandTilde(path string) string {
	home, ok := os.LookupEnv("HOME")
	if !ok {
		u, err := user.Current()
		if err != nil {
			return path
		}
		home = u.HomeDir
	}

	if home == "" {
		return path
	}

	parts := strings.Split(path, string(os.PathSeparator))
	if len(parts) > 0 && parts[0] == "~" {
		parts[0] = home
	}
	return strings.Join(parts, string(os.PathSeparator))
}
