package codec

import (
	"encoding/json"
)

// JSONCodec --
type JSONCodec struct{}

// Encode --
func (j *JSONCodec) Encode(value interface{}) ([]byte, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Decode --
func (j *JSONCodec) Decode(data []byte) (interface{}, error) {
	return data, nil
}
