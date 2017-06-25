package objects

import (
	"bufio"
	"bytes"
	"errors"
	"strconv"
)

type FileFragment struct {
	Blob ObjectId
	Size uint64
}

func (ff FileFragment) toProperties() properties {
	return properties{"blob": ff.Blob.String(), "size": strconv.FormatUint(ff.Size, 10)}
}

func (ff *FileFragment) fromProperties(p properties) error {
	blob, ok := p["blob"]
	if !ok {
		return errors.New("Field `blob` is missing")
	}

	var err error
	ff.Blob, err = ParseObjectId(blob)
	if err != nil {
		return err
	}

	size, ok := p["size"]
	if !ok {
		return errors.New("Field `size` is missing")
	}

	ff.Size, err = strconv.ParseUint(size, 10, 64)
	return err
}

func (a FileFragment) Equals(b FileFragment) bool {
	return a.Blob.Equals(b.Blob) && a.Size == b.Size
}

type File []FileFragment

func (f File) Type() ObjectType {
	return OTFile
}

func (f File) Payload() []byte {
	out := []byte{}

	for _, ff := range f {
		b, err := ff.toProperties().MarshalText()
		if err != nil {
			panic(err)
		}

		out = append(out, b...)
		out = append(out, '\n')
	}

	return out
}

func (f *File) FromPayload(payload []byte) error {
	sc := bufio.NewScanner(bytes.NewReader(payload))

	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}

		props := make(properties)
		if err := props.UnmarshalText(line); err != nil {
			return nil
		}

		ff := FileFragment{}
		if err := ff.fromProperties(props); err != nil {
			return err
		}

		*f = append(*f, ff)
	}

	return sc.Err()
}

func (a File) Equals(b File) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !a[i].Equals(b[i]) {
			return false
		}
	}

	return true
}
