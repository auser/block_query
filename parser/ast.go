package parser

import (
	"fmt"

	"github.com/xwb1989/sqlparser"
)

type AST struct {
	action string
}

var (
	ast *AST = &AST{}
)

func visit(node sqlparser.SQLNode) (cnt bool, err error) {
	fmt.Printf("visiting... %q\n", node)
	return true, nil
}

func (a *AST) intoAST(stmt sqlparser.Statement) (*AST, error) {
	sqlparser.Walk(visit, stmt)
	return &AST{}, nil
}
