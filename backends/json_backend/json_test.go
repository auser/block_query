package json_backend

import (
	"reflect"
	"testing"

	"github.com/auser/block_query/backends/json_backend"
)

type i map[string]interface{}

func TestJsonParsing(t *testing.T) {
	cases := []struct {
		input    string
		expected map[string]interface{}
	}{
		{
			input:    `{"hello": "world"}`,
			expected: i{"hello": "world"},
		},
		{
			input:    `{"hello": 10}`,
			expected: i{"hello": float64(10)},
		},
		{
			input:    `{"hello": null}`,
			expected: i{"hello": interface{}(nil)},
		},
		{
			input:    `{"name": "ari", "pets": []}`,
			expected: i{"name": "ari", "pets": []interface{}{}},
		},
		{
			input:    `{"name": "ari", "pets": {}}`,
			expected: i{"name": "ari", "pets": map[string]interface{}{}},
		},
		{
			input: `{"name": "ari", "pets": {
				"dogs": ["Ginger"]
			}}`,
			expected: i{"name": "ari", "pets": map[string]interface{}{"dogs": []interface{}{"Ginger"}}},
		},
	}

	for i, c := range cases {
		o, err := json_backend.Parse("filename", []byte(c.input))
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(o.(map[string]interface{}), c.expected) {
			t.Errorf("test[%d] got unexpected result. expected=%q, got=%q\n", i, c.expected, o)
		}
	}
}
