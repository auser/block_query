package backends

import (
	"errors"
	"regexp"
	"strings"
)

// FindKey finds a key by an interface
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

// FindIndex finds the index of a specific entry in an array
func FindIndex(idx int) OpFunc {
	return func(in Interface) (Interface, error) {
		arr := in.([]interface{})

		if idx > len(arr) {
			return in, errOutOfRange
		}

		return arr[idx], nil
	}
}

// FetchKeys fetchs only the keys requested
func FetchKeys(keys ...string) OpFunc {
	return func(in Interface) (Interface, error) {
		arr := in.([]interface{})
		ret := make([]Interface, 0)

		fetchKeys := func(in Interface) Interface {
			data := make(map[string]interface{}, len(keys))
			for _, key := range keys {
				val := in.(map[string]interface{})[key]
				data[key] = val
			}
			return data
		}

		for _, val := range arr {
			ret = append(ret, fetchKeys(val).(Interface))
		}
		return ret, nil
	}
}

// Limit limits the number of return values
func Limit(n int) OpFunc {
	return func(in Interface) (Interface, error) {
		// TODO
		return in, nil
	}
}

/******************************************
 * Matching operations (where)
 ******************************************/

// ContainsKey checks to see if a map contains a key
func ContainsKey(key string) OpFunc {
	return func(in Interface) (Interface, error) {
		if _, ok := in.(map[string]interface{})[key]; ok {
			return in, nil
		} else {
			return nil, errNotFound
		}
	}
}

// ContainsKeyEqualTo checks to see if a key equals a value
func ContainsKeyEqualTo(key string, value interface{}) OpFunc {
	return func(in Interface) (Interface, error) {
		if val, ok := in.(map[string]interface{})[key]; ok {
			if val == value {
				return in, nil
			} else {
				return nil, errKeyValueNotEqual
			}
		} else {
			return nil, errNotFound
		}
	}
}

// ContainsKeyLike checks to see if there is a likeness
// equality comparison in an interface
func ContainsKeyLike(key string, value string) OpFunc {
	return func(in Interface) (Interface, error) {
		if val, ok := in.(map[string]interface{})[key]; ok {
			r, err := regexp.Compile(value)
			if err != nil {
				return nil, errNotRegexpSupported
			}
			if r.MatchString(val.(string)) {
				return in, nil
			} else {
				return nil, errKeyValueNotEqual
			}
		} else {
			return nil, errNotFound
		}
	}
}

// ContainsKeyGreaterThan checks to see if the value is greater than
func ContainsKeyGreaterThan(key string, value float64) OpFunc {
	return containsKeyWithOp(key, value, func(a float64, b float64) bool {
		return a > b
	}, "value not greater than")
}

// ContainsKeyGreaterThanOrEqual checks to see if the value is greater than or equal to a value
func ContainsKeyGreaterThanOrEqual(key string, value float64) OpFunc {
	return containsKeyWithOp(key, value, func(a float64, b float64) bool {
		return a >= b
	}, "value not greater than or equal")
}

// ContainsKeyLessThan checks to see if the value is less than a value
func ContainsKeyLessThan(key string, value float64) OpFunc {
	return containsKeyWithOp(key, value, func(a float64, b float64) bool {
		return a < b
	}, "value not less than")
}

// ContainsKeyLessThanOrEqual checks to see if the value is less than a value
func ContainsKeyLessThanOrEqual(key string, value float64) OpFunc {
	return containsKeyWithOp(key, value, func(a float64, b float64) bool {
		return a <= b
	}, "value not less than or equal")
}

func containsKeyWithOp(key string, value float64, op ComparisonFunc, errStr string) OpFunc {
	return func(in Interface) (Interface, error) {
		if val, ok := in.(map[string]interface{})[key]; ok {
			res := op.Apply(val.(float64), value)
			if res == true {
				return in, nil
			} else {
				return nil, errors.New(errStr)
			}
		} else {
			return nil, errNotFound
		}
	}
}
