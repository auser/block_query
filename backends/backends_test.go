package backends

import (
	"fmt"
	"testing"

	"github.com/auser/block_query/backends"
)

func TestJsonParsing_Test(t *testing.T) {
	tests := []struct {
		str           string
		expectedValue []string
	}{
		{
			str: `10`,
			expectedValue: []string{
				"10",
			},
		},
		{
			str: `{"hello": "world"}`,
			expectedValue: []string{
				"{", "hello", "world", "}",
			},
		},
		{
			str: `{"hello": "world", "valid": true}`,
			expectedValue: []string{
				"{", "hello", "world", "}",
			},
		},
		{
			str: `[1, 2, 3, 4]`,
			expectedValue: []string{
				"{", "hello", "world", "}",
			},
		},
	}

	for i, tt := range tests {
		t.Logf("str: %s\n", tt.str)
		out, err := backends.Parse("", []byte(tt.str))
		if err != nil {
			t.Errorf("tests[%d] Error parsing: %s", i, err.Error())
		}

		// for k, v := range out {
		// 	fmt.Printf("Key: %q, Value: %q\n", k, v)
		// }
		fmt.Printf("val: %v\n", out.(*backends.JSON).String())
		t.Logf("Out: %#v %#v\n", out, err)
		t.Error()
	}
}
