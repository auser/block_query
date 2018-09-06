package backends

import (
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"
)

var (
	errUnexpectedEOF = errors.New("unexpected EOF")
)

type Interface interface{}

// Op defines a single transformation to be applied to a []byte
type Op interface {
	Apply(Interface) (Interface, error)
}

// OpFunc provides a convenient func type wrapper on Op
type OpFunc func(Interface) (Interface, error)

// Apply executes the transformation defined by OpFunc
func (fn OpFunc) Apply(in Interface) (Interface, error) {
	return fn(in)
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
