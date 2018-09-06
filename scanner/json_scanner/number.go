package json_scanner

import s "github.com/auser/block_query/scanner"

func Number(in []byte, pos int) (int, error) {
	pos = s.Must(s.SkipSpace(in, pos))

	max := len(in)
	for {
		v := in[pos]
		switch v {
		case '-', '+', '.', 'e', 'E', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			pos++
		default:
			return pos, nil
		}

		if pos >= max {
			return pos, nil
		}
	}

	return pos, nil
}
