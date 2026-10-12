package codec

import "encoding/json"

type jsonCodec[T any] struct{}

func NewCodecJson[T any]() Codec[T] {
	return jsonCodec[T]{}
}

func (jsonCodec[T]) Encode(v *T) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec[T]) Decode(data []byte) (*T, error) {
	var value T

	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}

	return &value, nil
}
