package util

import (
	"bytes"
	"encoding/gob"
)

func EncodeMap(m *map[string]interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(m)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeMap(buf []byte) (map[string]interface{}, error) {
	dec := gob.NewDecoder(bytes.NewReader(buf))
	var m map[string]interface{}
	err := dec.Decode(&m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
