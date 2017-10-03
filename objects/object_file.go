package objects

import (
	"bufio"
	"bytes"
	"errors"
	"strconv"
)

// FileFragment describes a fragment of a file. It consists of the ID of a blob and the size of the blob.
// The size is not necessary for reconstruction (the blob object already has a size)
// but it can speed up random access to the whole file by skipping previous fragments.
// It is serialized as `Properties` (see there) with the keys `blob` (ID of blob object) and `size` (decimal size of the blob)
type FileFragment struct {
	Blob ObjectId
	Size uint64
}

func (ff FileFragment) toProperties() Properties {
	return Properties{"blob": ff.Blob.String(), "size": strconv.FormatUint(ff.Size, 10)}
}

func (ff *FileFragment) fromProperties(p Properties) error {
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

// File describes a file object which ties together multiple Blob objects to a file.
// It is an ordered list of `FileFragment`s. The referenced blobs concatencated in the given order are the content of the file.
//
// Example:
//
//     blob=sha3-256:1111111111111111111111111111111111111111111111111111111111111111&size=123
//     blob=sha3-256:1111111111111111111111111111111111111111111111111111111111111111&size=123
//     blob=sha3-256:2222222222222222222222222222222222222222222222222222222222222222&size=10
//
// The file described by this object is 123+123+10 = 256 bytes long and consists of the blobs 111... (two times) and 222...
//
// Since the blob IDs and sizes don't change unless the file itself changes and the format will always serialize the same,
// serializing the same file twice results in exactly the same object with exactly the same ID. It will therefor only be stored once.
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

		props := make(Properties)
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
