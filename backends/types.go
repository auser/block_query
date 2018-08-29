package backends

import (
	"fmt"
	"strings"
)

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

	str = fmt.Sprintf("%s objs\n", str)
	for _, obj := range j.objects {
		str = fmt.Sprintf("%s\t%s\n", str, obj.String())
	}

	str = fmt.Sprintf("%s arrays\n", str)
	for _, arr := range j.arrays {
		str = fmt.Sprintf("%s\t%s\n", str, arr.String())
	}

	return str
}

type JSONObject struct {
	Pairs []*JSONPair
}

func (j *JSONObject) String() string {
	str := ""
	for _, pair := range j.Pairs {
		str = fmt.Sprintf("%s\n%s", str, pair.String())
	}
	return str
}

type JSONArray struct {
	Values []interface{} // strings, ints, floats, mixed, etc.
}

func (j *JSONArray) String() string {
	valStrs := []string{}
	for _, v := range j.Values {
		fmt.Printf("V: %s\n", v)
		valStrs = append(valStrs, fmt.Sprintf("%s", v))
	}
	str := strings.Join(valStrs, ",")
	return str
}

type JSONPair struct {
	Key   string
	Value interface{} // string / int / float / ...etc
}

func (j *JSONPair) String() string {
	return fmt.Sprintf("\t%s: %v", j.Key, j.Value)
}

type JSONNumber struct {
	Value float64
}
