package backends

import (
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
