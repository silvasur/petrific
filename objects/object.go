package objects

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ObjectType string

const (
	OTBlob     ObjectType = "blob"
	OTFile     ObjectType = "file"
	OTTree     ObjectType = "tree"
	OTSnapshot ObjectType = "snapshot"
)

type RawObject struct {
	Type    ObjectType
	Payload []byte
}

// Serialize writes the binary representation of an object to a io.Writer
func (o RawObject) Serialize(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s %d\n", o.Type, len(o.Payload)); err != nil {
		return err
	}

	_, err := w.Write(o.Payload)
	return err
}

func (o RawObject) SerializeAndId(w io.Writer, algo ObjectIdAlgo) (ObjectId, error) {
	gen := algo.Generator()

	if err := o.Serialize(io.MultiWriter(w, gen)); err != nil {
		return ObjectId{}, err
	}

	return gen.GetId(), nil
}

type UnserializeError struct {
	Reason error
}

func (err UnserializeError) Error() string {
	if err.Reason == nil {
		return "Invalid object"
	} else {
		return fmt.Sprintf("Invalid object: %s", err.Reason)
	}
}

type bytewiseReader struct {
	r   io.Reader
	buf []byte
}

func newBytewiseReader(r io.Reader) io.ByteReader {
	return &bytewiseReader{
		r:   r,
		buf: make([]byte, 1),
	}
}

func (br *bytewiseReader) ReadByte() (b byte, err error) {
	_, err = br.r.Read(br.buf)
	b = br.buf[0]

	return
}

// Unserialize attempts to read an object from a stream.
// It is advisable to pass a buffered reader, if feasible.
func Unserialize(r io.Reader) (RawObject, error) {
	br := newBytewiseReader(r)

	line := []byte{}

	for {
		b, err := br.ReadByte()
		if err != nil {
			return RawObject{}, UnserializeError{err}
		}

		if b == '\n' {
			break
		}
		line = append(line, b)
	}

	parts := strings.SplitN(string(line), " ", 2)
	if len(parts) != 2 {
		return RawObject{}, UnserializeError{}
	}

	size, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return RawObject{}, UnserializeError{err}
	}

	o := RawObject{
		Type:    ObjectType(parts[0]),
		Payload: make([]byte, size),
	}

	if _, err := io.ReadFull(r, o.Payload); err != nil {
		return RawObject{}, UnserializeError{err}
	}

	return o, nil
}

type Object interface {
	Type() ObjectType
	Payload() []byte
	FromPayload([]byte) error
}

func (ro RawObject) Object() (o Object, err error) {
	switch ro.Type {
	case OTBlob:
		o = new(Blob)
	case OTFile:
		o = new(File)
	case OTTree:
		o = make(Tree)
	case OTSnapshot:
		o = new(Snapshot)
	default:
		return nil, fmt.Errorf("Unknown object type %s", ro.Type)
	}

	if err = o.FromPayload(ro.Payload); err != nil {
		o = nil
	}
	return
}

func ToRawObject(o Object) RawObject {
	return RawObject{Type: o.Type(), Payload: o.Payload()}
}
