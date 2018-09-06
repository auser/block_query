package json_scanner

import (
	"testing"

	"github.com/auser/block_query/scanner/json_scanner"

	u "github.com/auser/block_query/utils"
)

func TestJsonScanning(t *testing.T) {
	data, err := u.ReadFixture("pets.json")
	if err != nil {
		t.Error(err)
	}

	parser := json_scanner.NewJSONParser(data)
	parser.Parse()
	t.Fail()
}
