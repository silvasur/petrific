package storage

import (
	"bytes"
	"code.laria.me/petrific/objects"
	"errors"
)

var (
	ObjectNotFound = errors.New("Object not found")
)

type Storage interface {
	Get(id objects.ObjectId) ([]byte, error)
	Has(id objects.ObjectId) (bool, error)
	Set(id objects.ObjectId, typ objects.ObjectType, raw []byte) error
	List(typ objects.ObjectType) ([]objects.ObjectId, error)
}

func SetObject(s Storage, o objects.RawObject) (id objects.ObjectId, err error) {
	buf := new(bytes.Buffer)

	id, err = o.SerializeAndId(buf, objects.OIdAlgoDefault)
	if err != nil {
		return
	}

	ok, err := s.Has(id)
	if err != nil {
		return
	}

	if !ok {
		err = s.Set(id, o.Type, buf.Bytes())
	}
	return
}
