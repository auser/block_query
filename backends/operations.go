package backends

import (
	"strings"
)

func FindKey(key string) OpFunc {
	return func(in Interface) (Interface, error) {

		for k, v := range in.(map[string]interface{}) {
			if strings.Compare(k, key) == 0 {
				return v, nil
			}
		}
		// pos, err := skipWhitespace(in, 0)
		// if err != nil {
		// 	return nil, err
		// }

		// if v := in[pos]; v != '{' {
		// 	return nil, newError(pos, v)
		// }

		// We're in an object now

		return in, nil
	}
}
