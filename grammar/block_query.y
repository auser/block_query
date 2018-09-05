// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

%{
package grammar

import (
  // "strconv"
  "github.com/auser/block_query/value"
)

%}

%union {
  program     []Statement
  statement   Statement
  queryexpr   QueryExpression
  queryexprs  []QueryExpression
  expression  Expression
  expressions []Expression
  identifier  Identifier
  variable    Variable
  variables   []Variable
  token       Token
  table       Table
}

// Tokens
%token LEX_ERROR
%token<token>   recursive

%token<token> IDENTIFIER STRING INTEGER FLOAT BOOLEAN TERNARY DATETIME VARIABLE FLAG

%token<token> SELECT FROM WITH ORDER BY LIMIT OFFSET PARTITION TABLES
%token<token> VIEWS CURSORS FUNCTIONS FUNCTION_NTH FUNCTION_WITH_INS

%token<token> AND OR NOT
%token<token> ASC DESC
%token<token> FIRST LAST
%token<token> ERROR UMINUS
%token<token> COUNT LISTAGG ROWS
%token<token> AGGREGATE_FUNCTION ANALYTIC_FUNCTION
%token<token> NULLS NULL IN EXISTS TIES FIELDS
%token<token> ALL AS
%token<token> COMPARISON_OP STRING_OP PERCENT STDIN SUBSTITUTION_OP
%token<token> DISTINCT
%token<token> ';' '*' '=' '-' '+' '!' '(' ')'

%type <program> program

%type <statement> procedure_statement
%type <statement> common_statement

%type<queryexpr>    select_query
%type<queryexpr>    select_clause
%type<queryexpr>    from_clause
%type<queryexprs>    tables
%type<queryexprs>  identifiers
%type<queryexpr>    table
%type<queryexpr>    with_clause
%type<queryexpr>   inline_table
%type<queryexprs>  inline_tables
%type<queryexpr>    select_entity
%type<queryexpr>    order_by_clause
%type<queryexpr>    limit_clause
%type<queryexpr>    limit_with
%type<queryexpr>    offset_clause
%type<queryexpr>    field_object
%type<queryexpr>    field
%type<queryexprs>   fields
%type<queryexprs>   field_references
%type<queryexpr>   field_reference
%type<queryexpr>    value
%type<queryexprs>   values
%type<queryexpr>    wildcard
%type<queryexprs>   arguments
%type<queryexprs>   order_items
%type<queryexpr>    order_item
%type<queryexpr>    row_value
%type<queryexprs>    row_values
%type<queryexpr>    order_value
%type<token>        order_direction
%type<queryexpr>    null
%type<queryexpr>    aggregate_function
%type<token>        order_null_position
%type<queryexpr>    limit_with
%type<queryexpr>    offset_clause
%type<queryexpr>    primitive_type
%type<queryexpr>   arithmetic
%type<queryexpr>   ternary

%type<queryexpr>   logic
%type<queryexpr>    string_operation
%type<queryexpr>    subquery
%type<queryexpr>   comparison

%type<variable>     variable
%type<variables>    variables

%type<queryexpr>    variable_substitution
%type<queryexpr>   table_identifier
%type<table>        identified_table
%type<queryexpr>    virtual_table_object

%type<queryexpr>   analytic_function
// %type<queryexpr>   analytic_clause
// %type<queryexpr>   partition_clause

%type<identifier>  identifier

%type<token>       distinct

%right SUBSTITUTION_OP
%left UNION EXCEPT
%left INTERSECT
%left CROSS FULL NATURAL JOIN
%left OR
%left AND
%right NOT
%nonassoc '=' COMPARISON_OP IS BETWEEN IN LIKE
%left STRING_OP
%left '+' '-'
%left '*' '/' '%'
%right UMINUS UPLUS '!'

%start program

%%

program:
    {
      $$ = nil
      yylex.(*Lexer).program = $$
    }
    | procedure_statement
    {
      $$ = []Statement{$1}
      yylex.(*Lexer).program = $$
    }
    | procedure_statement ';' program
    {
      $$ = append([]Statement{$1}, $3...)
      yylex.(*Lexer).program = $$
    }
    ;

procedure_statement
    : common_statement
    {
        $$ = $1
    }
    ;

common_statement
    : select_query { $$ = $1 }
    ;

select_query
    : with_clause select_entity order_by_clause limit_clause offset_clause
    {
        $$ = SelectQuery{
            WithClause:    $1,
            SelectEntity:  $2,
            OrderByClause: $3,
            LimitClause:   $4,
            OffsetClause:  $5,
        }
    }
    | select_entity order_by_clause limit_clause offset_clause
    {
      $$ = SelectQuery{
        SelectEntity: $1,
        OrderByClause: $2,
        LimitClause: $3,
        OffsetClause: $4,
      }
    }
    ;

select_entity
    : select_clause from_clause
    {
      $$ = SelectEntity{
        SelectClause: $1,
        FromClause: $2,
      }
    }
    ;

select_clause
    : SELECT distinct fields
    {
      $$ = SelectClause{
        BaseExpr:NewBaseExpr($1), Select: $1.Literal, Distinct: $2, Fields: $3 }
    }
    ;

from_clause: FROM tables
    {
      $$ = FromClause{From: $1.Literal, Tables: $2}
    }
    ;

order_by_clause
    : { $$ = nil }
    | ORDER BY order_items
    {
      $$ = OrderByClause{OrderBy: $1.Literal + " " + $2.Literal, Items: $3}
    }

order_items
    : order_item { $$ = []QueryExpression{$1} }
    | order_item ',' order_items { $$ = append([]QueryExpression{$1}, $3...)}
    ;

order_item
    : order_value order_direction { $$ = OrderItem{Value: $1, Direction: $2}}
    | order_value order_direction NULLS order_null_position
    {
      $$ = OrderItem{Value: $1, Direction: $2, Nulls: $3.Literal, Position: $4}
    }
    ;

order_value
    : value { $$ = $1 }
    | analytic_function { $$ = $1 }
    ;

order_direction
    : { $$ = Token{} }
    | ASC { $$ = $1 }
    | DESC { $$ = $1 }
    ;

order_null_position
    : FIRST { $$ = $1 }
    | LAST { $$ = $1 }
    ;

limit_clause
    : { $$ = nil }
    | LIMIT value limit_with
    {
      $$ = LimitClause{BaseExpr: NewBaseExpr($1), Limit: $1.Literal, Value: $2, With: $3}
    }
    | LIMIT value PERCENT limit_with
    {
      $$ = LimitClause{BaseExpr: NewBaseExpr($1), Limit: $1.Literal, Value: $2, Percent: $3.Literal, With: $4}
    }
    ;

limit_with
    :
    {
        $$ = nil
    }
    | WITH TIES
    {
        $$ = LimitWith{With: $1.Literal, Type: $2}
    }

offset_clause
    : { $$ = nil }
    | OFFSET value
    {
      $$ = OffsetClause{BaseExpr: NewBaseExpr($1), Offset: $1.Literal, Value: $2 }
    }
    ;

field_object
    : value { $$ = $1 }
    | analytic_function { $$ = $1 }

field
    : field_object { $$ = Field{Object: $1}}
    | field_object AS identifier
    {
      $$ = Field{Object: $1, As: $2.Literal, Alias: $3}
    }
    | wildcard { $$ = Field{Object: $1} }
    ;

fields:
    field { $$ = []QueryExpression{$1} }
    | field ',' fields { $$ = append([]QueryExpression{$1}, $3...)}
    ;

tables:
    table { $$ = []QueryExpression{$1} }
    | table ',' tables
    {
      $$ = append([]QueryExpression{$1}, $3...)
    }

table
    : identified_table { $$ = $1 }
    | virtual_table_object { $$ = Table{Object: $1} }
    | virtual_table_object identifier { $$ = Table{Object: $1, Alias: $2} }
    // add joins here
    | '(' table ')' { $$ = Parentheses{Expr: $2}}
    ;

table_identifier
    : identifier
    {
        $$ = $1
    }
    | STDIN
    {
        $$ = Stdin{BaseExpr: NewBaseExpr($1), Stdin: $1.Literal}
    }

identified_table: table_identifier
    {
        $$ = Table{Object: $1}
    }
    | table_identifier identifier
    {
        $$ = Table{Object: $1, Alias: $2}
    }
    | table_identifier AS identifier
    {
        $$ = Table{Object: $1, As: $2.Literal, Alias: $3}
    }
virtual_table_object
    : subquery
    {
        $$ = $1
    }
    ;

with_clause
    : WITH inline_tables { $$ = WithClause{With: $1.Literal, InlineTables: $2} }
    ;

inline_table: recursive identifier AS '(' select_query ')'
    {
        $$ = InlineTable{Recursive: $1, Name: $2, As: $3.Literal, Query: $5.(SelectQuery)}
    }
    | recursive identifier '(' identifiers ')' AS '(' select_query ')'
    {
        $$ = InlineTable{Recursive: $1, Name: $2, Fields: $4, As: $6.Literal, Query: $8.(SelectQuery)}
    }
    ;

inline_tables
    : inline_table
    {
        $$ = []QueryExpression{$1}
    }
    | inline_table ',' inline_tables
    {
        $$ = append([]QueryExpression{$1}, $3...)
    }
    ;

identifiers
    : identifier { $$ = []QueryExpression{$1} }
    | identifier ',' identifiers { $$ = append([]QueryExpression{$1}, $3...)}
    ;


values
    : value { $$ = []QueryExpression{$1} }
    | value ',' values { $$ = append([]QueryExpression{$1}, $3...)}
    ;

analytic_function
    : identifier '(' arguments ')'
    {
      $$ = AnalyticFunction{BaseExpr: $1.BaseExpr, Name: $1.Literal, Args: $3}
    }
    | COUNT '(' distinct arguments ')'
    {
      $$ = AnalyticFunction{BaseExpr: NewBaseExpr($1), Name: $1.Literal, Distinct: $3, Args: $4}
    }
    | COUNT '(' distinct wildcard ')'
    {
      $$ = AnalyticFunction{BaseExpr: NewBaseExpr($1), Name: $1.Literal, Distinct: $3, Args: []QueryExpression{$4}}
    }
    ;

// analytic_clause
//     : partition_clause order_by_clause
//     {
//         $$ = AnalyticClause{PartitionClause: $1, OrderByClause: $2}
//     }
//     ;

// partition_clause
//     :
//     {
//         $$ = nil
//     }
//     | PARTITION BY values
//     {
//         $$ = PartitionClause{PartitionBy: $1.Literal + " " + $2.Literal, Values: $3}
//     }

arguments
    : { $$ = nil }
    | values { $$ = $1 }
    ;

wildcard
    : '*'
    {
        $$ = AllColumns{BaseExpr: NewBaseExpr($1)}
    }
    ;

row_value
    : '(' values ')'
    {
      $$ = RowValue{BaseExpr: NewBaseExpr($1), Value: ValueList{Values: $2}}
    }
    | subquery
    {
      $$ = RowValue{BaseExpr: $1.GetBaseExpr(), Value: $1}
    }

row_values
    : row_value
    {
      $$ = []QueryExpression{$1}
    }
    | row_value ',' row_values
    {
      $$ = append([]QueryExpression{$1}, $3...)
    }
    ;

subquery
    : '(' select_query ')'
    {
        $$ = Subquery{BaseExpr: NewBaseExpr($1), Query: $2.(SelectQuery)}
    }
    ;

string_operation
    : value STRING_OP value
    {
      var item1 []QueryExpression
      var item2 []QueryExpression

      c1, ok := $1.(Concat)
      if ok {
        item1 = c1.Items
      } else {
        item1 = []QueryExpression{$1}
      }

      c2, ok := $3.(Concat)
      if ok {
        item2 = c2.Items
      } else {
        item2 = []QueryExpression{$3}
      }

      $$ = Concat{Items: append(item1, item2...)}
    }
    ;

comparison
    : value COMPARISON_OP value
    {
      $$ = Comparison{LHS: $1, Operator: $2.Literal, RHS: $3 }
    }
    | row_value COMPARISON_OP row_value
    {
      $$ = Comparison{LHS: $1, Operator: $2.Literal, RHS: $3 }
    }
    | value '=' value
    {
      $$ = Comparison{LHS: $1, Operator: "=", RHS: $3 }
    }
    | row_value '=' row_value
    {
      $$ = Comparison{LHS: $1, Operator: "=", RHS: $3 }
    }
    | value IN row_value
    {
      $$ = In{In: $2.Literal, LHS: $1, Values: $3 }
    }
    | value NOT IN row_value
    {
      $$ = In{In: $3.Literal, LHS: $1, Values: $4, Negation: $2 }
    }
    // Lots more to do here
    | EXISTS subquery
    {
      $$ = Exists{Exists: $1.Literal, Query: $2.(Subquery)}
    }

field_references
    : field_reference
    {
        $$ = []QueryExpression{$1}
    }
    | field_reference ',' field_references
    {
        $$ = append([]QueryExpression{$1}, $3...)
    }
    ;

field_reference
    : identifier
    {
      $$ = FieldReference{BaseExpr: $1.BaseExpr, Column: $1}
    }
    | identifier '.' identifier
    {
      $$ = FieldReference{BaseExpr: $1.BaseExpr, View: $1, Column: $3}
    }
    | identifier '.' INTEGER
    {
        $$ = ColumnNumber{BaseExpr: $1.BaseExpr, View: $1, Number: value.NewIntegerFromString($3.Literal)}
    }
    ;

arithmetic
    : value '+' value { $$ = Arithmetic{LHS: $1, Operator: int('+'), RHS: $3} }
    | value '-' value { $$ = Arithmetic{LHS: $1, Operator: int('-'), RHS: $3} }
    | value '*' value { $$ = Arithmetic{LHS: $1, Operator: int('*'), RHS: $3} }
    | value '/' value { $$ = Arithmetic{LHS: $1, Operator: int('/'), RHS: $3} }
    | value '%' value { $$ = Arithmetic{LHS: $1, Operator: int('%'), RHS: $3} }
    | '-' value %prec UMINUS { $$ = UnaryArithmetic{Operand: $2, Operator: $1} }
    | '+' value %prec UMINUS { $$ = UnaryArithmetic{Operand: $2, Operator: $1} }
    ;

logic
    : value OR value
    {
      $$ = Logic{LHS: $1, Operator: $2, RHS: $3 }
    }
    | value AND value
    {
      $$ = Logic{LHS: $1, Operator: $2, RHS: $3 }
    }
    | NOT value { $$ = UnaryLogic{Operand: $2, Operator: $1 }}
    | '!' value { $$ = UnaryLogic{Operand: $2, Operator: $1 }}
    ;

aggregate_function
  : identifier '(' distinct arguments ')'
  {
    $$ = AggregateFunction{BaseExpr: $1.BaseExpr, Name: $1.Literal, Distinct: $3, Args: $4 }
  }
  | AGGREGATE_FUNCTION '(' distinct arguments ')'
  {
    $$ = AggregateFunction{BaseExpr: NewBaseExpr($1), Name: $1.Literal, Distinct: $3, Args: $4}
  }
  | COUNT '(' distinct arguments ')'
  {
    $$ = AggregateFunction{BaseExpr: NewBaseExpr($1), Name: $1.Literal, Distinct: $3, Args: $4}
  }
  | COUNT '(' distinct wildcard ')'
  {
    $$ = AggregateFunction{BaseExpr: NewBaseExpr($1), Name: $1.Literal, Distinct: $3, Args: []QueryExpression{$4}}
  }
  ;

identifier
    : IDENTIFIER
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | TIES
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | NULLS
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | TABLES
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | VIEWS
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | CURSORS
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | FUNCTIONS
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | ROWS
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | FIELDS
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | COUNT
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | LISTAGG
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | AGGREGATE_FUNCTION
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | ANALYTIC_FUNCTION
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | FUNCTION_NTH
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | FUNCTION_WITH_INS
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    | ERROR
    {
        $$ = Identifier{BaseExpr: NewBaseExpr($1), Literal: $1.Literal, Quoted: $1.Quoted}
    }
    ;

null
    : NULL { $$ = NewNullValueFromString($1.Literal) }

variable
    : VARIABLE
      {
        $$ = Variable{BaseExpr: NewBaseExpr($1), Name: $1.Literal}
      }
    ;

variables
    : variable ',' variables { $$ = append([]Variable{$1}, $3...) }
    | variable
      {
        $$ = []Variable{$1}
      }
    ;

variable_substitution
    : variable SUBSTITUTION_OP value
    {
      $$ = VariableSubstitution{Variable: $1, Value: $3}
    }
    ;

ternary
    : TERNARY
    {
        $$ = NewTernaryValueFromString($1.Literal)
    }
    ;

value
    : field_reference { $$ = $1 }
    | primitive_type { $$ = $1 }
    | arithmetic { $$ = $1 }
    | string_operation { $$ = $1 }
    | subquery { $$ = $1 }
    | aggregate_function { $$ = $1 }
    | comparison { $$ = $1 }
    | logic { $$ = $1 }
    | variable { $$ = $1 }
    | variable_substitution { $$ = $1 }
    | '(' value ')' { $$ = Parentheses{Expr: $2} }
    ;

primitive_type
    : STRING
    {
        $$ = NewStringValue($1.Literal)
    }
    | INTEGER
    {
        $$ = NewIntegerValueFromString($1.Literal)
    }
    | FLOAT
    {
        $$ = NewFloatValueFromString($1.Literal)
    }
    | ternary
    {
        $$ = $1
    }
    | DATETIME
    {
        $$ = NewDatetimeValueFromString($1.Literal)
    }
    | null
    {
        $$ = $1
    }
    ;

distinct
    : { $$ = Token{} }
    | DISTINCT { $$ = $1 }
    ;

// all: { $$ = Token{} }
//     | ALL { $$ = $1 }
//     ;

// as: { $$ = Token{} }
//     | AS { $$ = $1 }
//     ;

%%

func SetDebugLevel(level int, verbose bool) {
	yyDebug        = level
	yyErrorVerbose = verbose
}

func Parse(s string, sourceFile string) ([]Statement, error) {
    l := new(Lexer)
    l.Init(s, sourceFile)
    yyParse(l)
    return l.program, l.err
}
