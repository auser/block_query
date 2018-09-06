package json_scanner

import (
	"errors"
	"fmt"

	s "github.com/auser/block_query/scanner"
)

func FindTable(in []byte, pos int, name []byte) ([]byte, error) {
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

		fmt.Printf("pos: %#v %s\n", pos, string(in[pos]))
		pos++
	}

	return nil, errors.New("Silly")
}
