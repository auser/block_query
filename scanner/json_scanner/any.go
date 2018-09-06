package json_scanner

import s "github.com/auser/block_query/scanner"

func Any(in []byte, pos int) (int, error) {
	pos = s.Must(s.SkipSpace(in, pos))

	switch in[pos] {
	case '"':
		return String(in, pos)
	// case "{":
	// 	return Object(in, pos)
	case '.', '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		return Number(in, pos)
	// case '[':
	// return Array(in, pos)
	// case 't', 'f':
	// return Boolean(in, pos)
	case 'n':
		return Null(in, pos)
	default:
		max := len(in) - pos
		if max > 20 {
			max = 20
		}

		return 0, s.OpErr{
			Pos:     pos,
			Msg:     "invalid object",
			Content: string(in[pos : pos+max]),
		}
	}
}
