package utils

import "fmt"

// Must requires the value exists or returns an error
func MustInt(val int, err error) int {
	if err != nil {
		panic(fmt.Errorf("scanner error; %v", err.Error()))
	}

	return val
}

func MustInterface(val interface{}, err error) interface{} {
	if err != nil {
		panic(fmt.Errorf("scanner error; %v", err.Error()))
	}

	return val
}

func MustBytes(val []byte, err error) []byte {
	if err != nil {
		panic(fmt.Errorf("Error: %v", err.Error()))
	}
	return val
}
