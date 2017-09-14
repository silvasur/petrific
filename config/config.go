package config

import (
	"code.laria.me/petrific/gpg"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
	"os"
	"os/user"
	"reflect"
	"strings"
)

type StorageConfig map[string]interface{}

type Config struct {
	DefaultStorage string `toml:"default_storage"`
	Signing        struct {
		Key string
	}
	Storage map[string]StorageConfig
}

func LoadConfig(path string) (config Config, err error) {
	if path == "" {
		path, err = xdg.ConfigFile("petrific/config.toml")
		if err != nil {
			return
		}
	}

	_, err = toml.DecodeFile(path, &config)
	return
}

func (c Config) GPGSigner() gpg.Signer {
	return gpg.Signer{c.Signing.Key}
}

// Get gets a value from the StorageConfig, taking care about type checking.
// ptr must be a pointer or this method will panic.
// ptr will only be changed, if returned error != nil
func (s StorageConfig) Get(k string, ptr interface{}) error {
	ptrval := reflect.ValueOf(ptr)

	if ptrval.Kind() != reflect.Ptr {
		panic("ptr must be a pointer")
	}

	ptrelem := ptrval.Elem()
	if !ptrelem.CanSet() {
		panic("*ptr not settable?!")
	}

	v, ok := s[k]
	if !ok {
		return fmt.Errorf("Key '%s' is missing", k)
	}

	vval := reflect.ValueOf(v)

	if vval.Type() != ptrelem.Type() {
		return fmt.Errorf("Expected config '%s' to be of type %s, got %s", k, ptrelem.Type(), vval.Type())
	}

	ptrelem.Set(vval)
	return nil
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
