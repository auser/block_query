%{

package json_scanner

import (
  "log"
  "fmt"
  "os"
)
%}

%union{
  tok int
  val interface{}
  pair struct{key, val interface{}}
  pairs map[interface{}]interface{}
}

%token KEY
%token VAL

%type <val> KEY VAL
%type <pair> pair
%type <pairs> pairs

%%

goal:
  '{' pairs '}'
  {
    yylex.(*lex).m = $2
  }
  ;

pairs:
  pair
  {
    $$ = map[interface{}]interface{}{$1.key: $1.val}
  }
  | pairs '|' pair
  {
    $$[$3.key] = $3.val
  }
  ;

pair:
  KEY '=' VAL
  {
    $$.key, $$.val = $1, $3
  }
  | KEY '=' '{' pairs '}'
  {
    $$.key, $$.val = $1, $4
  }
  ;

%%

type Lexer struct {
  s string
  pos int
}

func (l *Lexer) Lex(lval *JsonSymType) int {
  var c rune == ' '
}

func SetDebugLevel(level int, verbose bool) {
	yyDebug        = level
	yyErrorVerbose = verbose
}

func Parse(s string, sourceFile string) (map[interface{}]interface{}, error) {
    l := new(Lexer)
    l.Init(s, sourceFile)
    yyParse(l)
    return l.program, l.err
}
