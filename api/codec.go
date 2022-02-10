package api

import (
	"bytes"
	"encoding/json"
)

// TODO: To implement this from scratch after completing functions
type Codec struct{}

func (c *Codec) Encode(src interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(src)
	if err != nil {
		return []byte{}, nil
	}

	return buf.Bytes(), nil
}

func (c *Codec) Decode(data []byte, dest interface{}) error {
	buf := bytes.NewBuffer(data)
	err := json.NewDecoder(buf).Decode(dest)
	return err
}
