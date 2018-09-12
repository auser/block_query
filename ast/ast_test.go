package ast

import (
	"reflect"
	"testing"
)

func TestNewAstSlices(t *testing.T) {
	// Empty
	cases := []struct {
		src    interface{}
		output []interface{}
		want   bool
	}{
		{
			src:    []interface{}{1, 2, 3},
			output: []interface{}{1, 2, 3},
			want:   true,
		},
		{
			src:    []interface{}{1},
			output: []interface{}{1},
			want:   true,
		},
		{
			src:    []interface{}{1, 2, 3},
			output: []interface{}{1, 2},
			want:   false,
		},
	}

	for idx, tt := range cases {
		ast := NewAST(tt.src)
		res := make([]int, 0)

		ast.ToSlice(&res)
		if tt.want {
			if reflect.DeepEqual(tt.output, res) {
				t.Errorf("NewAST[%d] failed. expected=%v, got=%v", idx, tt.output, res)
			}
		} else {
			if reflect.DeepEqual(tt.output, res) {
				t.Errorf("NewAST[%d] should not equal, but they do. Expected=%q, got=%q", idx, tt.output, res)
			}
		}
	}
}

func TestNewAstMaps(t *testing.T) {
	// Empty
	// cases := []struct {
	// 	src    map[string]interface{}
	// 	output []interface{}
	// }{
	// 	{
	// 		src:    map[string]interface{}{"name": "world", "msg": "world"},
	// 		output: []interface{}{1, 2, 3},
	// 	},
	// }

	// for idx, tt := range cases {
	t.Skip()
	// }
}

func slicesAreEqual(a []interface{}, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for idx, aval := range a {
		if aval != b[idx] {
			return false
		}
	}

	return true
}
