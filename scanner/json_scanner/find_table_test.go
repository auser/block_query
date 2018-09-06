package json_scanner_test

import (
	"fmt"
	"testing"

	"github.com/auser/block_query/scanner/json_scanner"

	u "github.com/auser/block_query/utils"
)

func TestJSONFindTable(t *testing.T) {
	data, err := u.ReadFixture("pets.json")
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		in       string
		key      string
		expected string
	}{
		{
			in:       string(data),
			key:      "pets",
			expected: "dogs",
		},
	}

	for _, tt := range tests {
		table, err := json_scanner.FindTable([]byte(tt.in), 0, []byte(tt.key))

		if err != nil {
			t.Error(err)
		}

		fmt.Printf("table: %q\n", table)
	}
}
