package json_scanner

import s "github.com/auser/block_query/scanner"

func Object(in []byte, pos int) (int, error) {
	pos = s.Must(s.SkipSpace(in, pos))

	if v := in[pos]; v != '{' {
		return 0, s.NewError(pos, v)
	}
	pos++

	// Clean initial space
	pos = s.Must(s.SkipSpace(in, pos))
	if in[pos] == '}' {
		return pos + 1, nil
	}

	for {

	}

	return 0, nil
}
