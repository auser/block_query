package grammar

import (
	"testing"

	bq "github.com/auser/block_query/grammar"
)

func TestBlockQuery_Select(t *testing.T) {
	tests := []struct {
		query          string
		expectedASTLen int
	}{
		{"select * from transactions", 27},
		{"select ALL from transactions", 27},
		{"select * from transactions LIMIT 10", 34},
		{"select * from transactions LIMIT 10 ORDER BY id", 48},
		{"select * from transactions ORDER BY id LIMIT 10", 48}, // ordering doesn't matter
		{"select * from transactions WHERE from='0xdeadbeef'", 58},
		{"select * from transactions WHERE from='0xdeadbeef' AND value > 100", 76},
		{"select * from transactions WHERE from='0xdeadbeef' AND value < 100", 76},
		{"select * from transactions WHERE from='0xdeadbeef' AND to=0xalivebeef", 84},
	}

	for i, tt := range tests {
		q := &bq.BlockQuery{Buffer: tt.query}
		q.Init()

		if err := q.Parse(); err != nil {
			t.Fatalf("tests[%d] error: %s\n", i, err)
		}
		t.Logf("q: %v\n", len(q.Tokens()))

		if len(q.Tokens()) != tt.expectedASTLen {
			t.Errorf("tests[%d] incorrect AST len. expected=%q, got=%q", i, tt.expectedASTLen, len(q.Tokens()))
		}
	}
}
