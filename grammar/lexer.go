package grammar

import "fmt"

type Lexer struct {
	Scanner
	program []Statement
	token   Token
	err     error
}

func (l *Lexer) Lex(lval *yySymType) int {
	tok, err := l.Scan()
	if err != nil {
		l.Error(err.Error())
	}

	lval.token = tok
	l.token = lval.token
	return tok.Token
}

func (l *Lexer) Error(e string) {
	fmt.Printf("tok: %v %s\n", TOKEN_FROM <= l.token.Token && l.token.Token <= ERROR, string(l.token.Token))
	if 0 < l.token.Token {
		var lit string
		if TOKEN_FROM <= l.token.Token && l.token.Token <= ERROR {
			lit = TokenLiteral(l.token.Token)
		} else if COUNT <= l.token.Token && l.token.Token <= FUNCTION_WITH_INS {
			lit = TokenLiteral(IDENTIFIER)
		} else {
			lit = l.token.Literal
		}

		l.err = NewSyntaxError(fmt.Sprintf("%s: unexpected %s", e, lit), l.token)
	} else if e == "syntax error" && l.token.Token == -1 {
		l.err = NewSyntaxError(fmt.Sprintf("%s: unexpected termination", e), l.token)
	} else {
		l.err = NewSyntaxError(fmt.Sprintf("%s", e), l.token)
	}
}

type Token struct {
	Token      int
	Literal    string
	Quoted     bool
	Line       int
	Char       int
	SourceFile string
}

func (t *Token) IsEmpty() bool {
	return len(t.Literal) < 1
}

type SyntaxError struct {
	SourceFile string
	Line       int
	Char       int
	Message    string
}

func (e SyntaxError) Error() string {
	return e.Message
}

func NewSyntaxError(message string, token Token) error {
	return &SyntaxError{
		SourceFile: token.SourceFile,
		Line:       token.Line,
		Char:       token.Char,
		Message:    message,
	}
}
