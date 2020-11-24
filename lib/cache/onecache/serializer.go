package onecache

import (
	"encoding/json"
)

type Serializer interface {
	Encode(value interface{}) ([]byte, error)
	Decode(data []byte, object interface{}) error
}

type DefaultSerializer struct{}

func (_this *DefaultSerializer) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (_this *DefaultSerializer) Decode(data []byte, object interface{}) error {
	return json.Unmarshal(data, object)
}
