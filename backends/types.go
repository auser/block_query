package backends

import "fmt"

type JSON struct {
	nums    []*JSONNumber
	objects []*JSONObject
	arrays  []*JSONArray
	depth   int
}

func (j *JSON) String() string {
	str := ""

	str = fmt.Sprintf("%s nums:\n", str)
	for _, num := range j.nums {
		str = fmt.Sprintf("%s\t%d\n", str, num.Value)
	}

	return str
}

type JSONObject struct {
	Pairs []*JSONPair
}

type JSONArray struct {
	Values []interface{} // strings, ints, floats, mixed, etc.
}

type JSONPair struct {
	Key   string
	Value interface{} // string / int / float / ...etc
}

type JSONNumber struct {
	Value float64
}

func (j *JSON) addJson(json string) {
	fmt.Printf("JSON: -------------------> %v\n", json)
}
