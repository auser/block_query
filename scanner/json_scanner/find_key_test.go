package json_scanner_test

import (
	"testing"

	"github.com/auser/block_query/scanner/json_scanner"
)

func TestJSONFindKey(t *testing.T) {
	tests := []struct {
		in       string
		key      string
		expected string
		startAt  int
	}{
		{
			in:       `{"hello":"World"}`,
			key:      "hello",
			startAt:  0,
			expected: `"World"`,
		},
		{
			in:       `{ "hello" : "Ari" }`,
			key:      "hello",
			startAt:  0,
			expected: `"Ari"`,
		},
		{
			in:       `{ "name" : "   Ari   " }`,
			key:      "name",
			startAt:  0,
			expected: `"   Ari   "`,
		},
	}

	for i, tt := range tests {
		data, err := json_scanner.FindKey([]byte(tt.in), tt.startAt, []byte(tt.key))
		if err != nil {
			t.Error(err)
		}

		if string(data) != tt.expected {
			t.Errorf("test[%d] got unexpected result. expected=%q, got=%q\n", i, tt.expected, string(data))
		}
	}
}
