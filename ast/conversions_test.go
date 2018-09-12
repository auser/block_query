package ast

import (
	"reflect"
	"testing"
)

func TestToSlice(t *testing.T) {
	cases := []struct {
		src    []int
		output []interface{}
	}{
		{
			src:    []int{1, 2, 3},
			output: []interface{}{1, 2, 3},
		},
		{
			src:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			output: []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
	}

	result := make([]int, 0)
	for idx, tt := range cases {
		ast := NewAST(tt.src)
		ast.ToSlice(&result)
		if reflect.DeepEqual(result, tt.output) {
			t.Errorf("ToSlice test[%d] failed. Expected=%q, got=%q", idx, tt.output, result)
		}
	}
}

func TestToMap(t *testing.T) {
	cases := []struct {
		src interface{}
	}{
		{
			src: map[string]interface{}{"hello": "world"},
		},
		{
			src: map[string]interface{}{
				"name": "ari",
				"pets": map[string]interface{}{
					"dogs": []string{"Ginger"},
				},
			},
		},
	}

	for idx, tt := range cases {
		ast := NewAST(tt.src)
		res := make(map[string]interface{})
		ast.ToMap(&res)

		if !reflect.DeepEqual(res, tt.src) {
			t.Errorf("Test[%d] Expected the resulting values to equal. expected=%q, got=%q", idx, tt.src, res)
		}
	}
}
