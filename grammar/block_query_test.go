package grammar

import (
	"fmt"
	"testing"

	"github.com/auser/block_query/grammar"
)

func TestBlockQuery_Select(t *testing.T) {
	tests := []struct {
		query          string
		expectedASTLen int
	}{
		{"SELECT * from transactions", 28},
		// {"SELECT id FROM transactions", 27},
		// {"select * FROM transactions LIMIT 10", 34},
		// {"select * FROM transactions ORDER BY id ASC", 48},
		// {"select * FROM transactions ORDER BY id ASC LIMIT 10", 48}, // ordering doesn't matter
		// {"select * FROM transactions WHERE fromAddr='0xdeadbeef'", 58},
		// {"select * from transactions WHERE fromAddr='0xdeadbeef' AND value > 100", 76},
		// {"select * from transactions WHERE fromAddr='0xdeadbeef' AND value < 100", 76},
		// {"select * from transactions WHERE fromAddr='0xdeadbeef' AND toAddr='0xalivebeef'", 84},
		// {"select * from transactions WHERE fromAddr='0xdeadbeef' AND toAddr != '0xalivebeef'", 86},
		// {"select * from transactions WHERE fromAddr='0xdeadbeef' OR toAddr='0xalivebeef' LIMIT 10", 91},
	}

	for _, tt := range tests {
		grammar.SetDebugLevel(10, false)
		q, err := grammar.Parse(tt.query, "/tmp/test.go")

		if err != nil {
			t.Errorf(err.Error())
		}
		// fmt.Printf("q: %#v\n", q)
		fmt.Printf("q: %#v\n", q)
		t.Error()
	}
}
