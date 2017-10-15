package filter

import (
	"bytes"
	"github.com/silvasur/petrific/backup"
	"github.com/silvasur/petrific/storage/memory"
	"testing"
)

type XorFilter byte

func (xf XorFilter) Transform(b []byte) ([]byte, error) {
	for i := range b {
		b[i] ^= byte(xf)
	}
	return b, nil
}

func genTestData(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i & 0xff)
	}
	return data
}

func TestXorReverse(t *testing.T) {
	orig := genTestData(1 << 20)

	filter := XorFilter(42)

	filtered := make([]byte, len(orig))
	copy(filtered, orig)

	filtered, _ = filter.Transform(filtered)
	filtered, _ = filter.Transform(filtered)

	if !bytes.Equal(filtered, orig) {
		t.Errorf("orig != filtered:\n{%x}\n{%x}", orig, filtered)
	}
}

func testFilter(t *testing.T, encode, decode Filter) {
	base := memory.NewMemoryStorage()
	storage := FilterStorage{
		Base:   base,
		Encode: encode,
		Decode: decode,
	}

	size := backup.BlobChunkSize*2 + 10
	data := genTestData(size)

	id, err := backup.WriteFile(storage, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Unexpeced error from WriteFile: %s", err)
	}

	buf := new(bytes.Buffer)
	if err := backup.RestoreFile(storage, id, buf); err != nil {
		t.Fatalf("Unexpeced error from RestoreFile: %s", err)
	}

	if !bytes.Equal(data, buf.Bytes()) {
		t.Errorf("data != buf.Bytes()")
	}

}

func TestFilterStorage(t *testing.T) {
	testFilter(t, XorFilter(42), XorFilter(42))
}

func TestPipeFilter(t *testing.T) {
	testFilter(t, PipeFilter([]string{"cat"}), PipeFilter([]string{"cat"}))
}
