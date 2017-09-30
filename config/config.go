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
	return gpg.Signer{c.Signing.Key}
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
