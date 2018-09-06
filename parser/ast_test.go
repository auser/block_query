package parser

import (
	"fmt"
	"testing"

	"github.com/auser/block_query/parser"
	"github.com/auser/block_query/utils"
)

func TestAST(t *testing.T) {
	tests := []struct {
		query string
		ast   *parser.AST
	}{
		{
			query: "SELECT * from transactions",
			ast:   &parser.AST{},
		},
	}

	data, err := utils.ReadFixture("1.json")
	if err != nil {
		t.Error(err)
	}

	for i, tt := range tests {
		fmt.Printf("query: %q\n", tt.query)
		parser := parser.NewParser(tt.query)
		_, err := parser.Parse()
		parser.Run(data)
		if err != nil {
			t.Error(err)
		}

		ast, err := parser.AST()
		if err != nil {
			t.Error(err)
		}

		if ast != tt.ast {
			t.Errorf("test[%d] unexpected AST. expected=%q, got=%q\n", i, tt.ast, ast)
		}

		t.Logf("Stmt: %q\n", ast)
		t.Error()
	}
}
