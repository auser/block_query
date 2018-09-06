package json_backend

import (
	"bytes"
	"fmt"
)

type JSONBackend struct {
	json map[string]interface{}
}

// NewJSONBackend creates a new backend
func NewJSONBackend(b []byte) (*JSONBackend, error) {
	parsed, err := Parse("", b)
	if err != nil {
		return nil, err
	}

	backend := &JSONBackend{
		json: parsed.(map[string]interface{}),
	}

	return backend, nil
}

func (b *JSONBackend) Query(str string) {}

func (b *JSONBackend) FindKey(key string) int {
	for k, v := range b.json {
		if bytes.Compare([]byte(k), []byte(key)) == 0 {
			fmt.Printf("%s: %v\n", k, v)
		}
	}

	return -1
}
