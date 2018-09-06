package json_scanner

import (
	"bytes"

	s "github.com/auser/block_query/scanner"
)

func FindKey(in []byte, pos int, k []byte) ([]byte, error) {
	pos, err := s.SkipSpace(in, pos)
	if err != nil {
		return nil, err
	}

	if v := in[pos]; v != '{' {
		return nil, s.NewError(pos, v)
	}
	pos++

	for {
		pos = s.Must(s.SkipSpace(in, pos))

		keyStart := pos
		pos := s.Must(String(in, pos))

		key := in[keyStart+1 : pos-1]
		match := bytes.Equal(k, key)

		pos = s.Must(s.SkipSpace(in, pos))

		// colon
		pos = s.Must(s.Expect(in, pos, ':'))

		pos = s.Must(s.SkipSpace(in, pos))

		valueStart := pos

		pos = s.Must(Any(in, pos))

		if match {
			return in[valueStart:pos], nil
		}

		pos = s.Must(s.SkipSpace(in, pos))

		switch in[pos] {
		case ',':
			pos++
		case '}':
			return nil, s.ErrKeyNotFound
		}
	}
}
