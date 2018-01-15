package filter

import (
	"bytes"
	"code.laria.me/petrific/objects"
	"code.laria.me/petrific/storage"
	"errors"
	"os/exec"
)

type Filter interface {
	Transform([]byte) ([]byte, error)
}

type PipeFilter []string

var emptyCommand = errors.New("Need at least one argument for pipeFilter")

func (pf PipeFilter) Transform(b []byte) ([]byte, error) {
	if len(pf) == 0 {
		return []byte{}, emptyCommand
	}

	cmd := exec.Command(pf[0], pf[1:]...)

	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stdin = bytes.NewReader(b)

	if err := cmd.Run(); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

// FilterSorage is a storage implementation wrapping around another storage, sending each raw object through an extrenal
// binary for custom de/encoding (think encryption, compression, ...).
//
// It is used in a configuration by using the method "filter". It needs the config key "base" referencing the name of
// another configured storage. Also needed are the string lists "decode" and "encode", describing which binary to call
// with which parameters.
//
// For example, here is a configuration for a filter storage wrapping a storage "foo",
// encrypting the content with gpg for the key "foobar"
//
//     [storage.foo_encrypted]
//     method="filter"
//     base="foo"
//     encode=["gpg", "--encrypt", "--recipient", "foobar"]
//     decode=["gpg", "--decrypt"]
type FilterStorage struct {
	Base           storage.Storage
	Decode, Encode Filter
}

func (filt FilterStorage) Get(id objects.ObjectId) ([]byte, error) {
	data, err := filt.Base.Get(id)
	if err != nil {
		return data, err
	}

	if filt.Decode == nil {
		return data, nil
	}

	return filt.Decode.Transform(data)
}

func (filt FilterStorage) Has(id objects.ObjectId) (bool, error) {
	return filt.Base.Has(id)
}

func (filt FilterStorage) Set(id objects.ObjectId, typ objects.ObjectType, raw []byte) error {
	if filt.Encode != nil {
		var err error
		raw, err = filt.Encode.Transform(raw)
		if err != nil {
			return err
		}
	}

	return filt.Base.Set(id, typ, raw)
}

func (filt FilterStorage) List(typ objects.ObjectType) ([]objects.ObjectId, error) {
	return filt.Base.List(typ)
}

func (filt FilterStorage) Subcmds() map[string]storage.StorageSubcmd {
	return filt.Base.Subcmds()
}

func (filt FilterStorage) Close() error {
	return filt.Base.Close()
}
