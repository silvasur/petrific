package storage

import (
	"bytes"
	"code.laria.me/petrific/objects"
	"errors"
	"fmt"
	"io"
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

type IdMismatchErr struct {
	Want, Have objects.ObjectId
}

func (iderr IdMismatchErr) Error() string {
	return fmt.Sprintf("ID verification failed: want %s, have %s", iderr.Want, iderr.Have)
}

// GetObjects gets an object from a Storage and parses and verifies it (check it's checksum/id)
func GetObject(s Storage, id objects.ObjectId) (objects.RawObject, error) {
	raw, err := s.Get(id)
	if err != nil {
		return objects.RawObject{}, err
	}

	idgen := id.Algo.Generator()
	r := io.TeeReader(bytes.NewReader(raw), idgen)

	obj, err := objects.Unserialize(r)
	if err != nil {
		return objects.RawObject{}, err
	}

	if have_id := idgen.GetId(); !have_id.Equals(id) {
		return objects.RawObject{}, IdMismatchErr{id, have_id}
	}
	return obj, nil
}

func GetObjectOfType(s Storage, id objects.ObjectId, t objects.ObjectType) (objects.Object, error) {
	rawobj, err := GetObject(s, id)
	if err != nil {
		return nil, err
	}

	if rawobj.Type != t {
		return nil, fmt.Errorf("GetObjectOfType: Wrong object type %s (want %s)", rawobj.Type, t)
	}

	return rawobj.Object()
}
