package ast

import (
	"fmt"
	"reflect"
	"testing"
)

func TestContainsKeyEqual(t *testing.T) {
	cases := []struct {
		in  map[string]interface{}
		out map[string]interface{}
	}{
		{
			in: map[string]interface{}{
				"one": map[string]interface{}{
					"name": "ari",
				},
				"two": map[string]interface{}{
					"name": "zach",
				},
			},
			out: map[string]interface{}{
				"one": map[string]interface{}{
					"name": "ari",
				},
			},
		},
		{
			in: map[string]interface{}{
				"one": map[string]interface{}{
					"name": "ari",
				},
				"two": map[string]interface{}{
					"name": "zach",
				},
				"three": map[string]interface{}{
					"name":    "bob",
					"friends": []string{"ari", "zach"},
				},
			},
			out: map[string]interface{}{
				"one": map[string]interface{}{
					"name": "ari",
				},
			},
		},
	}

	for i, tt := range cases {
		ast := NewAST(tt.in).ContainsKeyEqual("name", "ari")
		res := make(map[string]interface{})
		ast.ToMap(&res)

		fmt.Printf("res: %#v\n", res)
		if !reflect.DeepEqual(res, tt.out) {
			t.Errorf("TestContainsKeyEqual[%d] failed. Should have received %q, but got %q\n", i, tt.out, res)
		}
	}
}
