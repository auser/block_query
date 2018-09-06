package json_scanner

import s "github.com/auser/block_query/scanner"

var n = []byte("null")

// Null verifies the contents of bytes provided is a null as pos
func Null(in []byte, pos int) (int, error) {
	switch in[pos] {
	case 'n':
		return s.Expect(in, pos, n...)
		return pos + 4, nil
	default:
		return 0, s.ErrUnexpectedValue
	}
}
