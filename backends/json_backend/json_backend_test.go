package json_backend

import (
	"fmt"
	"testing"

	u "github.com/auser/block_query/utils"
)

func TestNewJSONBackend(t *testing.T) {

	data := u.MustBytes(u.ReadFixture("1.json"))
	n := u.MustInterface(NewJSONBackend(data))

	fmt.Printf("n: %#v\n", n)
	t.Error()
}
