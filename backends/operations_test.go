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
