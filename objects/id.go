package objects

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"hash"
	"io"
	"strings"
)

type ObjectIdAlgo string

const (
	OIdAlgoSHA3_256 ObjectIdAlgo = "sha3-256"
	OIdAlgoDefault               = OIdAlgoSHA3_256
)

var allowedAlgos = map[ObjectIdAlgo]struct{}{OIdAlgoSHA3_256: {}}

func (algo ObjectIdAlgo) checkAlgo() bool {
	_, ok := allowedAlgos[algo]
	return ok
}

func (algo ObjectIdAlgo) sumLength() int {
	if algo != OIdAlgoSHA3_256 {
		panic("Only sha3-256 is implemented!")
	}

	return 32
}

// ObjectId identifies an object using a cryptogrpahic hash
type ObjectId struct {
	Algo ObjectIdAlgo
	Sum  []byte
}

func (oid ObjectId) Wellformed() bool {
	return oid.Algo.checkAlgo() && len(oid.Sum) == oid.Algo.sumLength()
}

func (oid ObjectId) String() string {
	return fmt.Sprintf("%s:%s", oid.Algo, hex.EncodeToString(oid.Sum))
}

func ParseObjectId(s string) (oid ObjectId, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		err = errors.New("Invalid ObjectId format")
		return
	}

	oid.Algo = ObjectIdAlgo(parts[0])

	oid.Sum, err = hex.DecodeString(parts[1])
	if err != nil {
		return
	}

	if !oid.Wellformed() {
		err = errors.New("Object ID is malformed")
	}

	return
}

// Set implements flag.Value for ObjectId
func (oid *ObjectId) Set(s string) (err error) {
	*oid, err = ParseObjectId(s)
	return
}

func MustParseObjectId(s string) ObjectId {
	id, err := ParseObjectId(s)
	if err != nil {
		panic(err)
	}
	return id
}

func (a ObjectId) Equals(b ObjectId) bool {
	return a.Algo == b.Algo && bytes.Equal(a.Sum, b.Sum)
}

// ObjectIdGenerator generates an ObjectId from the binary representation of an object
type ObjectIdGenerator interface {
	io.Writer // Must not fail
	GetId() ObjectId
}

func (algo ObjectIdAlgo) Generator() ObjectIdGenerator {
	if algo != OIdAlgoSHA3_256 {
		panic("Only sha3-256 is implemented!")
	}

	return hashObjectIdGenerator{
		algo: algo,
		Hash: sha3.New256(),
	}
}

type hashObjectIdGenerator struct {
	algo ObjectIdAlgo
	hash.Hash
}

func (h hashObjectIdGenerator) GetId() ObjectId {
	return ObjectId{
		Algo: h.algo,
		Sum:  h.Sum([]byte{}),
	}
}

func (oid ObjectId) VerifyObject(o RawObject) bool {
	gen := oid.Algo.Generator()
	if err := o.Serialize(gen); err != nil {
		panic(err)
	}

	return gen.GetId().Equals(oid)
}
