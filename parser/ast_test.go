package parser

import (
	"fmt"
	"testing"

	"github.com/auser/block_query/utils"
)

func TestAST(t *testing.T) {
	tests := []struct {
		query string
		ast   *AST
	}{
		{
			query: "SELECT * from transactions",
			ast:   &AST{},
		},
	}

	_, err := utils.ReadFixture("1.json")
	if err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		fmt.Printf("query: %q\n", tt.query)
		parser := NewParser(tt.query)
		_, err := parser.Parse()
		if err != nil {
			t.Error(err)
		}
		// parser.Run([]byte(data))

		// ast, err := AST(*parser.AST)
		// if err != nil {
		// 	t.Error(err)
		// }

		// if ast != tt.ast {
		// 	t.Errorf("test[%d] unexpected AST. expected=%q, got=%q\n", i, tt.ast, ast)
		// }

		// t.Logf("Stmt: %q\n", ast)
		// t.Error()
	}
}
