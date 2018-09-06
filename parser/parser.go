package parser

import (
	sqlparser "github.com/xwb1989/sqlparser"
)

// Parser structure
type Parser struct {
	query       string
	statement   sqlparser.Statement
	latestError error
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
