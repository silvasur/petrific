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

func (filt FilterStorage) Close() error {
	return filt.Base.Close()
}
