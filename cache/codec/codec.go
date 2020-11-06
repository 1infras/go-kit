package codec

// ICodec --
type ICodec interface {
	Encode(value interface{}) ([]byte, error)
	Decode(data []byte, value interface{}) error
}
