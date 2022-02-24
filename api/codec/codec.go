package codec

type Codec struct{}

func (c *Codec) Encode(src interface{}) ([]byte, error) {
	enc := Encoder{}
	return enc.Marshall(src)
}

func (c *Codec) Decode(data []byte, dest interface{}) error {
	dec := Decoder{}
	return dec.Unmarshall(data, dest)
}

func (c *Codec) DecodeAsInterface(src interface{}, dest interface{}) error {
	dec := Decoder{}
	return dec.UnmarshallFromInterface(src, dest)
}
