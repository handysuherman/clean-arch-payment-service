package serializer

import (
	"io"

	"github.com/bytedance/sonic"
)

var json = sonic.ConfigStd

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func NewDecoder(r io.Reader) sonic.Decoder {
	return json.NewDecoder(r)
}

func NewEncoder(w io.Writer) sonic.Encoder {
	return json.NewEncoder(w)
}
