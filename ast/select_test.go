package ast

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestSelect(t *testing.T) {
	cases := []struct {
		src      interface{}
		selector func(interface{}) interface{}
		output   []interface{}
	}{
		{
			src: map[string]string{"hello": "world"},
			selector: func(i interface{}) interface{} {
				val := i.(KeyValue).Value
				return strings.ToUpper(val.(string))
			},
			output: []interface{}{"WORLD"},
		},
	}

	for idx, tt := range cases {
		ast := NewAST(tt.src).Select(tt.selector)
		res := make([]interface{}, 0)
		ast.ToSlice(&res)

		if !reflect.DeepEqual(tt.output, res) {
			t.Errorf("test[%d] did not equal expected output. Expected=%q, got=%q", idx, tt.output, res)
		}
		fmt.Printf("ast: %#v\n", res)
	}
}

func TestSelectKeys(t *testing.T) {
	cases := []struct {
		src        interface{}
		selectKeys []string
		output     map[string]interface{}
	}{
		{
			src:        map[string]string{"hello": "world"},
			selectKeys: []string{"hello"},
			output:     map[string]interface{}{"hello": "world"},
		},
		{
			src: map[string]interface{}{
				"name":      "ari",
				"job_title": "co-founder",
				"pets": map[string]interface{}{
					"dogs": []string{
						"Ginger",
					},
				},
			},
			selectKeys: []string{"name", "job_title"},
			output: map[string]interface{}{
				"name":      "ari",
				"job_title": "co-founder",
			},
		},
	}

	for idx, tt := range cases {
		ast := NewAST(tt.src).SelectKeys(tt.selectKeys...)
		res := make(map[string]interface{})
		ast.ToMap(&res)

		if !reflect.DeepEqual(tt.output, res) {
			t.Errorf("test[%d] did not equal expected output. Expected=%q, got=%q", idx, tt.output, res)
		}
	}
}
