package ast

import (
	"reflect"
)

// ToMap iterates over a collection and populates the resulting map
func (ast AST) ToMap(result interface{}) {
	keySelector := func(i interface{}) interface{} {
		return i.(KeyValue).Key
	}
	valueSelector := func(i interface{}) interface{} {
		return i.(KeyValue).Value
	}
	ast.ToMapBy(result, keySelector, valueSelector)
}

// ToMapBy iterates over a collection and populates a resulting map
// with the elements
func (ast AST) ToMapBy(result interface{},
	keySelector func(interface{}) interface{},
	valueSelector func(interface{}) interface{}) {
	res := reflect.ValueOf(result)
	temp := reflect.Indirect(res)
	next := ast.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		if item != nil {
			key := reflect.ValueOf(keySelector(item))
			val := reflect.ValueOf(valueSelector(item))

			temp.SetMapIndex(key, val)
		}
	}

	res.Elem().Set(temp)
}

// ToSlice iterates over a collection and saves the result to
// the pointer pointed by the input. It overwrites the existing slice
// starting from 0
func (ast AST) ToSlice(result interface{}) {
	res := reflect.ValueOf(result)
	slice := reflect.Indirect(res)

	cap := slice.Cap()
	res.Elem().Set(slice.Slice(0, cap))

	next := ast.Iterate()
	idx := 0

	for item, ok := next(); ok; item, ok = next() {
		if idx >= cap {
			slice, cap = grow(slice)
		}
		slice.Index(idx).Set(reflect.ValueOf(item))
		idx++
	}

	res.Elem().Set(slice.Slice(0, idx))
}

// grow grows the slice s by doubling its capacity, then it returns the new
// slice (resliced to its full capacity) and the new capacity.
func grow(s reflect.Value) (v reflect.Value, newCap int) {
	cap := s.Cap()
	if cap == 0 {
		cap = 1
	} else {
		cap *= 2
	}
	newSlice := reflect.MakeSlice(s.Type(), cap, cap)
	reflect.Copy(newSlice, s)
	return newSlice, cap
}
