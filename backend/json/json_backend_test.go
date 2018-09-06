package json_backend

import (
	"testing"

	u "github.com/auser/block_query/utils"
)

func TestParseJson(t *testing.T) {
	data, err := u.ReadFixture("pets.json")
	if err != nil {
		t.Error(err)
	}

	Parse(data)
	t.Fail()
}
