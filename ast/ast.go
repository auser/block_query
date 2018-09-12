package ast

import (
	"reflect"
)

// Iterator is a function for iterating over data
type Iterator func() (item interface{}, ok bool)

// AST is a struct that is used for
type AST struct {
	Iterate func() Iterator
}

// KeyValue is a struct for handling objects
type KeyValue struct {
	Key   interface{}
	Value interface{}
}

// Iterable is an interface for a source type
type Iterable interface {
	Iterate() Iterator
}

// NewAST creates a new AST interface
func NewAST(source interface{}) AST {
	src := reflect.ValueOf(source)

	switch src.Kind() {
	case reflect.Slice, reflect.Array:
		len := src.Len()

		return AST{
			Iterate: func() Iterator {
				idx := 0

				return func() (item interface{}, ok bool) {
					ok = idx < len
					if ok {
						item = src.Index(idx).Interface()
						idx++
					}

					return
				}
			},
		}
	case reflect.Map:
		len := src.Len()
		return AST{
			Iterate: func() Iterator {
				idx := 0
				keys := src.MapKeys()

				return func() (item interface{}, ok bool) {
					ok = idx < len
					if ok {
						key := keys[idx]
						item = KeyValue{
							Key:   key.Interface(),
							Value: src.MapIndex(key).Interface(),
						}
						idx++
					}
					return
				}
			},
		}
	default:
		return FromIterable(source.(Iterable))
	}
}

// FromIterable creates a new AST iterating over items
// in the iterate function
func FromIterable(source Iterable) AST {
	return AST{
		Iterate: source.Iterate,
	}
}
