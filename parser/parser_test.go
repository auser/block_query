package parser_test

import (
	"fmt"
	"testing"

	"github.com/auser/block_query/parser"
	sqlparser "github.com/xwb1989/sqlparser"
)

func TestBlockQuery_Select(t *testing.T) {
	tests := []struct {
		query          string
		expectedOutput string
	}{
		{
			"SELECT * from transactions",
			"select * from transactions",
		},
		{
			"SELECT id from transactions",
			"select id from transactions",
		},
		{
			"SELECT *",
			"select * from dual",
		},
		{
			"SELECT * from transactions where fromAddr='0xdeadbeef'",
			"select * from transactions where fromAddr = '0xdeadbeef'",
		},
		{
			"SELECT * from transactions where fromAddr='0xdeadbeef' LIMIT 10",
			"select * from transactions where fromAddr = '0xdeadbeef' limit 10",
		},
		{
			"SELECT * from transactions where toAddr='0xdeadbeef' ORDER BY id",
			"select * from transactions where toAddr = '0xdeadbeef' order by id asc",
		},
		{
			"SELECT * from transactions where toAddr='0xdeadbeef' ORDER BY id LIMIT 10",
			"select * from transactions where toAddr = '0xdeadbeef' order by id asc limit 10",
		},
	}

	for i, tt := range tests {

		parser := parser.NewParser(tt.query)
		stmt, _ := parser.Parse()
		switch stmt.(type) {
		case *sqlparser.Select:
			if sqlparser.String(stmt) != tt.expectedOutput {
				t.Errorf("test[%d] unexpected output. expected=%q, got=%q\n", i, tt.expectedOutput, sqlparser.String(stmt))
			}
		default:
			fmt.Printf("Unknown: %q\n", stmt)
		}
	}
}

func TestBlockQueryAST(t *testing.T) {
	tests := []struct {
		query string
	}{
		{
			"SELECT * from transactions",
		},
		{
			"SELECT  * from transactions LIMIT 1",
		},
	}

	for i, tt := range tests {
		parser := parser.NewParser(tt.query)
		stmt, err := parser.Parse()

		if err != nil {
			t.Error(err)
		}

		switch stmt.(type) {
		case *sqlparser.Select:
			t.Logf("Got a select: %q\n", stmt)
		default:
			t.Errorf("Should never get here (test: %d)\n", i)
		}
	}
}

func TestSubStr(t *testing.T) {

	validSQL := []struct {
		input  string
		output string
	}{{
		input: "select substr(a, 1) from t",
	}, {
		input: "select substr(a, 1, 6) from t",
	}, {
		input:  "select substring(a, 1) from t",
		output: "select substr(a, 1) from t",
	}, {
		input:  "select substring(a, 1, 6) from t",
		output: "select substr(a, 1, 6) from t",
	}, {
		input:  "select substr(a from 1 for 6) from t",
		output: "select substr(a, 1, 6) from t",
	}, {
		input:  "select substring(a from 1 for 6) from t",
		output: "select substr(a, 1, 6) from t",
	}}

	for _, tcase := range validSQL {
		if tcase.output == "" {
			tcase.output = tcase.input
		}
		parser := parser.NewParser(tcase.input)
		tree, err := parser.Parse()
		if err != nil {
			t.Errorf("input: %s, err: %v", tcase.input, err)
			continue
		}
		out := sqlparser.String(tree)
		if out != tcase.output {
			t.Errorf("out: %s, want %s", out, tcase.output)
		}
	}
}
