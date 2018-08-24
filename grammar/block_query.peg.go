package grammar

//go:generate /Users/auser/.gvm/pkgsets/go1.9.2/global/bin/peg grammar/block_query.peg

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleResult
	rulequeryStmt
	rulequeryExprs
	rulequeryExpr
	ruleselectStmt
	rulefromStmt
	rulelimitStmt
	ruleorderStmt
	rulewhereStmt
	rulewhereExprs
	rulewhereExpr
	ruleasc
	ruledesc
	ruleselect
	rulefrom
	ruleorder
	ruleby
	rulestar
	rulelimit
	rulewhere
	ruleand
	ruleor
	ruleinStmt
	rulegreaterThan
	rulelessThan
	rulegreaterThanOrEq
	rulelessThanOrEq
	rulenotEq
	rulesemi
	rulelparen
	rulerparen
	rulecomma
	ruleequals
	rulesquote
	ruledquote
	ruleStringLiteralList
	ruleNumericList
	ruleLetter
	ruleLetterOrDigit
	ruleHexDigit
	ruleDecimalNumeral
	ruleStringChar
	ruleStringLiteral
	ruleEscape
	ruleOctalEscape
	ruleUnicodeEscape
	ruleWordList
	ruleWord
	rulespace
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	rulePegText
	ruleAction17
	ruleAction18
)

var rul3s = [...]string{
	"Unknown",
	"Result",
	"queryStmt",
	"queryExprs",
	"queryExpr",
	"selectStmt",
	"fromStmt",
	"limitStmt",
	"orderStmt",
	"whereStmt",
	"whereExprs",
	"whereExpr",
	"asc",
	"desc",
	"select",
	"from",
	"order",
	"by",
	"star",
	"limit",
	"where",
	"and",
	"or",
	"inStmt",
	"greaterThan",
	"lessThan",
	"greaterThanOrEq",
	"lessThanOrEq",
	"notEq",
	"semi",
	"lparen",
	"rparen",
	"comma",
	"equals",
	"squote",
	"dquote",
	"StringLiteralList",
	"NumericList",
	"Letter",
	"LetterOrDigit",
	"HexDigit",
	"DecimalNumeral",
	"StringChar",
	"StringLiteral",
	"Escape",
	"OctalEscape",
	"UnicodeEscape",
	"WordList",
	"Word",
	"space",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"PegText",
	"Action17",
	"Action18",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Printf("%v %v\n", rule, quote)
			} else {
				fmt.Printf("\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(buffer string) {
	node.print(false, buffer)
}

func (node *node32) PrettyPrint(buffer string) {
	node.print(true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type BlockQuery struct {
	Query

	strStack    []string
	workingWord string
	ExprStack   ExprStack

	Buffer string
	buffer []rune
	rules  [70]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *BlockQuery) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *BlockQuery) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *BlockQuery
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *BlockQuery) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *BlockQuery) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:

			p.Fields = p.strStack
			p.strStack = []string{}

		case ruleAction1:
			p.Table = buffer[begin:end]
		case ruleAction2:

			f, _ := strconv.Atoi(text)
			p.Limit.ShouldLimit = true
			p.Limit.Count = f

		case ruleAction3:
			p.Order.Field = text
		case ruleAction4:
			p.Order.Ordering = ASC
		case ruleAction5:
			p.Order.Ordering = DESC
		case ruleAction6:

			p.ExprStack.Push(&NodeAnd{
				left:  p.ExprStack.Pop(),
				right: p.ExprStack.Pop(),
			})

		case ruleAction7:

		case ruleAction8:

			p.ExprStack.Push(&NodeEquals{
				left:  p.workingWord,
				right: text,
			})

		case ruleAction9:

			p.ExprStack.Push(&NodeEquals{
				left:  p.workingWord,
				right: text,
			})

		case ruleAction10:

			f, _ := strconv.Atoi(text)
			p.ExprStack.Push(&NodeGreaterThan{
				left:  p.workingWord,
				right: f,
			})

		case ruleAction11:

			f, _ := strconv.Atoi(text)
			p.ExprStack.Push(&NodeGreaterThanOrEqual{
				left:  p.workingWord,
				right: f,
			})

		case ruleAction12:

			f, _ := strconv.Atoi(text)
			p.ExprStack.Push(&NodeLessThan{
				left:  p.workingWord,
				right: f,
			})

		case ruleAction13:

			f, _ := strconv.Atoi(text)
			p.ExprStack.Push(&NodeLessThanOrEqual{
				left:  p.workingWord,
				right: f,
			})

		case ruleAction14:

			p.ExprStack.Push(&NodeNotEqual{
				left:  p.workingWord,
				right: text,
			})

		case ruleAction15:

			p.ExprStack.Push(&NodeNotEqual{
				left:  p.workingWord,
				right: text,
			})

		case ruleAction16:
			text = "star"
		case ruleAction17:

			p.strStack = append(p.strStack, text)

		case ruleAction18:
			p.workingWord = text

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *BlockQuery) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Result <- <(queryStmt semi? !.)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[rulequeryStmt]() {
					goto l0
				}
				{
					position2, tokenIndex2 := position, tokenIndex
					if !_rules[rulesemi]() {
						goto l2
					}
					goto l3
				l2:
					position, tokenIndex = position2, tokenIndex2
				}
			l3:
				{
					position4, tokenIndex4 := position, tokenIndex
					if !matchDot() {
						goto l4
					}
					goto l0
				l4:
					position, tokenIndex = position4, tokenIndex4
				}
				add(ruleResult, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 queryStmt <- <(selectStmt (space queryExprs)*)> */
		func() bool {
			position5, tokenIndex5 := position, tokenIndex
			{
				position6 := position
				if !_rules[ruleselectStmt]() {
					goto l5
				}
			l7:
				{
					position8, tokenIndex8 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l8
					}
					if !_rules[rulequeryExprs]() {
						goto l8
					}
					goto l7
				l8:
					position, tokenIndex = position8, tokenIndex8
				}
				add(rulequeryStmt, position6)
			}
			return true
		l5:
			position, tokenIndex = position5, tokenIndex5
			return false
		},
		/* 2 queryExprs <- <(queryExpr (space queryExprs)*)> */
		func() bool {
			position9, tokenIndex9 := position, tokenIndex
			{
				position10 := position
				if !_rules[rulequeryExpr]() {
					goto l9
				}
			l11:
				{
					position12, tokenIndex12 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l12
					}
					if !_rules[rulequeryExprs]() {
						goto l12
					}
					goto l11
				l12:
					position, tokenIndex = position12, tokenIndex12
				}
				add(rulequeryExprs, position10)
			}
			return true
		l9:
			position, tokenIndex = position9, tokenIndex9
			return false
		},
		/* 3 queryExpr <- <(limitStmt / orderStmt / whereStmt)> */
		func() bool {
			position13, tokenIndex13 := position, tokenIndex
			{
				position14 := position
				{
					position15, tokenIndex15 := position, tokenIndex
					if !_rules[rulelimitStmt]() {
						goto l16
					}
					goto l15
				l16:
					position, tokenIndex = position15, tokenIndex15
					if !_rules[ruleorderStmt]() {
						goto l17
					}
					goto l15
				l17:
					position, tokenIndex = position15, tokenIndex15
					if !_rules[rulewhereStmt]() {
						goto l13
					}
				}
			l15:
				add(rulequeryExpr, position14)
			}
			return true
		l13:
			position, tokenIndex = position13, tokenIndex13
			return false
		},
		/* 4 selectStmt <- <(select (star / WordList) fromStmt Action0)> */
		func() bool {
			position18, tokenIndex18 := position, tokenIndex
			{
				position19 := position
				if !_rules[ruleselect]() {
					goto l18
				}
				{
					position20, tokenIndex20 := position, tokenIndex
					if !_rules[rulestar]() {
						goto l21
					}
					goto l20
				l21:
					position, tokenIndex = position20, tokenIndex20
					if !_rules[ruleWordList]() {
						goto l18
					}
				}
			l20:
				if !_rules[rulefromStmt]() {
					goto l18
				}
				if !_rules[ruleAction0]() {
					goto l18
				}
				add(ruleselectStmt, position19)
			}
			return true
		l18:
			position, tokenIndex = position18, tokenIndex18
			return false
		},
		/* 5 fromStmt <- <(space? from Word Action1)> */
		func() bool {
			position22, tokenIndex22 := position, tokenIndex
			{
				position23 := position
				{
					position24, tokenIndex24 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l24
					}
					goto l25
				l24:
					position, tokenIndex = position24, tokenIndex24
				}
			l25:
				if !_rules[rulefrom]() {
					goto l22
				}
				if !_rules[ruleWord]() {
					goto l22
				}
				if !_rules[ruleAction1]() {
					goto l22
				}
				add(rulefromStmt, position23)
			}
			return true
		l22:
			position, tokenIndex = position22, tokenIndex22
			return false
		},
		/* 6 limitStmt <- <(limit DecimalNumeral Action2)> */
		func() bool {
			position26, tokenIndex26 := position, tokenIndex
			{
				position27 := position
				if !_rules[rulelimit]() {
					goto l26
				}
				if !_rules[ruleDecimalNumeral]() {
					goto l26
				}
				if !_rules[ruleAction2]() {
					goto l26
				}
				add(rulelimitStmt, position27)
			}
			return true
		l26:
			position, tokenIndex = position26, tokenIndex26
			return false
		},
		/* 7 orderStmt <- <(order by WordList Action3 ((space asc Action4) / (space desc Action5))?)> */
		func() bool {
			position28, tokenIndex28 := position, tokenIndex
			{
				position29 := position
				if !_rules[ruleorder]() {
					goto l28
				}
				if !_rules[ruleby]() {
					goto l28
				}
				if !_rules[ruleWordList]() {
					goto l28
				}
				if !_rules[ruleAction3]() {
					goto l28
				}
				{
					position30, tokenIndex30 := position, tokenIndex
					{
						position32, tokenIndex32 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l33
						}
						if !_rules[ruleasc]() {
							goto l33
						}
						if !_rules[ruleAction4]() {
							goto l33
						}
						goto l32
					l33:
						position, tokenIndex = position32, tokenIndex32
						if !_rules[rulespace]() {
							goto l30
						}
						if !_rules[ruledesc]() {
							goto l30
						}
						if !_rules[ruleAction5]() {
							goto l30
						}
					}
				l32:
					goto l31
				l30:
					position, tokenIndex = position30, tokenIndex30
				}
			l31:
				add(ruleorderStmt, position29)
			}
			return true
		l28:
			position, tokenIndex = position28, tokenIndex28
			return false
		},
		/* 8 whereStmt <- <(where whereExprs)> */
		func() bool {
			position34, tokenIndex34 := position, tokenIndex
			{
				position35 := position
				if !_rules[rulewhere]() {
					goto l34
				}
				if !_rules[rulewhereExprs]() {
					goto l34
				}
				add(rulewhereStmt, position35)
			}
			return true
		l34:
			position, tokenIndex = position34, tokenIndex34
			return false
		},
		/* 9 whereExprs <- <(whereExpr (space? whereExprs)*)> */
		func() bool {
			position36, tokenIndex36 := position, tokenIndex
			{
				position37 := position
				if !_rules[rulewhereExpr]() {
					goto l36
				}
			l38:
				{
					position39, tokenIndex39 := position, tokenIndex
					{
						position40, tokenIndex40 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l40
						}
						goto l41
					l40:
						position, tokenIndex = position40, tokenIndex40
					}
				l41:
					if !_rules[rulewhereExprs]() {
						goto l39
					}
					goto l38
				l39:
					position, tokenIndex = position39, tokenIndex39
				}
				add(rulewhereExprs, position37)
			}
			return true
		l36:
			position, tokenIndex = position36, tokenIndex36
			return false
		},
		/* 10 whereExpr <- <((and whereExpr Action6) / (or whereExpr Action7) / (Word equals StringLiteral Action8) / (Word equals DecimalNumeral Action9) / (Word greaterThan DecimalNumeral Action10) / (Word greaterThanOrEq DecimalNumeral Action11) / (Word lessThan DecimalNumeral Action12) / (Word lessThanOrEq DecimalNumeral Action13) / (Word notEq StringLiteral Action14) / (Word notEq DecimalNumeral Action15) / (Word inStmt (StringLiteralList / NumericList)))> */
		func() bool {
			position42, tokenIndex42 := position, tokenIndex
			{
				position43 := position
				{
					position44, tokenIndex44 := position, tokenIndex
					if !_rules[ruleand]() {
						goto l45
					}
					if !_rules[rulewhereExpr]() {
						goto l45
					}
					if !_rules[ruleAction6]() {
						goto l45
					}
					goto l44
				l45:
					position, tokenIndex = position44, tokenIndex44
					if !_rules[ruleor]() {
						goto l46
					}
					if !_rules[rulewhereExpr]() {
						goto l46
					}
					if !_rules[ruleAction7]() {
						goto l46
					}
					goto l44
				l46:
					position, tokenIndex = position44, tokenIndex44
					if !_rules[ruleWord]() {
						goto l47
					}
					if !_rules[ruleequals]() {
						goto l47
					}
					if !_rules[ruleStringLiteral]() {
						goto l47
					}
					if !_rules[ruleAction8]() {
						goto l47
					}
					goto l44
				l47:
					position, tokenIndex = position44, tokenIndex44
					if !_rules[ruleWord]() {
						goto l48
					}
					if !_rules[ruleequals]() {
						goto l48
					}
					if !_rules[ruleDecimalNumeral]() {
						goto l48
					}
					if !_rules[ruleAction9]() {
						goto l48
					}
					goto l44
				l48:
					position, tokenIndex = position44, tokenIndex44
					if !_rules[ruleWord]() {
						goto l49
					}
					if !_rules[rulegreaterThan]() {
						goto l49
					}
					if !_rules[ruleDecimalNumeral]() {
						goto l49
					}
					if !_rules[ruleAction10]() {
						goto l49
					}
					goto l44
				l49:
					position, tokenIndex = position44, tokenIndex44
					if !_rules[ruleWord]() {
						goto l50
					}
					if !_rules[rulegreaterThanOrEq]() {
						goto l50
					}
					if !_rules[ruleDecimalNumeral]() {
						goto l50
					}
					if !_rules[ruleAction11]() {
						goto l50
					}
					goto l44
				l50:
					position, tokenIndex = position44, tokenIndex44
					if !_rules[ruleWord]() {
						goto l51
					}
					if !_rules[rulelessThan]() {
						goto l51
					}
					if !_rules[ruleDecimalNumeral]() {
						goto l51
					}
					if !_rules[ruleAction12]() {
						goto l51
					}
					goto l44
				l51:
					position, tokenIndex = position44, tokenIndex44
					if !_rules[ruleWord]() {
						goto l52
					}
					if !_rules[rulelessThanOrEq]() {
						goto l52
					}
					if !_rules[ruleDecimalNumeral]() {
						goto l52
					}
					if !_rules[ruleAction13]() {
						goto l52
					}
					goto l44
				l52:
					position, tokenIndex = position44, tokenIndex44
					if !_rules[ruleWord]() {
						goto l53
					}
					if !_rules[rulenotEq]() {
						goto l53
					}
					if !_rules[ruleStringLiteral]() {
						goto l53
					}
					if !_rules[ruleAction14]() {
						goto l53
					}
					goto l44
				l53:
					position, tokenIndex = position44, tokenIndex44
					if !_rules[ruleWord]() {
						goto l54
					}
					if !_rules[rulenotEq]() {
						goto l54
					}
					if !_rules[ruleDecimalNumeral]() {
						goto l54
					}
					if !_rules[ruleAction15]() {
						goto l54
					}
					goto l44
				l54:
					position, tokenIndex = position44, tokenIndex44
					if !_rules[ruleWord]() {
						goto l42
					}
					if !_rules[ruleinStmt]() {
						goto l42
					}
					{
						position55, tokenIndex55 := position, tokenIndex
						if !_rules[ruleStringLiteralList]() {
							goto l56
						}
						goto l55
					l56:
						position, tokenIndex = position55, tokenIndex55
						if !_rules[ruleNumericList]() {
							goto l42
						}
					}
				l55:
				}
			l44:
				add(rulewhereExpr, position43)
			}
			return true
		l42:
			position, tokenIndex = position42, tokenIndex42
			return false
		},
		/* 11 asc <- <(('a' / 'A') ('s' / 'S') ('c' / 'C'))> */
		func() bool {
			position57, tokenIndex57 := position, tokenIndex
			{
				position58 := position
				{
					position59, tokenIndex59 := position, tokenIndex
					if buffer[position] != rune('a') {
						goto l60
					}
					position++
					goto l59
				l60:
					position, tokenIndex = position59, tokenIndex59
					if buffer[position] != rune('A') {
						goto l57
					}
					position++
				}
			l59:
				{
					position61, tokenIndex61 := position, tokenIndex
					if buffer[position] != rune('s') {
						goto l62
					}
					position++
					goto l61
				l62:
					position, tokenIndex = position61, tokenIndex61
					if buffer[position] != rune('S') {
						goto l57
					}
					position++
				}
			l61:
				{
					position63, tokenIndex63 := position, tokenIndex
					if buffer[position] != rune('c') {
						goto l64
					}
					position++
					goto l63
				l64:
					position, tokenIndex = position63, tokenIndex63
					if buffer[position] != rune('C') {
						goto l57
					}
					position++
				}
			l63:
				add(ruleasc, position58)
			}
			return true
		l57:
			position, tokenIndex = position57, tokenIndex57
			return false
		},
		/* 12 desc <- <(('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C'))> */
		func() bool {
			position65, tokenIndex65 := position, tokenIndex
			{
				position66 := position
				{
					position67, tokenIndex67 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l68
					}
					position++
					goto l67
				l68:
					position, tokenIndex = position67, tokenIndex67
					if buffer[position] != rune('D') {
						goto l65
					}
					position++
				}
			l67:
				{
					position69, tokenIndex69 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l70
					}
					position++
					goto l69
				l70:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('E') {
						goto l65
					}
					position++
				}
			l69:
				{
					position71, tokenIndex71 := position, tokenIndex
					if buffer[position] != rune('s') {
						goto l72
					}
					position++
					goto l71
				l72:
					position, tokenIndex = position71, tokenIndex71
					if buffer[position] != rune('S') {
						goto l65
					}
					position++
				}
			l71:
				{
					position73, tokenIndex73 := position, tokenIndex
					if buffer[position] != rune('c') {
						goto l74
					}
					position++
					goto l73
				l74:
					position, tokenIndex = position73, tokenIndex73
					if buffer[position] != rune('C') {
						goto l65
					}
					position++
				}
			l73:
				add(ruledesc, position66)
			}
			return true
		l65:
			position, tokenIndex = position65, tokenIndex65
			return false
		},
		/* 13 select <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') space)> */
		func() bool {
			position75, tokenIndex75 := position, tokenIndex
			{
				position76 := position
				{
					position77, tokenIndex77 := position, tokenIndex
					if buffer[position] != rune('s') {
						goto l78
					}
					position++
					goto l77
				l78:
					position, tokenIndex = position77, tokenIndex77
					if buffer[position] != rune('S') {
						goto l75
					}
					position++
				}
			l77:
				{
					position79, tokenIndex79 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l80
					}
					position++
					goto l79
				l80:
					position, tokenIndex = position79, tokenIndex79
					if buffer[position] != rune('E') {
						goto l75
					}
					position++
				}
			l79:
				{
					position81, tokenIndex81 := position, tokenIndex
					if buffer[position] != rune('l') {
						goto l82
					}
					position++
					goto l81
				l82:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('L') {
						goto l75
					}
					position++
				}
			l81:
				{
					position83, tokenIndex83 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l84
					}
					position++
					goto l83
				l84:
					position, tokenIndex = position83, tokenIndex83
					if buffer[position] != rune('E') {
						goto l75
					}
					position++
				}
			l83:
				{
					position85, tokenIndex85 := position, tokenIndex
					if buffer[position] != rune('c') {
						goto l86
					}
					position++
					goto l85
				l86:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('C') {
						goto l75
					}
					position++
				}
			l85:
				{
					position87, tokenIndex87 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l88
					}
					position++
					goto l87
				l88:
					position, tokenIndex = position87, tokenIndex87
					if buffer[position] != rune('T') {
						goto l75
					}
					position++
				}
			l87:
				if !_rules[rulespace]() {
					goto l75
				}
				add(ruleselect, position76)
			}
			return true
		l75:
			position, tokenIndex = position75, tokenIndex75
			return false
		},
		/* 14 from <- <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M') space)> */
		func() bool {
			position89, tokenIndex89 := position, tokenIndex
			{
				position90 := position
				{
					position91, tokenIndex91 := position, tokenIndex
					if buffer[position] != rune('f') {
						goto l92
					}
					position++
					goto l91
				l92:
					position, tokenIndex = position91, tokenIndex91
					if buffer[position] != rune('F') {
						goto l89
					}
					position++
				}
			l91:
				{
					position93, tokenIndex93 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l94
					}
					position++
					goto l93
				l94:
					position, tokenIndex = position93, tokenIndex93
					if buffer[position] != rune('R') {
						goto l89
					}
					position++
				}
			l93:
				{
					position95, tokenIndex95 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l96
					}
					position++
					goto l95
				l96:
					position, tokenIndex = position95, tokenIndex95
					if buffer[position] != rune('O') {
						goto l89
					}
					position++
				}
			l95:
				{
					position97, tokenIndex97 := position, tokenIndex
					if buffer[position] != rune('m') {
						goto l98
					}
					position++
					goto l97
				l98:
					position, tokenIndex = position97, tokenIndex97
					if buffer[position] != rune('M') {
						goto l89
					}
					position++
				}
			l97:
				if !_rules[rulespace]() {
					goto l89
				}
				add(rulefrom, position90)
			}
			return true
		l89:
			position, tokenIndex = position89, tokenIndex89
			return false
		},
		/* 15 order <- <(('o' / 'O') ('r' / 'R') ('d' / 'D') ('e' / 'E') ('r' / 'R') space)> */
		func() bool {
			position99, tokenIndex99 := position, tokenIndex
			{
				position100 := position
				{
					position101, tokenIndex101 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l102
					}
					position++
					goto l101
				l102:
					position, tokenIndex = position101, tokenIndex101
					if buffer[position] != rune('O') {
						goto l99
					}
					position++
				}
			l101:
				{
					position103, tokenIndex103 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l104
					}
					position++
					goto l103
				l104:
					position, tokenIndex = position103, tokenIndex103
					if buffer[position] != rune('R') {
						goto l99
					}
					position++
				}
			l103:
				{
					position105, tokenIndex105 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l106
					}
					position++
					goto l105
				l106:
					position, tokenIndex = position105, tokenIndex105
					if buffer[position] != rune('D') {
						goto l99
					}
					position++
				}
			l105:
				{
					position107, tokenIndex107 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l108
					}
					position++
					goto l107
				l108:
					position, tokenIndex = position107, tokenIndex107
					if buffer[position] != rune('E') {
						goto l99
					}
					position++
				}
			l107:
				{
					position109, tokenIndex109 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l110
					}
					position++
					goto l109
				l110:
					position, tokenIndex = position109, tokenIndex109
					if buffer[position] != rune('R') {
						goto l99
					}
					position++
				}
			l109:
				if !_rules[rulespace]() {
					goto l99
				}
				add(ruleorder, position100)
			}
			return true
		l99:
			position, tokenIndex = position99, tokenIndex99
			return false
		},
		/* 16 by <- <(('b' / 'B') ('y' / 'Y') space)> */
		func() bool {
			position111, tokenIndex111 := position, tokenIndex
			{
				position112 := position
				{
					position113, tokenIndex113 := position, tokenIndex
					if buffer[position] != rune('b') {
						goto l114
					}
					position++
					goto l113
				l114:
					position, tokenIndex = position113, tokenIndex113
					if buffer[position] != rune('B') {
						goto l111
					}
					position++
				}
			l113:
				{
					position115, tokenIndex115 := position, tokenIndex
					if buffer[position] != rune('y') {
						goto l116
					}
					position++
					goto l115
				l116:
					position, tokenIndex = position115, tokenIndex115
					if buffer[position] != rune('Y') {
						goto l111
					}
					position++
				}
			l115:
				if !_rules[rulespace]() {
					goto l111
				}
				add(ruleby, position112)
			}
			return true
		l111:
			position, tokenIndex = position111, tokenIndex111
			return false
		},
		/* 17 star <- <(('*' / (('a' / 'A') ('l' / 'L') ('l' / 'L'))) space Action16)> */
		func() bool {
			position117, tokenIndex117 := position, tokenIndex
			{
				position118 := position
				{
					position119, tokenIndex119 := position, tokenIndex
					if buffer[position] != rune('*') {
						goto l120
					}
					position++
					goto l119
				l120:
					position, tokenIndex = position119, tokenIndex119
					{
						position121, tokenIndex121 := position, tokenIndex
						if buffer[position] != rune('a') {
							goto l122
						}
						position++
						goto l121
					l122:
						position, tokenIndex = position121, tokenIndex121
						if buffer[position] != rune('A') {
							goto l117
						}
						position++
					}
				l121:
					{
						position123, tokenIndex123 := position, tokenIndex
						if buffer[position] != rune('l') {
							goto l124
						}
						position++
						goto l123
					l124:
						position, tokenIndex = position123, tokenIndex123
						if buffer[position] != rune('L') {
							goto l117
						}
						position++
					}
				l123:
					{
						position125, tokenIndex125 := position, tokenIndex
						if buffer[position] != rune('l') {
							goto l126
						}
						position++
						goto l125
					l126:
						position, tokenIndex = position125, tokenIndex125
						if buffer[position] != rune('L') {
							goto l117
						}
						position++
					}
				l125:
				}
			l119:
				if !_rules[rulespace]() {
					goto l117
				}
				if !_rules[ruleAction16]() {
					goto l117
				}
				add(rulestar, position118)
			}
			return true
		l117:
			position, tokenIndex = position117, tokenIndex117
			return false
		},
		/* 18 limit <- <(('l' / 'L') ('i' / 'I') ('m' / 'M') ('i' / 'I') ('t' / 'T') space)> */
		func() bool {
			position127, tokenIndex127 := position, tokenIndex
			{
				position128 := position
				{
					position129, tokenIndex129 := position, tokenIndex
					if buffer[position] != rune('l') {
						goto l130
					}
					position++
					goto l129
				l130:
					position, tokenIndex = position129, tokenIndex129
					if buffer[position] != rune('L') {
						goto l127
					}
					position++
				}
			l129:
				{
					position131, tokenIndex131 := position, tokenIndex
					if buffer[position] != rune('i') {
						goto l132
					}
					position++
					goto l131
				l132:
					position, tokenIndex = position131, tokenIndex131
					if buffer[position] != rune('I') {
						goto l127
					}
					position++
				}
			l131:
				{
					position133, tokenIndex133 := position, tokenIndex
					if buffer[position] != rune('m') {
						goto l134
					}
					position++
					goto l133
				l134:
					position, tokenIndex = position133, tokenIndex133
					if buffer[position] != rune('M') {
						goto l127
					}
					position++
				}
			l133:
				{
					position135, tokenIndex135 := position, tokenIndex
					if buffer[position] != rune('i') {
						goto l136
					}
					position++
					goto l135
				l136:
					position, tokenIndex = position135, tokenIndex135
					if buffer[position] != rune('I') {
						goto l127
					}
					position++
				}
			l135:
				{
					position137, tokenIndex137 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l138
					}
					position++
					goto l137
				l138:
					position, tokenIndex = position137, tokenIndex137
					if buffer[position] != rune('T') {
						goto l127
					}
					position++
				}
			l137:
				if !_rules[rulespace]() {
					goto l127
				}
				add(rulelimit, position128)
			}
			return true
		l127:
			position, tokenIndex = position127, tokenIndex127
			return false
		},
		/* 19 where <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') space)> */
		func() bool {
			position139, tokenIndex139 := position, tokenIndex
			{
				position140 := position
				{
					position141, tokenIndex141 := position, tokenIndex
					if buffer[position] != rune('w') {
						goto l142
					}
					position++
					goto l141
				l142:
					position, tokenIndex = position141, tokenIndex141
					if buffer[position] != rune('W') {
						goto l139
					}
					position++
				}
			l141:
				{
					position143, tokenIndex143 := position, tokenIndex
					if buffer[position] != rune('h') {
						goto l144
					}
					position++
					goto l143
				l144:
					position, tokenIndex = position143, tokenIndex143
					if buffer[position] != rune('H') {
						goto l139
					}
					position++
				}
			l143:
				{
					position145, tokenIndex145 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l146
					}
					position++
					goto l145
				l146:
					position, tokenIndex = position145, tokenIndex145
					if buffer[position] != rune('E') {
						goto l139
					}
					position++
				}
			l145:
				{
					position147, tokenIndex147 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l148
					}
					position++
					goto l147
				l148:
					position, tokenIndex = position147, tokenIndex147
					if buffer[position] != rune('R') {
						goto l139
					}
					position++
				}
			l147:
				{
					position149, tokenIndex149 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l150
					}
					position++
					goto l149
				l150:
					position, tokenIndex = position149, tokenIndex149
					if buffer[position] != rune('E') {
						goto l139
					}
					position++
				}
			l149:
				if !_rules[rulespace]() {
					goto l139
				}
				add(rulewhere, position140)
			}
			return true
		l139:
			position, tokenIndex = position139, tokenIndex139
			return false
		},
		/* 20 and <- <(('a' / 'A') ('n' / 'N') ('d' / 'D') space)> */
		func() bool {
			position151, tokenIndex151 := position, tokenIndex
			{
				position152 := position
				{
					position153, tokenIndex153 := position, tokenIndex
					if buffer[position] != rune('a') {
						goto l154
					}
					position++
					goto l153
				l154:
					position, tokenIndex = position153, tokenIndex153
					if buffer[position] != rune('A') {
						goto l151
					}
					position++
				}
			l153:
				{
					position155, tokenIndex155 := position, tokenIndex
					if buffer[position] != rune('n') {
						goto l156
					}
					position++
					goto l155
				l156:
					position, tokenIndex = position155, tokenIndex155
					if buffer[position] != rune('N') {
						goto l151
					}
					position++
				}
			l155:
				{
					position157, tokenIndex157 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l158
					}
					position++
					goto l157
				l158:
					position, tokenIndex = position157, tokenIndex157
					if buffer[position] != rune('D') {
						goto l151
					}
					position++
				}
			l157:
				if !_rules[rulespace]() {
					goto l151
				}
				add(ruleand, position152)
			}
			return true
		l151:
			position, tokenIndex = position151, tokenIndex151
			return false
		},
		/* 21 or <- <(('o' / 'O') ('r' / 'R') space)> */
		func() bool {
			position159, tokenIndex159 := position, tokenIndex
			{
				position160 := position
				{
					position161, tokenIndex161 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l162
					}
					position++
					goto l161
				l162:
					position, tokenIndex = position161, tokenIndex161
					if buffer[position] != rune('O') {
						goto l159
					}
					position++
				}
			l161:
				{
					position163, tokenIndex163 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l164
					}
					position++
					goto l163
				l164:
					position, tokenIndex = position163, tokenIndex163
					if buffer[position] != rune('R') {
						goto l159
					}
					position++
				}
			l163:
				if !_rules[rulespace]() {
					goto l159
				}
				add(ruleor, position160)
			}
			return true
		l159:
			position, tokenIndex = position159, tokenIndex159
			return false
		},
		/* 22 inStmt <- <(space? (('i' / 'I') ('n' / 'N')) space?)> */
		func() bool {
			position165, tokenIndex165 := position, tokenIndex
			{
				position166 := position
				{
					position167, tokenIndex167 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l167
					}
					goto l168
				l167:
					position, tokenIndex = position167, tokenIndex167
				}
			l168:
				{
					position169, tokenIndex169 := position, tokenIndex
					if buffer[position] != rune('i') {
						goto l170
					}
					position++
					goto l169
				l170:
					position, tokenIndex = position169, tokenIndex169
					if buffer[position] != rune('I') {
						goto l165
					}
					position++
				}
			l169:
				{
					position171, tokenIndex171 := position, tokenIndex
					if buffer[position] != rune('n') {
						goto l172
					}
					position++
					goto l171
				l172:
					position, tokenIndex = position171, tokenIndex171
					if buffer[position] != rune('N') {
						goto l165
					}
					position++
				}
			l171:
				{
					position173, tokenIndex173 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l173
					}
					goto l174
				l173:
					position, tokenIndex = position173, tokenIndex173
				}
			l174:
				add(ruleinStmt, position166)
			}
			return true
		l165:
			position, tokenIndex = position165, tokenIndex165
			return false
		},
		/* 23 greaterThan <- <(space? '>' space?)> */
		func() bool {
			position175, tokenIndex175 := position, tokenIndex
			{
				position176 := position
				{
					position177, tokenIndex177 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l177
					}
					goto l178
				l177:
					position, tokenIndex = position177, tokenIndex177
				}
			l178:
				if buffer[position] != rune('>') {
					goto l175
				}
				position++
				{
					position179, tokenIndex179 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l179
					}
					goto l180
				l179:
					position, tokenIndex = position179, tokenIndex179
				}
			l180:
				add(rulegreaterThan, position176)
			}
			return true
		l175:
			position, tokenIndex = position175, tokenIndex175
			return false
		},
		/* 24 lessThan <- <(space? '<' space?)> */
		func() bool {
			position181, tokenIndex181 := position, tokenIndex
			{
				position182 := position
				{
					position183, tokenIndex183 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l183
					}
					goto l184
				l183:
					position, tokenIndex = position183, tokenIndex183
				}
			l184:
				if buffer[position] != rune('<') {
					goto l181
				}
				position++
				{
					position185, tokenIndex185 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l185
					}
					goto l186
				l185:
					position, tokenIndex = position185, tokenIndex185
				}
			l186:
				add(rulelessThan, position182)
			}
			return true
		l181:
			position, tokenIndex = position181, tokenIndex181
			return false
		},
		/* 25 greaterThanOrEq <- <(space? ('>' '=') space?)> */
		func() bool {
			position187, tokenIndex187 := position, tokenIndex
			{
				position188 := position
				{
					position189, tokenIndex189 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l189
					}
					goto l190
				l189:
					position, tokenIndex = position189, tokenIndex189
				}
			l190:
				if buffer[position] != rune('>') {
					goto l187
				}
				position++
				if buffer[position] != rune('=') {
					goto l187
				}
				position++
				{
					position191, tokenIndex191 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l191
					}
					goto l192
				l191:
					position, tokenIndex = position191, tokenIndex191
				}
			l192:
				add(rulegreaterThanOrEq, position188)
			}
			return true
		l187:
			position, tokenIndex = position187, tokenIndex187
			return false
		},
		/* 26 lessThanOrEq <- <(space? ('<' '=') space?)> */
		func() bool {
			position193, tokenIndex193 := position, tokenIndex
			{
				position194 := position
				{
					position195, tokenIndex195 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l195
					}
					goto l196
				l195:
					position, tokenIndex = position195, tokenIndex195
				}
			l196:
				if buffer[position] != rune('<') {
					goto l193
				}
				position++
				if buffer[position] != rune('=') {
					goto l193
				}
				position++
				{
					position197, tokenIndex197 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l197
					}
					goto l198
				l197:
					position, tokenIndex = position197, tokenIndex197
				}
			l198:
				add(rulelessThanOrEq, position194)
			}
			return true
		l193:
			position, tokenIndex = position193, tokenIndex193
			return false
		},
		/* 27 notEq <- <(space? ('!' '=') space?)> */
		func() bool {
			position199, tokenIndex199 := position, tokenIndex
			{
				position200 := position
				{
					position201, tokenIndex201 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l201
					}
					goto l202
				l201:
					position, tokenIndex = position201, tokenIndex201
				}
			l202:
				if buffer[position] != rune('!') {
					goto l199
				}
				position++
				if buffer[position] != rune('=') {
					goto l199
				}
				position++
				{
					position203, tokenIndex203 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l203
					}
					goto l204
				l203:
					position, tokenIndex = position203, tokenIndex203
				}
			l204:
				add(rulenotEq, position200)
			}
			return true
		l199:
			position, tokenIndex = position199, tokenIndex199
			return false
		},
		/* 28 semi <- <(';' space?)> */
		func() bool {
			position205, tokenIndex205 := position, tokenIndex
			{
				position206 := position
				if buffer[position] != rune(';') {
					goto l205
				}
				position++
				{
					position207, tokenIndex207 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l207
					}
					goto l208
				l207:
					position, tokenIndex = position207, tokenIndex207
				}
			l208:
				add(rulesemi, position206)
			}
			return true
		l205:
			position, tokenIndex = position205, tokenIndex205
			return false
		},
		/* 29 lparen <- <('(' space?)> */
		func() bool {
			position209, tokenIndex209 := position, tokenIndex
			{
				position210 := position
				if buffer[position] != rune('(') {
					goto l209
				}
				position++
				{
					position211, tokenIndex211 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l211
					}
					goto l212
				l211:
					position, tokenIndex = position211, tokenIndex211
				}
			l212:
				add(rulelparen, position210)
			}
			return true
		l209:
			position, tokenIndex = position209, tokenIndex209
			return false
		},
		/* 30 rparen <- <(space? ')' space?)> */
		func() bool {
			position213, tokenIndex213 := position, tokenIndex
			{
				position214 := position
				{
					position215, tokenIndex215 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l215
					}
					goto l216
				l215:
					position, tokenIndex = position215, tokenIndex215
				}
			l216:
				if buffer[position] != rune(')') {
					goto l213
				}
				position++
				{
					position217, tokenIndex217 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l217
					}
					goto l218
				l217:
					position, tokenIndex = position217, tokenIndex217
				}
			l218:
				add(rulerparen, position214)
			}
			return true
		l213:
			position, tokenIndex = position213, tokenIndex213
			return false
		},
		/* 31 comma <- <(space? ',' space?)> */
		func() bool {
			position219, tokenIndex219 := position, tokenIndex
			{
				position220 := position
				{
					position221, tokenIndex221 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l221
					}
					goto l222
				l221:
					position, tokenIndex = position221, tokenIndex221
				}
			l222:
				if buffer[position] != rune(',') {
					goto l219
				}
				position++
				{
					position223, tokenIndex223 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l223
					}
					goto l224
				l223:
					position, tokenIndex = position223, tokenIndex223
				}
			l224:
				add(rulecomma, position220)
			}
			return true
		l219:
			position, tokenIndex = position219, tokenIndex219
			return false
		},
		/* 32 equals <- <(space? '=' space?)> */
		func() bool {
			position225, tokenIndex225 := position, tokenIndex
			{
				position226 := position
				{
					position227, tokenIndex227 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l227
					}
					goto l228
				l227:
					position, tokenIndex = position227, tokenIndex227
				}
			l228:
				if buffer[position] != rune('=') {
					goto l225
				}
				position++
				{
					position229, tokenIndex229 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l229
					}
					goto l230
				l229:
					position, tokenIndex = position229, tokenIndex229
				}
			l230:
				add(ruleequals, position226)
			}
			return true
		l225:
			position, tokenIndex = position225, tokenIndex225
			return false
		},
		/* 33 squote <- <'\''> */
		func() bool {
			position231, tokenIndex231 := position, tokenIndex
			{
				position232 := position
				if buffer[position] != rune('\'') {
					goto l231
				}
				position++
				add(rulesquote, position232)
			}
			return true
		l231:
			position, tokenIndex = position231, tokenIndex231
			return false
		},
		/* 34 dquote <- <'"'> */
		func() bool {
			position233, tokenIndex233 := position, tokenIndex
			{
				position234 := position
				if buffer[position] != rune('"') {
					goto l233
				}
				position++
				add(ruledquote, position234)
			}
			return true
		l233:
			position, tokenIndex = position233, tokenIndex233
			return false
		},
		/* 35 StringLiteralList <- <(lparen StringLiteral (space? comma space? StringLiteral)* rparen)> */
		func() bool {
			position235, tokenIndex235 := position, tokenIndex
			{
				position236 := position
				if !_rules[rulelparen]() {
					goto l235
				}
				if !_rules[ruleStringLiteral]() {
					goto l235
				}
			l237:
				{
					position238, tokenIndex238 := position, tokenIndex
					{
						position239, tokenIndex239 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l239
						}
						goto l240
					l239:
						position, tokenIndex = position239, tokenIndex239
					}
				l240:
					if !_rules[rulecomma]() {
						goto l238
					}
					{
						position241, tokenIndex241 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l241
						}
						goto l242
					l241:
						position, tokenIndex = position241, tokenIndex241
					}
				l242:
					if !_rules[ruleStringLiteral]() {
						goto l238
					}
					goto l237
				l238:
					position, tokenIndex = position238, tokenIndex238
				}
				if !_rules[rulerparen]() {
					goto l235
				}
				add(ruleStringLiteralList, position236)
			}
			return true
		l235:
			position, tokenIndex = position235, tokenIndex235
			return false
		},
		/* 36 NumericList <- <(lparen DecimalNumeral (space? comma space? DecimalNumeral)* rparen)> */
		func() bool {
			position243, tokenIndex243 := position, tokenIndex
			{
				position244 := position
				if !_rules[rulelparen]() {
					goto l243
				}
				if !_rules[ruleDecimalNumeral]() {
					goto l243
				}
			l245:
				{
					position246, tokenIndex246 := position, tokenIndex
					{
						position247, tokenIndex247 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l247
						}
						goto l248
					l247:
						position, tokenIndex = position247, tokenIndex247
					}
				l248:
					if !_rules[rulecomma]() {
						goto l246
					}
					{
						position249, tokenIndex249 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l249
						}
						goto l250
					l249:
						position, tokenIndex = position249, tokenIndex249
					}
				l250:
					if !_rules[ruleDecimalNumeral]() {
						goto l246
					}
					goto l245
				l246:
					position, tokenIndex = position246, tokenIndex246
				}
				if !_rules[rulerparen]() {
					goto l243
				}
				add(ruleNumericList, position244)
			}
			return true
		l243:
			position, tokenIndex = position243, tokenIndex243
			return false
		},
		/* 37 Letter <- <([a-z] / [A-Z])> */
		func() bool {
			position251, tokenIndex251 := position, tokenIndex
			{
				position252 := position
				{
					position253, tokenIndex253 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l254
					}
					position++
					goto l253
				l254:
					position, tokenIndex = position253, tokenIndex253
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l251
					}
					position++
				}
			l253:
				add(ruleLetter, position252)
			}
			return true
		l251:
			position, tokenIndex = position251, tokenIndex251
			return false
		},
		/* 38 LetterOrDigit <- <([a-z] / [A-Z] / [0-9])> */
		func() bool {
			position255, tokenIndex255 := position, tokenIndex
			{
				position256 := position
				{
					position257, tokenIndex257 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l258
					}
					position++
					goto l257
				l258:
					position, tokenIndex = position257, tokenIndex257
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l259
					}
					position++
					goto l257
				l259:
					position, tokenIndex = position257, tokenIndex257
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l255
					}
					position++
				}
			l257:
				add(ruleLetterOrDigit, position256)
			}
			return true
		l255:
			position, tokenIndex = position255, tokenIndex255
			return false
		},
		/* 39 HexDigit <- <([a-f] / [A-F] / [0-9])> */
		func() bool {
			position260, tokenIndex260 := position, tokenIndex
			{
				position261 := position
				{
					position262, tokenIndex262 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('f') {
						goto l263
					}
					position++
					goto l262
				l263:
					position, tokenIndex = position262, tokenIndex262
					if c := buffer[position]; c < rune('A') || c > rune('F') {
						goto l264
					}
					position++
					goto l262
				l264:
					position, tokenIndex = position262, tokenIndex262
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l260
					}
					position++
				}
			l262:
				add(ruleHexDigit, position261)
			}
			return true
		l260:
			position, tokenIndex = position260, tokenIndex260
			return false
		},
		/* 40 DecimalNumeral <- <<('0' / ([0-9] ('_'* [0-9])*))>> */
		func() bool {
			position265, tokenIndex265 := position, tokenIndex
			{
				position266 := position
				{
					position267 := position
					{
						position268, tokenIndex268 := position, tokenIndex
						if buffer[position] != rune('0') {
							goto l269
						}
						position++
						goto l268
					l269:
						position, tokenIndex = position268, tokenIndex268
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l265
						}
						position++
					l270:
						{
							position271, tokenIndex271 := position, tokenIndex
						l272:
							{
								position273, tokenIndex273 := position, tokenIndex
								if buffer[position] != rune('_') {
									goto l273
								}
								position++
								goto l272
							l273:
								position, tokenIndex = position273, tokenIndex273
							}
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l271
							}
							position++
							goto l270
						l271:
							position, tokenIndex = position271, tokenIndex271
						}
					}
				l268:
					add(rulePegText, position267)
				}
				add(ruleDecimalNumeral, position266)
			}
			return true
		l265:
			position, tokenIndex = position265, tokenIndex265
			return false
		},
		/* 41 StringChar <- <(Escape / (!(' ' / '\'' / '"' / '\n' / '\\') .))> */
		func() bool {
			position274, tokenIndex274 := position, tokenIndex
			{
				position275 := position
				{
					position276, tokenIndex276 := position, tokenIndex
					if !_rules[ruleEscape]() {
						goto l277
					}
					goto l276
				l277:
					position, tokenIndex = position276, tokenIndex276
					{
						position278, tokenIndex278 := position, tokenIndex
						{
							position279, tokenIndex279 := position, tokenIndex
							if buffer[position] != rune(' ') {
								goto l280
							}
							position++
							goto l279
						l280:
							position, tokenIndex = position279, tokenIndex279
							if buffer[position] != rune('\'') {
								goto l281
							}
							position++
							goto l279
						l281:
							position, tokenIndex = position279, tokenIndex279
							if buffer[position] != rune('"') {
								goto l282
							}
							position++
							goto l279
						l282:
							position, tokenIndex = position279, tokenIndex279
							if buffer[position] != rune('\n') {
								goto l283
							}
							position++
							goto l279
						l283:
							position, tokenIndex = position279, tokenIndex279
							if buffer[position] != rune('\\') {
								goto l278
							}
							position++
						}
					l279:
						goto l274
					l278:
						position, tokenIndex = position278, tokenIndex278
					}
					if !matchDot() {
						goto l274
					}
				}
			l276:
				add(ruleStringChar, position275)
			}
			return true
		l274:
			position, tokenIndex = position274, tokenIndex274
			return false
		},
		/* 42 StringLiteral <- <<((squote / dquote)? StringChar* (squote / dquote)?)>> */
		func() bool {
			{
				position285 := position
				{
					position286 := position
					{
						position287, tokenIndex287 := position, tokenIndex
						{
							position289, tokenIndex289 := position, tokenIndex
							if !_rules[rulesquote]() {
								goto l290
							}
							goto l289
						l290:
							position, tokenIndex = position289, tokenIndex289
							if !_rules[ruledquote]() {
								goto l287
							}
						}
					l289:
						goto l288
					l287:
						position, tokenIndex = position287, tokenIndex287
					}
				l288:
				l291:
					{
						position292, tokenIndex292 := position, tokenIndex
						if !_rules[ruleStringChar]() {
							goto l292
						}
						goto l291
					l292:
						position, tokenIndex = position292, tokenIndex292
					}
					{
						position293, tokenIndex293 := position, tokenIndex
						{
							position295, tokenIndex295 := position, tokenIndex
							if !_rules[rulesquote]() {
								goto l296
							}
							goto l295
						l296:
							position, tokenIndex = position295, tokenIndex295
							if !_rules[ruledquote]() {
								goto l293
							}
						}
					l295:
						goto l294
					l293:
						position, tokenIndex = position293, tokenIndex293
					}
				l294:
					add(rulePegText, position286)
				}
				add(ruleStringLiteral, position285)
			}
			return true
		},
		/* 43 Escape <- <('\\' ('b' / 't' / 'n' / 'f' / 'r' / '"' / '\'' / '\\' / OctalEscape / UnicodeEscape))> */
		func() bool {
			position297, tokenIndex297 := position, tokenIndex
			{
				position298 := position
				if buffer[position] != rune('\\') {
					goto l297
				}
				position++
				{
					position299, tokenIndex299 := position, tokenIndex
					if buffer[position] != rune('b') {
						goto l300
					}
					position++
					goto l299
				l300:
					position, tokenIndex = position299, tokenIndex299
					if buffer[position] != rune('t') {
						goto l301
					}
					position++
					goto l299
				l301:
					position, tokenIndex = position299, tokenIndex299
					if buffer[position] != rune('n') {
						goto l302
					}
					position++
					goto l299
				l302:
					position, tokenIndex = position299, tokenIndex299
					if buffer[position] != rune('f') {
						goto l303
					}
					position++
					goto l299
				l303:
					position, tokenIndex = position299, tokenIndex299
					if buffer[position] != rune('r') {
						goto l304
					}
					position++
					goto l299
				l304:
					position, tokenIndex = position299, tokenIndex299
					if buffer[position] != rune('"') {
						goto l305
					}
					position++
					goto l299
				l305:
					position, tokenIndex = position299, tokenIndex299
					if buffer[position] != rune('\'') {
						goto l306
					}
					position++
					goto l299
				l306:
					position, tokenIndex = position299, tokenIndex299
					if buffer[position] != rune('\\') {
						goto l307
					}
					position++
					goto l299
				l307:
					position, tokenIndex = position299, tokenIndex299
					if !_rules[ruleOctalEscape]() {
						goto l308
					}
					goto l299
				l308:
					position, tokenIndex = position299, tokenIndex299
					if !_rules[ruleUnicodeEscape]() {
						goto l297
					}
				}
			l299:
				add(ruleEscape, position298)
			}
			return true
		l297:
			position, tokenIndex = position297, tokenIndex297
			return false
		},
		/* 44 OctalEscape <- <(([0-3] [0-7] [0-7]) / ([0-7] [0-7]) / [0-7])> */
		func() bool {
			position309, tokenIndex309 := position, tokenIndex
			{
				position310 := position
				{
					position311, tokenIndex311 := position, tokenIndex
					if c := buffer[position]; c < rune('0') || c > rune('3') {
						goto l312
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l312
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l312
					}
					position++
					goto l311
				l312:
					position, tokenIndex = position311, tokenIndex311
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l313
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l313
					}
					position++
					goto l311
				l313:
					position, tokenIndex = position311, tokenIndex311
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l309
					}
					position++
				}
			l311:
				add(ruleOctalEscape, position310)
			}
			return true
		l309:
			position, tokenIndex = position309, tokenIndex309
			return false
		},
		/* 45 UnicodeEscape <- <('u'+ HexDigit HexDigit HexDigit HexDigit)> */
		func() bool {
			position314, tokenIndex314 := position, tokenIndex
			{
				position315 := position
				if buffer[position] != rune('u') {
					goto l314
				}
				position++
			l316:
				{
					position317, tokenIndex317 := position, tokenIndex
					if buffer[position] != rune('u') {
						goto l317
					}
					position++
					goto l316
				l317:
					position, tokenIndex = position317, tokenIndex317
				}
				if !_rules[ruleHexDigit]() {
					goto l314
				}
				if !_rules[ruleHexDigit]() {
					goto l314
				}
				if !_rules[ruleHexDigit]() {
					goto l314
				}
				if !_rules[ruleHexDigit]() {
					goto l314
				}
				add(ruleUnicodeEscape, position315)
			}
			return true
		l314:
			position, tokenIndex = position314, tokenIndex314
			return false
		},
		/* 46 WordList <- <(Word Action17 (space ',' space? WordList)*)> */
		func() bool {
			position318, tokenIndex318 := position, tokenIndex
			{
				position319 := position
				if !_rules[ruleWord]() {
					goto l318
				}
				if !_rules[ruleAction17]() {
					goto l318
				}
			l320:
				{
					position321, tokenIndex321 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l321
					}
					if buffer[position] != rune(',') {
						goto l321
					}
					position++
					{
						position322, tokenIndex322 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l322
						}
						goto l323
					l322:
						position, tokenIndex = position322, tokenIndex322
					}
				l323:
					if !_rules[ruleWordList]() {
						goto l321
					}
					goto l320
				l321:
					position, tokenIndex = position321, tokenIndex321
				}
				add(ruleWordList, position319)
			}
			return true
		l318:
			position, tokenIndex = position318, tokenIndex318
			return false
		},
		/* 47 Word <- <(<(Letter LetterOrDigit*)> Action18)> */
		func() bool {
			position324, tokenIndex324 := position, tokenIndex
			{
				position325 := position
				{
					position326 := position
					if !_rules[ruleLetter]() {
						goto l324
					}
				l327:
					{
						position328, tokenIndex328 := position, tokenIndex
						if !_rules[ruleLetterOrDigit]() {
							goto l328
						}
						goto l327
					l328:
						position, tokenIndex = position328, tokenIndex328
					}
					add(rulePegText, position326)
				}
				if !_rules[ruleAction18]() {
					goto l324
				}
				add(ruleWord, position325)
			}
			return true
		l324:
			position, tokenIndex = position324, tokenIndex324
			return false
		},
		/* 48 space <- <((' ' / '\t' / '\r' / '\n')+ / ('/' '*' (!('*' '/') .)* ('*' '/')) / ('/' '/' (!('\r' / '\n') .)* ('\r' / '\n')))*> */
		func() bool {
			{
				position330 := position
			l331:
				{
					position332, tokenIndex332 := position, tokenIndex
					{
						position333, tokenIndex333 := position, tokenIndex
						{
							position337, tokenIndex337 := position, tokenIndex
							if buffer[position] != rune(' ') {
								goto l338
							}
							position++
							goto l337
						l338:
							position, tokenIndex = position337, tokenIndex337
							if buffer[position] != rune('\t') {
								goto l339
							}
							position++
							goto l337
						l339:
							position, tokenIndex = position337, tokenIndex337
							if buffer[position] != rune('\r') {
								goto l340
							}
							position++
							goto l337
						l340:
							position, tokenIndex = position337, tokenIndex337
							if buffer[position] != rune('\n') {
								goto l334
							}
							position++
						}
					l337:
					l335:
						{
							position336, tokenIndex336 := position, tokenIndex
							{
								position341, tokenIndex341 := position, tokenIndex
								if buffer[position] != rune(' ') {
									goto l342
								}
								position++
								goto l341
							l342:
								position, tokenIndex = position341, tokenIndex341
								if buffer[position] != rune('\t') {
									goto l343
								}
								position++
								goto l341
							l343:
								position, tokenIndex = position341, tokenIndex341
								if buffer[position] != rune('\r') {
									goto l344
								}
								position++
								goto l341
							l344:
								position, tokenIndex = position341, tokenIndex341
								if buffer[position] != rune('\n') {
									goto l336
								}
								position++
							}
						l341:
							goto l335
						l336:
							position, tokenIndex = position336, tokenIndex336
						}
						goto l333
					l334:
						position, tokenIndex = position333, tokenIndex333
						if buffer[position] != rune('/') {
							goto l345
						}
						position++
						if buffer[position] != rune('*') {
							goto l345
						}
						position++
					l346:
						{
							position347, tokenIndex347 := position, tokenIndex
							{
								position348, tokenIndex348 := position, tokenIndex
								if buffer[position] != rune('*') {
									goto l348
								}
								position++
								if buffer[position] != rune('/') {
									goto l348
								}
								position++
								goto l347
							l348:
								position, tokenIndex = position348, tokenIndex348
							}
							if !matchDot() {
								goto l347
							}
							goto l346
						l347:
							position, tokenIndex = position347, tokenIndex347
						}
						if buffer[position] != rune('*') {
							goto l345
						}
						position++
						if buffer[position] != rune('/') {
							goto l345
						}
						position++
						goto l333
					l345:
						position, tokenIndex = position333, tokenIndex333
						if buffer[position] != rune('/') {
							goto l332
						}
						position++
						if buffer[position] != rune('/') {
							goto l332
						}
						position++
					l349:
						{
							position350, tokenIndex350 := position, tokenIndex
							{
								position351, tokenIndex351 := position, tokenIndex
								{
									position352, tokenIndex352 := position, tokenIndex
									if buffer[position] != rune('\r') {
										goto l353
									}
									position++
									goto l352
								l353:
									position, tokenIndex = position352, tokenIndex352
									if buffer[position] != rune('\n') {
										goto l351
									}
									position++
								}
							l352:
								goto l350
							l351:
								position, tokenIndex = position351, tokenIndex351
							}
							if !matchDot() {
								goto l350
							}
							goto l349
						l350:
							position, tokenIndex = position350, tokenIndex350
						}
						{
							position354, tokenIndex354 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l355
							}
							position++
							goto l354
						l355:
							position, tokenIndex = position354, tokenIndex354
							if buffer[position] != rune('\n') {
								goto l332
							}
							position++
						}
					l354:
					}
				l333:
					goto l331
				l332:
					position, tokenIndex = position332, tokenIndex332
				}
				add(rulespace, position330)
			}
			return true
		},
		/* 50 Action0 <- <{
		  p.Fields = p.strStack; p.strStack = []string{};
		}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 51 Action1 <- <{ p.Table = buffer[begin:end] }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 52 Action2 <- <{
		  f, _ := strconv.Atoi(text)
		  p.Limit.ShouldLimit = true
		  p.Limit.Count = f
		}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 53 Action3 <- <{ p.Order.Field = text }> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 54 Action4 <- <{ p.Order.Ordering = ASC }> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 55 Action5 <- <{ p.Order.Ordering = DESC }> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 56 Action6 <- <{
		   p.ExprStack.Push(&NodeAnd{
		     left: p.ExprStack.Pop(),
		     right: p.ExprStack.Pop(),
		   })
		 }> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 57 Action7 <- <{}> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 58 Action8 <- <{
		   p.ExprStack.Push(&NodeEquals{
		     left: p.workingWord,
		     right: text,
		   })
		 }> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 59 Action9 <- <{
		   p.ExprStack.Push(&NodeEquals{
		     left: p.workingWord,
		     right: text,
		   })
		 }> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		/* 60 Action10 <- <{
		   f, _ := strconv.Atoi(text)
		   p.ExprStack.Push(&NodeGreaterThan{
		     left: p.workingWord,
		     right: f,
		   })
		 }> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 61 Action11 <- <{
		   f, _ := strconv.Atoi(text)
		   p.ExprStack.Push(&NodeGreaterThanOrEqual{
		     left: p.workingWord,
		     right: f,
		   })
		 }> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 62 Action12 <- <{
		   f, _ := strconv.Atoi(text)
		   p.ExprStack.Push(&NodeLessThan{
		     left: p.workingWord,
		     right: f,
		   })
		 }> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 63 Action13 <- <{
		   f, _ := strconv.Atoi(text)
		   p.ExprStack.Push(&NodeLessThanOrEqual{
		     left: p.workingWord,
		     right: f,
		   })
		 }> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 64 Action14 <- <{
		   p.ExprStack.Push(&NodeNotEqual{
		     left: p.workingWord,
		     right: text,
		   })
		 }> */
		func() bool {
			{
				add(ruleAction14, position)
			}
			return true
		},
		/* 65 Action15 <- <{
		   p.ExprStack.Push(&NodeNotEqual{
		     left: p.workingWord,
		     right: text,
		   })
		 }> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 66 Action16 <- <{ text = "star" }> */
		func() bool {
			{
				add(ruleAction16, position)
			}
			return true
		},
		nil,
		/* 68 Action17 <- <{
		   p.strStack = append(p.strStack, text)
		 }> */
		func() bool {
			{
				add(ruleAction17, position)
			}
			return true
		},
		/* 69 Action18 <- <{ p.workingWord = text }> */
		func() bool {
			{
				add(ruleAction18, position)
			}
			return true
		},
	}
	p.rules = _rules
}
