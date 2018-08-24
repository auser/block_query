package grammar

import (
	"fmt"
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
	}

	for i, tt := range tests {
		q := &bq.BlockQuery{Buffer: tt.query}
		q.Init()

		if err := q.Parse(); err != nil {
			t.Fatalf("tests[%d] error: %s\n", i, err)
		}
		fmt.Printf("q: %v\n", len(q.Tokens()))

		if len(q.Tokens()) != tt.expectedASTLen {
			t.Error("tests[%d] error in expected length", i)
		}
	}
}
