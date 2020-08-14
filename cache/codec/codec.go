package codec

import (
	"encoding/json"
	"fmt"

	"github.com/1infras/go-kit/logger"
)

// ICodec --
type ICodec interface {
	Encode(value interface{}) ([]byte, error)
	Decode(data []byte) (interface{}, error)
}

// JSONCodec --
type JSONCodec struct{}

// Encode --
func (j *JSONCodec) Encode(value interface{}) ([]byte, error) {
	b, err := json.Marshal(value)
	if err != nil {
		logger.Errorw(fmt.Sprintf("Failed Encoding message %v", err))
		return nil, err
	}

	return b, nil
}

// Decode --
func (j *JSONCodec) Decode(data []byte) (interface{}, error) {
	var item interface{}
	err := json.Unmarshal(data, &item)
	if err != nil {
		return nil, err
	}
	return item, nil
}
