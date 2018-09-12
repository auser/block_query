package backend

// Iterator is a function for iterating over data
type Iterator func() (item interface{}, ok bool)

// Backend is a struct that is used for
type Backend struct {
	Iterate func() Iterator
}
