package parser

import (
	"fmt"

	sqlparser "github.com/xwb1989/sqlparser"
)

// Parser structure
type Parser struct {
	query       string
	statement   sqlparser.Statement
	latestError error
}

func Must(op Op, err error) Op {
	if err != nil {
		panic(fmt.Errorf("unable to parse selector; %v", err.Error()))
	}

	return op
}

// NewParser generates a new parser
func NewParser(s string) *Parser {
	return &Parser{
		query: s,
	}
}

// Parse parses a sql statement using xwb1989
func (p *Parser) Parse() (sqlparser.Statement, error) {
	stmt, err := sqlparser.Parse(p.query)
	if err != nil {
		p.latestError = err
		return nil, err
	}

	p.statement = stmt
	return stmt, err
}

func (p *Parser) hasBeenParsed() bool {
	return p.statement != nil
}

// Statement of the parsed statement
func (p *Parser) Statement() sqlparser.Statement {
	if !p.hasBeenParsed() {
		p.Parse()
	}

	return p.statement
}

// Run executes the parser over the data
func (p *Parser) Run(str string) (interface{}, error) {
	fmt.Printf("Running %s\n", str)
	var res map[string]interface{}
	return res, nil
}

// AST gets the latest ast
func (p *Parser) AST() (*AST, error) {
	ast := &AST{}
	stmt := p.Statement()
	return ast.intoAST(stmt)
}
