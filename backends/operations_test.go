package backends

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/auser/block_query/backends/json_backend"

	u "github.com/auser/block_query/utils"
)

func getParser(t *testing.T) Interface {
	data := u.MustBytes(u.ReadFixture("1.json"))
	parser, err := json_backend.Parse("", data)
	if err != nil {
		t.FailNow()
	}
	return parser
}

func TestFindKey_Exists(t *testing.T) {
	parser := getParser(t)
	f := FindKey("transactions")

	output := u.MustInterface(f(parser.(map[string]interface{})))

	if output == nil {
		t.Errorf("Expected output, but got none: %v\n", output)
	}

	first := output.([]interface{})[0]
	txID := first.(map[string]interface{})["transactionId"]

	if txID != "123456" {
		t.Errorf("Incorrect transactionID: %#v\n", txID)
	}
}

func TestFindKey_NotExists(t *testing.T) {
	parser := getParser(t)
	f := FindKey("blankenships")
	output := u.MustInterface(f(parser.(map[string]interface{})))

	if !reflect.DeepEqual(output.(map[string]interface{}), parser.(map[string]interface{})) {
		t.Errorf("Expecting full response back, but got something different: %v\n", output)
	}
}

func TestFindKey_IncorrectDatatype(t *testing.T) {
	t.Skip("Pending")
}

func TestFindIndex_Exists(t *testing.T) {
	parser := getParser(t)
	getArr := FindKey("transactions")
	arr := u.MustInterface(getArr(parser.(map[string]interface{})))

	f := FindIndex(0)
	output := u.MustInterface(f(arr))
	data := output.(map[string]interface{})

	if data["id"] != float64(1) || data["transactionId"] != "123456" {
		t.Errorf("Fetched incorrect array element\n")
	}
}

func TestFindIndex_NotExists(t *testing.T) {
	parser := getParser(t)
	getArr := FindKey("transactions")
	arr := u.MustInterface(getArr(parser.(map[string]interface{})))

	f := FindIndex(40)
	_, err := f(arr)

	if err != errOutOfRange {
		t.Errorf("Expected out of range error, but got something else")
	}
}

// Match tests
func TestContainsKey_Exists(t *testing.T) {
	parser := getParser(t)
	f := FindKey("accounts")
	output := u.MustInterface(f(parser.(map[string]interface{})))

	f = ContainsKey("ari")
	filter := Filter(f)
	key, err := filter(output)

	if key == nil {
		t.Errorf("Expected to find 'ari' key, but found none: %#v.\n", err)
	}
}

func TestContainsKey_NotExists(t *testing.T) {
	parser := getParser(t)
	f := FindKey("accounts")
	output := u.MustInterface(f(parser.(map[string]interface{})))

	f = ContainsKey("bob")
	filter := Filter(f)
	key, err := filter(output)

	if len(key.(map[string]interface{})) != 0 {
		t.Errorf("Expected not to find 'bob' key, but found one: %#v.\n", err)
	}
}

func TestContainsKeyEqualTo(t *testing.T) {
	parser := getParser(t)

	filter := Filter(ContainsKeyEqualTo("blockNumber", float64(1)))
	data, err := filter(parser.(map[string]interface{}))

	keys := allKeys(data)
	fmt.Printf("key: %#v\n", keys)
	if len(keys) == 1 {
		t.Errorf("Key blockNumber was expected, but got an error: %#v\n", err)
	}

	filter = Filter(ContainsKeyEqualTo("badMan", "doc"))
	data, _ = filter(parser.(map[string]interface{}))

	if len(allKeys(data)) != 0 {
		t.Errorf("Expecting value not equal, but got a different error: %#v\n", len(allKeys(data)))
	}

	filter = Filter(ContainsKeyEqualTo("blockNumber", "doc"))
	data, _ = filter(parser.(map[string]interface{}))

	if len(allKeys(data)) != 0 {
		t.Errorf("Expecting value not equal, but got a different error: %#v\n", err)
	}
}

func TestContainsKeyLike(t *testing.T) {
	parser := getParser(t)

	filter := Filter(ContainsKeyLike("blockchain", "eth*"))
	data, err := filter(parser.(map[string]interface{}))
	keys := allKeys(data)

	if len(keys) == 1 {
		t.Errorf("Key blockchain was expected, but got an error: %#v\n", err)
	}

	filter = Filter(ContainsKeyLike("blockchain", "eo"))
	data, err = filter(parser.(map[string]interface{}))
	keys = allKeys(data)

	if len(keys) != 0 {
		t.Errorf("expected key not equal, but got different error: %#v\n", err)
	}

	filter = Filter(ContainsKeyLike("blockchain", "something else"))
	data, err = filter(parser.(map[string]interface{}))
	keys = allKeys(data)

	if len(keys) != 0 {
		t.Errorf("expected key not equal, but got different error: %#v\n", err)
	}
}

func TestFetchKeys(t *testing.T) {
	parser := getParser(t)

	getArr := FindKey("transactions")
	arr := u.MustInterface(getArr(parser.(map[string]interface{})))

	f := FetchKeys("id")
	output := u.MustInterface(f(arr)).([]Interface)

	for idx := range output {
		if len(allKeys(output[idx])) != 1 {
			t.Errorf("output included more than just a 'id' when it should not have: %#v\n", output[idx])
		}
	}
}

func TestComparisonOperators(t *testing.T) {
	parser := getParser(t)

	filter := Filter(ContainsKeyGreaterThan("someRandomNumber", float64(1)))

	origLen := len(parser.(map[string]interface{}))
	data, _ := filter(parser.(map[string]interface{}))
	keys := allKeys(data)

	if len(keys) != origLen {
		t.Errorf("Expected someRandomNumber to be greater than 1, but it was not: %#v\n", keys)
	}

	filter = Filter(ContainsKeyGreaterThan("somethingElse", 234567))
	data, _ = filter(parser.(map[string]interface{}))
	keys = allKeys(data)

	if len(keys) != 0 {
		t.Errorf("Expected somethingElse not to be greater than 234567, but it was: %#v\n", keys)
	}

	// GTE
	filter = Filter(ContainsKeyGreaterThanOrEqual("someRandomNumber", 123456))
	data, _ = filter(parser.(map[string]interface{}))
	keys = allKeys(data)

	if len(keys) != origLen {
		t.Errorf("Expected someRandomNumber to be greater than or equal to 123456, but it was not: %#v\n", keys)
	}

	filter = Filter(ContainsKeyGreaterThanOrEqual("someRandomNumber", 123457))
	data, _ = filter(parser.(map[string]interface{}))
	keys = allKeys(data)

	if len(keys) != 0 {
		t.Errorf("Expected someRandomNumber to be greater than or equal to 123456, but it was not: %#v\n", keys)
	}

	// LT
	filter = Filter(ContainsKeyLessThan("someRandomNumber", 234578))
	data, _ = filter(parser.(map[string]interface{}))
	keys = allKeys(data)

	if len(keys) != origLen {
		t.Errorf("Expected someRandomNumber to be less than 23478, but it was not: %#v\n", keys)
	}

	filter = Filter(ContainsKeyLessThan("someRandomNumber", 1))
	data, _ = filter(parser.(map[string]interface{}))
	keys = allKeys(data)

	if len(keys) != 0 {
		t.Errorf("Expected someRandomNumber to be less than 1, but it was not: %#v\n", keys)
	}

	// LTE
	filter = Filter(ContainsKeyLessThanOrEqual("someRandomNumber", 123456))
	data, _ = filter(parser.(map[string]interface{}))
	keys = allKeys(data)

	if len(keys) != origLen {
		t.Errorf("Expected someRandomNumber to be less than or equal to 123456, but it was not: %#v\n", keys)
	}

	filter = Filter(ContainsKeyLessThan("someRandomNumber", 1))
	data, _ = filter(parser.(map[string]interface{}))
	keys = allKeys(data)

	if len(keys) != 0 {
		t.Errorf("Expected someRandomNumber to be less than 1, but it was not: %#v\n", keys)
	}
}

func allKeys(data Interface) []string {
	keys := make([]string, 0)
	for key := range data.(map[string]interface{}) {
		keys = append(keys, key)
	}
	return keys
}
