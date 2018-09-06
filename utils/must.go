package utils

import "fmt"

// Must requires the value exists or returns an error
func Must(val int, err error) int {
	if err != nil {
		panic(fmt.Errorf("scanner error; %v", err.Error()))
	}

	return val
}
