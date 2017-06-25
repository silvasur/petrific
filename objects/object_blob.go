package objects

type Blob []byte

func (b Blob) Type() ObjectType {
	return OTBlob
}

func (b Blob) Payload() []byte {
	return []byte(b)
}

func (b *Blob) FromPayload(bytes []byte) error {
	// TODO: perhaps it is better to copy the bytes?
	*b = bytes
	return nil
}
