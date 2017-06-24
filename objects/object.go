package objects

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ObjectId interface {
	fmt.Stringer
}

type ObjectType string

const (
	OTBlob     ObjectType = "blob"
	OTFile     ObjectType = "file"
	OTTree     ObjectType = "tree"
	OTSnapshot ObjectType = "snapshot"
)

type Object struct {
	Type    ObjectType
	Payload []byte
}

func (o Object) Serialize(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s %d\n", o.Type, len(o.Payload)); err != nil {
		return err
	}

	_, err := w.Write(o.Payload)
	return err
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

func UnserializeObject(r io.Reader) (Object, error) {
	br := newBytewiseReader(r)

	line := []byte{}

	for {
		b, err := br.ReadByte()
		if err != nil {
			return Object{}, UnserializeError{err}
		}

		if b == '\n' {
			break
		}
		line = append(line, b)
	}

	parts := strings.SplitN(string(line), " ", 2)
	if len(parts) != 2 {
		return Object{}, UnserializeError{}
	}

	size, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return Object{}, UnserializeError{err}
	}

	o := Object{
		Type:    ObjectType(parts[0]),
		Payload: make([]byte, size),
	}

	if _, err := io.ReadFull(r, o.Payload); err != nil {
		return Object{}, UnserializeError{err}
	}

	return o, nil
}
