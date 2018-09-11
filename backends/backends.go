package backends

import (
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"
)

var (
	errUnexpectedEOF     = errors.New("unexpected EOF")
	errOutOfRange        = errors.New("Out of range")
	errNotFound          = errors.New("Not found")
	errUnhandledDatatype = errors.New("Unhandled datatype")
	//
	errKeyValueNotEqual              = errors.New("Key not equal")
	errNotRegexpSupported            = errors.New("Not a regexp")
	errKeyValueNotGreaterThan        = errors.New("Value not greater than")
	errKeyValueNotGreaterThanOrEqual = errors.New("Value not greater than or equal")
	errKeyValueNotLessThan           = errors.New("Value not less than")
	errKeyValueNotLessThanOrEqual    = errors.New("Value not less than or equal")
)

// Interface maps an anonymous interface
type Interface interface{}

// Op defines a single transformation to be applied to an Interface
type Op interface {
	Apply(Interface) (Interface, error)
}

// OpFunc provides a convenient func type wrapper on Op
type OpFunc func(Interface) (Interface, error)

// Apply executes the transformation defined by OpFunc
func (fn OpFunc) Apply(in Interface) (Interface, error) {
	return fn(in)
}

// MatchOp defines a matching operation
type MatchOp interface {
	Apply(Interface) (Interface, error)
}

// MatchFunc is a matcher operation
type MatchFunc func(Interface) (Interface, error)

// ComparisonFunc defines a comparison operation
type ComparisonFunc func(a float64, b float64) bool

// Apply executes the comparison operation
func (fn ComparisonFunc) Apply(a float64, b float64) bool {
	return fn(a, b)
}

// Chain executes a series of operations
func Chain(filters ...Op) OpFunc {
	return func(in Interface) (Interface, error) {
		if filters == nil {
			return in, nil
		}

		var err error
		data := in
		for _, filter := range filters {
			data, err = filter.Apply(data)
			if err != nil {
				return nil, err
			}
		}

		return data, nil
	}
}

// Filter executes a series of filters on an Interface
func Filter(filters ...MatchOp) OpFunc {
	return func(in Interface) (Interface, error) {
		if filters == nil {
			return in, nil
		}

		data := in

		for _, filter := range filters {
			val, err := filter.Apply(data)
			if err == nil {
				data = val
			} else {
				data = make(map[string]interface{}, 0)
				break
			}
			// for key := range in.(map[string]interface{}) {
			// 	data, _ = applyAndAppend(filter, key, in, data)
			// 	// 		switch val.(type) {
			// 	// 		case []interface{}:
			// 	// 			for _, nVal := range val {
			// 	// 				res, err = filter(nVal)
			// 	// 				data = append(data, applyAndAppend())
			// 	// 			}
			// 	// 		}
			// }
		}

		return data, nil
	}
}

// skipWhitespace skips whitespace
func skipWhitespace(in []byte, pos int) (int, error) {
	for {
		r, size := utf8.DecodeRune(in[pos:])
		if size == 0 {
			return 0, errUnexpectedEOF
		}
		if !unicode.IsSpace(r) {
			break
		}
		pos += size
	}
	return pos, nil
}

func newError(pos int, b byte) error {
	return fmt.Errorf("Invalid character at position: %v; %v", pos, string([]byte{b}))
}
