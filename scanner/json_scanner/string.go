package json_scanner

import (
	"errors"

	"github.com/auser/block_query/scanner"
)

func String(in []byte, pos int) (int, error) {
	pos, err := scanner.SkipSpace(in, pos)
	if err != nil {
		return 0, err
	}

	max := len(in)
	if v := in[pos]; v != '"' {
		return 0, scanner.NewError(pos, v)
	}
	pos++

	for {
		switch in[pos] {
		case '\\':
			if in[pos+1] == '"' {
				pos++
			}
		case '"':
			return pos + 1, nil
		}
		pos++

		if pos >= max {
			break
		}
	}

	return 0, errors.New("unclosed string")
}
