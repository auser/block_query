package json_scanner

import (
	"fmt"
	"testing"

	"github.com/auser/block_query/scanner/json_scanner"
)

func TestJsonParsing(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    `{"hello": "world"}`,
			expected: "hello world",
		},
	}

	for _, c := range cases {
		o, err := json_scanner.Parse(c.input, "")
		if err != nil {
			t.Error(err)
		}

		fmt.Printf("O: %#v\n", o)
		t.Fail()
	}
}
