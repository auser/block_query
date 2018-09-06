package json_scanner

import s "github.com/auser/block_query/scanner"

// FindIndex finds an element in position of an array
func FindIndex(in []byte, pos, index int) ([]byte, error) {
	pos = s.Must(s.SkipSpace(in, pos))

	if v := in[pos]; v != '[' {
		return nil, s.NewError(pos, v)
	}
	pos++

	idx := 0
	for {
		pos = s.Must(s.SkipSpace(in, pos))

		itemStart := pos
		// data
		pos = s.Must(Any(in, pos))

		if index == idx {
			return in[itemStart:pos], nil
		}

		pos = s.Must(s.SkipSpace(in, pos))
		switch in[pos] {
		case ',':
			pos++
		case ']':
			return nil, s.ErrIndexOutOfBounds
		}
		idx++
	}
}
