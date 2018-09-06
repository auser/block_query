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

		return in, nil
	}
}

func FindIndex(idx int) OpFunc {
	return func(in Interface) (Interface, error) {
		arr := in.([]interface{})

		if idx > len(arr) {
			return in, errOutOfRange
		}

		return arr[idx], nil
	}
}
