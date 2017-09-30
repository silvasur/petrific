// Package cloud provides utilities to implement a petrific storage using a cloud-based object-storage (S3/Openstack Swift style)

package cloud

import (
	"bytes"
	"code.laria.me/petrific/config"
	"code.laria.me/petrific/gpg"
	"code.laria.me/petrific/objects"
	"code.laria.me/petrific/storage"
	"errors"
	"fmt"
	"math/rand"
)

type CloudStorage interface {
	Get(key string) ([]byte, error)
	Has(key string) (bool, error)
	Put(key string, content []byte) error
	Delete(key string) error
	List(prefix string) ([]string, error)

	Close() error
}

var (
	NotFoundErr = errors.New("Object not found") // Cloud object could not be found
)

// Crypter provides de-/encrypting facilities for CloudBasedObjectStorage
type Crypter interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

// NopCrypter implements Crypter by not de/-encrypting at all
type NopCrypter struct{}

func (NopCrypter) Encrypt(in []byte) ([]byte, error) { return in, nil }
func (NopCrypter) Decrypt(in []byte) ([]byte, error) { return in, nil }

type CloudBasedObjectStorage struct {
	CS      CloudStorage
	Prefix  string
	Crypter Crypter

	index storage.Index
}

func (cbos CloudBasedObjectStorage) objidToKey(id objects.ObjectId) string {
	return cbos.Prefix + "obj/" + id.String()
}

func (cbos CloudBasedObjectStorage) readIndex(name string) (storage.Index, error) {
	index := storage.NewIndex()

	b, err := cbos.CS.Get(name)
	if err != nil {
		return index, err
	}

	err = index.Load(bytes.NewReader(b))
	return index, err
}

func (cbos *CloudBasedObjectStorage) Init() error {
	cbos.index = storage.NewIndex()

	// Load and combine all indexes, keep only the one with the "largest" name (see also Close())
	index_names, err := cbos.CS.List(cbos.Prefix + "index/")
	if err != nil {
		return err
	}

	max_index := ""
	for _, index_name := range index_names {
		index, err := cbos.readIndex(index_name)
		if err != nil {
			return err
		}

		cbos.index.Combine(index)
	}

	for _, index_name := range index_names {
		if index_name != max_index {
			if err := cbos.CS.Delete(index_name); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cbos CloudBasedObjectStorage) Get(id objects.ObjectId) ([]byte, error) {
	b, err := cbos.CS.Get(cbos.objidToKey(id))
	if err != nil {
		return []byte{}, err
	}

	return cbos.Crypter.Decrypt(b)
}

func (cbos CloudBasedObjectStorage) Has(id objects.ObjectId) (bool, error) {
	return cbos.CS.Has(cbos.objidToKey(id))
}

func (cbos CloudBasedObjectStorage) Set(id objects.ObjectId, typ objects.ObjectType, raw []byte) error {
	b, err := cbos.Crypter.Encrypt(raw)
	if err != nil {
		return err
	}

	if err := cbos.CS.Put(cbos.objidToKey(id), b); err != nil {
		return err
	}

	// could be used to repopulate the index (not implemented yet)
	if err := cbos.CS.Put(cbos.Prefix+"typeof/"+id.String(), []byte(typ)); err != nil {
		return err
	}

	cbos.index.Set(id, typ)

	return nil
}

func (cbos CloudBasedObjectStorage) List(typ objects.ObjectType) ([]objects.ObjectId, error) {
	return cbos.index.List(typ), nil
}

func (cbos CloudBasedObjectStorage) Close() (outerr error) {
	defer func() {
		err := cbos.CS.Close()
		if outerr == nil {
			outerr = err
		}
	}()

	// We need to adress the problem of parallel index creation here.
	// We handle this by adding a random hex number to the index name.
	// When loading the index, all "index/*" objects will be read and combined
	// and all but the one with the largest number will be deleted.

	buf := new(bytes.Buffer)
	if outerr = cbos.index.Save(buf); outerr != nil {
		return outerr
	}

	index_name := fmt.Sprintf("%sindex/%016x", cbos.Prefix, rand.Int63())
	return cbos.CS.Put(index_name, buf.Bytes())
}

type cloudObjectStorageCreator func(conf config.Config, name string) (CloudStorage, error)

func cloudStorageCreator(cloudCreator cloudObjectStorageCreator) storage.CreateStorageFromConfig {
	return func(conf config.Config, name string) (storage.Storage, error) {
		var cbos CloudBasedObjectStorage

		var storageconf struct {
			Prefix     string `toml:"prefix,omitempty"`
			EncryptFor string `toml:"encrypt_for,omitempty"`
		}

		if err := conf.GetStorageConfData(name, &storageconf); err != nil {
			return nil, err
		}

		cbos.Prefix = storageconf.Prefix

		if storageconf.EncryptFor != "" {
			cbos.Crypter = gpg.Crypter{
				gpg.Encrypter{Key: storageconf.EncryptFor},
				gpg.Decrypter{},
			}
		}

		var err error
		if cbos.CS, err = cloudCreator(conf, name); err != nil {
			return nil, err
		}

		err = cbos.Init()
		return cbos, err
	}
}
