package ast

import (
	"reflect"
)

// ContainsKeyEqual includes an item if the key listed is contained
// within the KeyValue list
func (ast AST) ContainsKeyEqual(key string, val interface{}) AST {
	return AST{
		Iterate: func() Iterator {
			next := ast.Iterate()

			return func() (item interface{}, ok bool) {
				it, ok := next()

				if ok {
					in := it.(KeyValue).Value.(map[string]interface{})

					if value, ok := in[key]; ok {
						if reflect.DeepEqual(value, val) {
							item = it
						}
					}
				}

				return
			}
		},
	}
}
