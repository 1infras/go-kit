package codec

import (
	"encoding/json"
	"fmt"

	"github.com/1infras/go-kit/logger"
)

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
	return data, nil
}
