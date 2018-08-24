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
	ruleselect
	rulefrom
	ruleorder
	ruleby
	rulestar
	rulelimit
	rulewhere
	ruleand
	ruleor
	rulegreaterThan
	ruleequals
	rulesquote
	ruledquote
	ruleLetter
	ruleLetterOrDigit
	ruleHexDigit
	ruleDecimalNumeral
	ruleStringLiteral
	ruleEscape
	ruleOctalEscape
	ruleUnicodeEscape
	ruleWordList
	ruleWord
	rulespace
	rulePegText
	ruleAction0
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
	"select",
	"from",
	"order",
	"by",
	"star",
	"limit",
	"where",
	"and",
	"or",
	"greaterThan",
	"equals",
	"squote",
	"dquote",
	"Letter",
	"LetterOrDigit",
	"HexDigit",
	"DecimalNumeral",
	"StringLiteral",
	"Escape",
	"OctalEscape",
	"UnicodeEscape",
	"WordList",
	"Word",
	"space",
	"PegText",
	"Action0",
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
	Buffer string
	buffer []rune
	rules  [38]func() bool
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
		/* 0 Result <- <(queryStmt !.)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[rulequeryStmt]() {
					goto l0
				}
				{
					position2, tokenIndex2 := position, tokenIndex
					if !matchDot() {
						goto l2
					}
					goto l0
				l2:
					position, tokenIndex = position2, tokenIndex2
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
			position3, tokenIndex3 := position, tokenIndex
			{
				position4 := position
				if !_rules[ruleselectStmt]() {
					goto l3
				}
			l5:
				{
					position6, tokenIndex6 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l6
					}
					if !_rules[rulequeryExprs]() {
						goto l6
					}
					goto l5
				l6:
					position, tokenIndex = position6, tokenIndex6
				}
				add(rulequeryStmt, position4)
			}
			return true
		l3:
			position, tokenIndex = position3, tokenIndex3
			return false
		},
		/* 2 queryExprs <- <(queryExpr (space queryExpr)*)> */
		func() bool {
			position7, tokenIndex7 := position, tokenIndex
			{
				position8 := position
				if !_rules[rulequeryExpr]() {
					goto l7
				}
			l9:
				{
					position10, tokenIndex10 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l10
					}
					if !_rules[rulequeryExpr]() {
						goto l10
					}
					goto l9
				l10:
					position, tokenIndex = position10, tokenIndex10
				}
				add(rulequeryExprs, position8)
			}
			return true
		l7:
			position, tokenIndex = position7, tokenIndex7
			return false
		},
		/* 3 queryExpr <- <(limitStmt / orderStmt / whereStmt)> */
		func() bool {
			position11, tokenIndex11 := position, tokenIndex
			{
				position12 := position
				{
					position13, tokenIndex13 := position, tokenIndex
					if !_rules[rulelimitStmt]() {
						goto l14
					}
					goto l13
				l14:
					position, tokenIndex = position13, tokenIndex13
					if !_rules[ruleorderStmt]() {
						goto l15
					}
					goto l13
				l15:
					position, tokenIndex = position13, tokenIndex13
					if !_rules[rulewhereStmt]() {
						goto l11
					}
				}
			l13:
				add(rulequeryExpr, position12)
			}
			return true
		l11:
			position, tokenIndex = position11, tokenIndex11
			return false
		},
		/* 4 selectStmt <- <(select space (star / WordList) space fromStmt)> */
		func() bool {
			position16, tokenIndex16 := position, tokenIndex
			{
				position17 := position
				if !_rules[ruleselect]() {
					goto l16
				}
				if !_rules[rulespace]() {
					goto l16
				}
				{
					position18, tokenIndex18 := position, tokenIndex
					if !_rules[rulestar]() {
						goto l19
					}
					goto l18
				l19:
					position, tokenIndex = position18, tokenIndex18
					if !_rules[ruleWordList]() {
						goto l16
					}
				}
			l18:
				if !_rules[rulespace]() {
					goto l16
				}
				if !_rules[rulefromStmt]() {
					goto l16
				}
				add(ruleselectStmt, position17)
			}
			return true
		l16:
			position, tokenIndex = position16, tokenIndex16
			return false
		},
		/* 5 fromStmt <- <(from Word)> */
		func() bool {
			position20, tokenIndex20 := position, tokenIndex
			{
				position21 := position
				if !_rules[rulefrom]() {
					goto l20
				}
				if !_rules[ruleWord]() {
					goto l20
				}
				add(rulefromStmt, position21)
			}
			return true
		l20:
			position, tokenIndex = position20, tokenIndex20
			return false
		},
		/* 6 limitStmt <- <(limit DecimalNumeral)> */
		func() bool {
			position22, tokenIndex22 := position, tokenIndex
			{
				position23 := position
				if !_rules[rulelimit]() {
					goto l22
				}
				if !_rules[ruleDecimalNumeral]() {
					goto l22
				}
				add(rulelimitStmt, position23)
			}
			return true
		l22:
			position, tokenIndex = position22, tokenIndex22
			return false
		},
		/* 7 orderStmt <- <(order by WordList)> */
		func() bool {
			position24, tokenIndex24 := position, tokenIndex
			{
				position25 := position
				if !_rules[ruleorder]() {
					goto l24
				}
				if !_rules[ruleby]() {
					goto l24
				}
				if !_rules[ruleWordList]() {
					goto l24
				}
				add(ruleorderStmt, position25)
			}
			return true
		l24:
			position, tokenIndex = position24, tokenIndex24
			return false
		},
		/* 8 whereStmt <- <(where whereExprs)> */
		func() bool {
			position26, tokenIndex26 := position, tokenIndex
			{
				position27 := position
				if !_rules[rulewhere]() {
					goto l26
				}
				if !_rules[rulewhereExprs]() {
					goto l26
				}
				add(rulewhereStmt, position27)
			}
			return true
		l26:
			position, tokenIndex = position26, tokenIndex26
			return false
		},
		/* 9 whereExprs <- <(whereExpr (space (and / or) whereExpr)*)> */
		func() bool {
			position28, tokenIndex28 := position, tokenIndex
			{
				position29 := position
				if !_rules[rulewhereExpr]() {
					goto l28
				}
			l30:
				{
					position31, tokenIndex31 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l31
					}
					{
						position32, tokenIndex32 := position, tokenIndex
						if !_rules[ruleand]() {
							goto l33
						}
						goto l32
					l33:
						position, tokenIndex = position32, tokenIndex32
						if !_rules[ruleor]() {
							goto l31
						}
					}
				l32:
					if !_rules[rulewhereExpr]() {
						goto l31
					}
					goto l30
				l31:
					position, tokenIndex = position31, tokenIndex31
				}
				add(rulewhereExprs, position29)
			}
			return true
		l28:
			position, tokenIndex = position28, tokenIndex28
			return false
		},
		/* 10 whereExpr <- <((Word equals StringLiteral) / (Word greaterThan DecimalNumeral))> */
		func() bool {
			position34, tokenIndex34 := position, tokenIndex
			{
				position35 := position
				{
					position36, tokenIndex36 := position, tokenIndex
					if !_rules[ruleWord]() {
						goto l37
					}
					if !_rules[ruleequals]() {
						goto l37
					}
					if !_rules[ruleStringLiteral]() {
						goto l37
					}
					goto l36
				l37:
					position, tokenIndex = position36, tokenIndex36
					if !_rules[ruleWord]() {
						goto l34
					}
					if !_rules[rulegreaterThan]() {
						goto l34
					}
					if !_rules[ruleDecimalNumeral]() {
						goto l34
					}
				}
			l36:
				add(rulewhereExpr, position35)
			}
			return true
		l34:
			position, tokenIndex = position34, tokenIndex34
			return false
		},
		/* 11 select <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') space)> */
		func() bool {
			position38, tokenIndex38 := position, tokenIndex
			{
				position39 := position
				{
					position40, tokenIndex40 := position, tokenIndex
					if buffer[position] != rune('s') {
						goto l41
					}
					position++
					goto l40
				l41:
					position, tokenIndex = position40, tokenIndex40
					if buffer[position] != rune('S') {
						goto l38
					}
					position++
				}
			l40:
				{
					position42, tokenIndex42 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l43
					}
					position++
					goto l42
				l43:
					position, tokenIndex = position42, tokenIndex42
					if buffer[position] != rune('E') {
						goto l38
					}
					position++
				}
			l42:
				{
					position44, tokenIndex44 := position, tokenIndex
					if buffer[position] != rune('l') {
						goto l45
					}
					position++
					goto l44
				l45:
					position, tokenIndex = position44, tokenIndex44
					if buffer[position] != rune('L') {
						goto l38
					}
					position++
				}
			l44:
				{
					position46, tokenIndex46 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l47
					}
					position++
					goto l46
				l47:
					position, tokenIndex = position46, tokenIndex46
					if buffer[position] != rune('E') {
						goto l38
					}
					position++
				}
			l46:
				{
					position48, tokenIndex48 := position, tokenIndex
					if buffer[position] != rune('c') {
						goto l49
					}
					position++
					goto l48
				l49:
					position, tokenIndex = position48, tokenIndex48
					if buffer[position] != rune('C') {
						goto l38
					}
					position++
				}
			l48:
				{
					position50, tokenIndex50 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l51
					}
					position++
					goto l50
				l51:
					position, tokenIndex = position50, tokenIndex50
					if buffer[position] != rune('T') {
						goto l38
					}
					position++
				}
			l50:
				if !_rules[rulespace]() {
					goto l38
				}
				add(ruleselect, position39)
			}
			return true
		l38:
			position, tokenIndex = position38, tokenIndex38
			return false
		},
		/* 12 from <- <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M') space)> */
		func() bool {
			position52, tokenIndex52 := position, tokenIndex
			{
				position53 := position
				{
					position54, tokenIndex54 := position, tokenIndex
					if buffer[position] != rune('f') {
						goto l55
					}
					position++
					goto l54
				l55:
					position, tokenIndex = position54, tokenIndex54
					if buffer[position] != rune('F') {
						goto l52
					}
					position++
				}
			l54:
				{
					position56, tokenIndex56 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l57
					}
					position++
					goto l56
				l57:
					position, tokenIndex = position56, tokenIndex56
					if buffer[position] != rune('R') {
						goto l52
					}
					position++
				}
			l56:
				{
					position58, tokenIndex58 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l59
					}
					position++
					goto l58
				l59:
					position, tokenIndex = position58, tokenIndex58
					if buffer[position] != rune('O') {
						goto l52
					}
					position++
				}
			l58:
				{
					position60, tokenIndex60 := position, tokenIndex
					if buffer[position] != rune('m') {
						goto l61
					}
					position++
					goto l60
				l61:
					position, tokenIndex = position60, tokenIndex60
					if buffer[position] != rune('M') {
						goto l52
					}
					position++
				}
			l60:
				if !_rules[rulespace]() {
					goto l52
				}
				add(rulefrom, position53)
			}
			return true
		l52:
			position, tokenIndex = position52, tokenIndex52
			return false
		},
		/* 13 order <- <(('o' / 'O') ('r' / 'R') ('d' / 'D') ('e' / 'E') ('r' / 'R') space)> */
		func() bool {
			position62, tokenIndex62 := position, tokenIndex
			{
				position63 := position
				{
					position64, tokenIndex64 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l65
					}
					position++
					goto l64
				l65:
					position, tokenIndex = position64, tokenIndex64
					if buffer[position] != rune('O') {
						goto l62
					}
					position++
				}
			l64:
				{
					position66, tokenIndex66 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l67
					}
					position++
					goto l66
				l67:
					position, tokenIndex = position66, tokenIndex66
					if buffer[position] != rune('R') {
						goto l62
					}
					position++
				}
			l66:
				{
					position68, tokenIndex68 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l69
					}
					position++
					goto l68
				l69:
					position, tokenIndex = position68, tokenIndex68
					if buffer[position] != rune('D') {
						goto l62
					}
					position++
				}
			l68:
				{
					position70, tokenIndex70 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l71
					}
					position++
					goto l70
				l71:
					position, tokenIndex = position70, tokenIndex70
					if buffer[position] != rune('E') {
						goto l62
					}
					position++
				}
			l70:
				{
					position72, tokenIndex72 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l73
					}
					position++
					goto l72
				l73:
					position, tokenIndex = position72, tokenIndex72
					if buffer[position] != rune('R') {
						goto l62
					}
					position++
				}
			l72:
				if !_rules[rulespace]() {
					goto l62
				}
				add(ruleorder, position63)
			}
			return true
		l62:
			position, tokenIndex = position62, tokenIndex62
			return false
		},
		/* 14 by <- <(('b' / 'B') ('y' / 'Y') space)> */
		func() bool {
			position74, tokenIndex74 := position, tokenIndex
			{
				position75 := position
				{
					position76, tokenIndex76 := position, tokenIndex
					if buffer[position] != rune('b') {
						goto l77
					}
					position++
					goto l76
				l77:
					position, tokenIndex = position76, tokenIndex76
					if buffer[position] != rune('B') {
						goto l74
					}
					position++
				}
			l76:
				{
					position78, tokenIndex78 := position, tokenIndex
					if buffer[position] != rune('y') {
						goto l79
					}
					position++
					goto l78
				l79:
					position, tokenIndex = position78, tokenIndex78
					if buffer[position] != rune('Y') {
						goto l74
					}
					position++
				}
			l78:
				if !_rules[rulespace]() {
					goto l74
				}
				add(ruleby, position75)
			}
			return true
		l74:
			position, tokenIndex = position74, tokenIndex74
			return false
		},
		/* 15 star <- <(('*' / (('a' / 'A') ('l' / 'L') ('l' / 'L'))) space)> */
		func() bool {
			position80, tokenIndex80 := position, tokenIndex
			{
				position81 := position
				{
					position82, tokenIndex82 := position, tokenIndex
					if buffer[position] != rune('*') {
						goto l83
					}
					position++
					goto l82
				l83:
					position, tokenIndex = position82, tokenIndex82
					{
						position84, tokenIndex84 := position, tokenIndex
						if buffer[position] != rune('a') {
							goto l85
						}
						position++
						goto l84
					l85:
						position, tokenIndex = position84, tokenIndex84
						if buffer[position] != rune('A') {
							goto l80
						}
						position++
					}
				l84:
					{
						position86, tokenIndex86 := position, tokenIndex
						if buffer[position] != rune('l') {
							goto l87
						}
						position++
						goto l86
					l87:
						position, tokenIndex = position86, tokenIndex86
						if buffer[position] != rune('L') {
							goto l80
						}
						position++
					}
				l86:
					{
						position88, tokenIndex88 := position, tokenIndex
						if buffer[position] != rune('l') {
							goto l89
						}
						position++
						goto l88
					l89:
						position, tokenIndex = position88, tokenIndex88
						if buffer[position] != rune('L') {
							goto l80
						}
						position++
					}
				l88:
				}
			l82:
				if !_rules[rulespace]() {
					goto l80
				}
				add(rulestar, position81)
			}
			return true
		l80:
			position, tokenIndex = position80, tokenIndex80
			return false
		},
		/* 16 limit <- <(('l' / 'L') ('i' / 'I') ('m' / 'M') ('i' / 'I') ('t' / 'T') space)> */
		func() bool {
			position90, tokenIndex90 := position, tokenIndex
			{
				position91 := position
				{
					position92, tokenIndex92 := position, tokenIndex
					if buffer[position] != rune('l') {
						goto l93
					}
					position++
					goto l92
				l93:
					position, tokenIndex = position92, tokenIndex92
					if buffer[position] != rune('L') {
						goto l90
					}
					position++
				}
			l92:
				{
					position94, tokenIndex94 := position, tokenIndex
					if buffer[position] != rune('i') {
						goto l95
					}
					position++
					goto l94
				l95:
					position, tokenIndex = position94, tokenIndex94
					if buffer[position] != rune('I') {
						goto l90
					}
					position++
				}
			l94:
				{
					position96, tokenIndex96 := position, tokenIndex
					if buffer[position] != rune('m') {
						goto l97
					}
					position++
					goto l96
				l97:
					position, tokenIndex = position96, tokenIndex96
					if buffer[position] != rune('M') {
						goto l90
					}
					position++
				}
			l96:
				{
					position98, tokenIndex98 := position, tokenIndex
					if buffer[position] != rune('i') {
						goto l99
					}
					position++
					goto l98
				l99:
					position, tokenIndex = position98, tokenIndex98
					if buffer[position] != rune('I') {
						goto l90
					}
					position++
				}
			l98:
				{
					position100, tokenIndex100 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l101
					}
					position++
					goto l100
				l101:
					position, tokenIndex = position100, tokenIndex100
					if buffer[position] != rune('T') {
						goto l90
					}
					position++
				}
			l100:
				if !_rules[rulespace]() {
					goto l90
				}
				add(rulelimit, position91)
			}
			return true
		l90:
			position, tokenIndex = position90, tokenIndex90
			return false
		},
		/* 17 where <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') space)> */
		func() bool {
			position102, tokenIndex102 := position, tokenIndex
			{
				position103 := position
				{
					position104, tokenIndex104 := position, tokenIndex
					if buffer[position] != rune('w') {
						goto l105
					}
					position++
					goto l104
				l105:
					position, tokenIndex = position104, tokenIndex104
					if buffer[position] != rune('W') {
						goto l102
					}
					position++
				}
			l104:
				{
					position106, tokenIndex106 := position, tokenIndex
					if buffer[position] != rune('h') {
						goto l107
					}
					position++
					goto l106
				l107:
					position, tokenIndex = position106, tokenIndex106
					if buffer[position] != rune('H') {
						goto l102
					}
					position++
				}
			l106:
				{
					position108, tokenIndex108 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l109
					}
					position++
					goto l108
				l109:
					position, tokenIndex = position108, tokenIndex108
					if buffer[position] != rune('E') {
						goto l102
					}
					position++
				}
			l108:
				{
					position110, tokenIndex110 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l111
					}
					position++
					goto l110
				l111:
					position, tokenIndex = position110, tokenIndex110
					if buffer[position] != rune('R') {
						goto l102
					}
					position++
				}
			l110:
				{
					position112, tokenIndex112 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l113
					}
					position++
					goto l112
				l113:
					position, tokenIndex = position112, tokenIndex112
					if buffer[position] != rune('E') {
						goto l102
					}
					position++
				}
			l112:
				if !_rules[rulespace]() {
					goto l102
				}
				add(rulewhere, position103)
			}
			return true
		l102:
			position, tokenIndex = position102, tokenIndex102
			return false
		},
		/* 18 and <- <(('a' / 'A') ('n' / 'N') ('d' / 'D') space)> */
		func() bool {
			position114, tokenIndex114 := position, tokenIndex
			{
				position115 := position
				{
					position116, tokenIndex116 := position, tokenIndex
					if buffer[position] != rune('a') {
						goto l117
					}
					position++
					goto l116
				l117:
					position, tokenIndex = position116, tokenIndex116
					if buffer[position] != rune('A') {
						goto l114
					}
					position++
				}
			l116:
				{
					position118, tokenIndex118 := position, tokenIndex
					if buffer[position] != rune('n') {
						goto l119
					}
					position++
					goto l118
				l119:
					position, tokenIndex = position118, tokenIndex118
					if buffer[position] != rune('N') {
						goto l114
					}
					position++
				}
			l118:
				{
					position120, tokenIndex120 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l121
					}
					position++
					goto l120
				l121:
					position, tokenIndex = position120, tokenIndex120
					if buffer[position] != rune('D') {
						goto l114
					}
					position++
				}
			l120:
				if !_rules[rulespace]() {
					goto l114
				}
				add(ruleand, position115)
			}
			return true
		l114:
			position, tokenIndex = position114, tokenIndex114
			return false
		},
		/* 19 or <- <(('o' / 'O') ('r' / 'R') space)> */
		func() bool {
			position122, tokenIndex122 := position, tokenIndex
			{
				position123 := position
				{
					position124, tokenIndex124 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l125
					}
					position++
					goto l124
				l125:
					position, tokenIndex = position124, tokenIndex124
					if buffer[position] != rune('O') {
						goto l122
					}
					position++
				}
			l124:
				{
					position126, tokenIndex126 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l127
					}
					position++
					goto l126
				l127:
					position, tokenIndex = position126, tokenIndex126
					if buffer[position] != rune('R') {
						goto l122
					}
					position++
				}
			l126:
				if !_rules[rulespace]() {
					goto l122
				}
				add(ruleor, position123)
			}
			return true
		l122:
			position, tokenIndex = position122, tokenIndex122
			return false
		},
		/* 20 greaterThan <- <(space? '>' space?)> */
		func() bool {
			position128, tokenIndex128 := position, tokenIndex
			{
				position129 := position
				{
					position130, tokenIndex130 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l130
					}
					goto l131
				l130:
					position, tokenIndex = position130, tokenIndex130
				}
			l131:
				if buffer[position] != rune('>') {
					goto l128
				}
				position++
				{
					position132, tokenIndex132 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l132
					}
					goto l133
				l132:
					position, tokenIndex = position132, tokenIndex132
				}
			l133:
				add(rulegreaterThan, position129)
			}
			return true
		l128:
			position, tokenIndex = position128, tokenIndex128
			return false
		},
		/* 21 equals <- <(space? '=' space?)> */
		func() bool {
			position134, tokenIndex134 := position, tokenIndex
			{
				position135 := position
				{
					position136, tokenIndex136 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l136
					}
					goto l137
				l136:
					position, tokenIndex = position136, tokenIndex136
				}
			l137:
				if buffer[position] != rune('=') {
					goto l134
				}
				position++
				{
					position138, tokenIndex138 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l138
					}
					goto l139
				l138:
					position, tokenIndex = position138, tokenIndex138
				}
			l139:
				add(ruleequals, position135)
			}
			return true
		l134:
			position, tokenIndex = position134, tokenIndex134
			return false
		},
		/* 22 squote <- <'\''> */
		func() bool {
			position140, tokenIndex140 := position, tokenIndex
			{
				position141 := position
				if buffer[position] != rune('\'') {
					goto l140
				}
				position++
				add(rulesquote, position141)
			}
			return true
		l140:
			position, tokenIndex = position140, tokenIndex140
			return false
		},
		/* 23 dquote <- <'"'> */
		func() bool {
			position142, tokenIndex142 := position, tokenIndex
			{
				position143 := position
				if buffer[position] != rune('"') {
					goto l142
				}
				position++
				add(ruledquote, position143)
			}
			return true
		l142:
			position, tokenIndex = position142, tokenIndex142
			return false
		},
		/* 24 Letter <- <([a-z] / [A-Z])> */
		func() bool {
			position144, tokenIndex144 := position, tokenIndex
			{
				position145 := position
				{
					position146, tokenIndex146 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l147
					}
					position++
					goto l146
				l147:
					position, tokenIndex = position146, tokenIndex146
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l144
					}
					position++
				}
			l146:
				add(ruleLetter, position145)
			}
			return true
		l144:
			position, tokenIndex = position144, tokenIndex144
			return false
		},
		/* 25 LetterOrDigit <- <([a-z] / [A-Z] / [0-9])> */
		func() bool {
			position148, tokenIndex148 := position, tokenIndex
			{
				position149 := position
				{
					position150, tokenIndex150 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l151
					}
					position++
					goto l150
				l151:
					position, tokenIndex = position150, tokenIndex150
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l152
					}
					position++
					goto l150
				l152:
					position, tokenIndex = position150, tokenIndex150
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l148
					}
					position++
				}
			l150:
				add(ruleLetterOrDigit, position149)
			}
			return true
		l148:
			position, tokenIndex = position148, tokenIndex148
			return false
		},
		/* 26 HexDigit <- <([a-f] / [A-F] / [0-9])> */
		func() bool {
			position153, tokenIndex153 := position, tokenIndex
			{
				position154 := position
				{
					position155, tokenIndex155 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('f') {
						goto l156
					}
					position++
					goto l155
				l156:
					position, tokenIndex = position155, tokenIndex155
					if c := buffer[position]; c < rune('A') || c > rune('F') {
						goto l157
					}
					position++
					goto l155
				l157:
					position, tokenIndex = position155, tokenIndex155
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l153
					}
					position++
				}
			l155:
				add(ruleHexDigit, position154)
			}
			return true
		l153:
			position, tokenIndex = position153, tokenIndex153
			return false
		},
		/* 27 DecimalNumeral <- <('0' / ([0-9] ('_'* [0-9])*))> */
		func() bool {
			position158, tokenIndex158 := position, tokenIndex
			{
				position159 := position
				{
					position160, tokenIndex160 := position, tokenIndex
					if buffer[position] != rune('0') {
						goto l161
					}
					position++
					goto l160
				l161:
					position, tokenIndex = position160, tokenIndex160
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l158
					}
					position++
				l162:
					{
						position163, tokenIndex163 := position, tokenIndex
					l164:
						{
							position165, tokenIndex165 := position, tokenIndex
							if buffer[position] != rune('_') {
								goto l165
							}
							position++
							goto l164
						l165:
							position, tokenIndex = position165, tokenIndex165
						}
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l163
						}
						position++
						goto l162
					l163:
						position, tokenIndex = position163, tokenIndex163
					}
				}
			l160:
				add(ruleDecimalNumeral, position159)
			}
			return true
		l158:
			position, tokenIndex = position158, tokenIndex158
			return false
		},
		/* 28 StringLiteral <- <((squote / dquote)? (Escape / (!('"' / '\\' / '\n' / '\r') .))* (squote / dquote)?)> */
		func() bool {
			{
				position167 := position
				{
					position168, tokenIndex168 := position, tokenIndex
					{
						position170, tokenIndex170 := position, tokenIndex
						if !_rules[rulesquote]() {
							goto l171
						}
						goto l170
					l171:
						position, tokenIndex = position170, tokenIndex170
						if !_rules[ruledquote]() {
							goto l168
						}
					}
				l170:
					goto l169
				l168:
					position, tokenIndex = position168, tokenIndex168
				}
			l169:
			l172:
				{
					position173, tokenIndex173 := position, tokenIndex
					{
						position174, tokenIndex174 := position, tokenIndex
						if !_rules[ruleEscape]() {
							goto l175
						}
						goto l174
					l175:
						position, tokenIndex = position174, tokenIndex174
						{
							position176, tokenIndex176 := position, tokenIndex
							{
								position177, tokenIndex177 := position, tokenIndex
								if buffer[position] != rune('"') {
									goto l178
								}
								position++
								goto l177
							l178:
								position, tokenIndex = position177, tokenIndex177
								if buffer[position] != rune('\\') {
									goto l179
								}
								position++
								goto l177
							l179:
								position, tokenIndex = position177, tokenIndex177
								if buffer[position] != rune('\n') {
									goto l180
								}
								position++
								goto l177
							l180:
								position, tokenIndex = position177, tokenIndex177
								if buffer[position] != rune('\r') {
									goto l176
								}
								position++
							}
						l177:
							goto l173
						l176:
							position, tokenIndex = position176, tokenIndex176
						}
						if !matchDot() {
							goto l173
						}
					}
				l174:
					goto l172
				l173:
					position, tokenIndex = position173, tokenIndex173
				}
				{
					position181, tokenIndex181 := position, tokenIndex
					{
						position183, tokenIndex183 := position, tokenIndex
						if !_rules[rulesquote]() {
							goto l184
						}
						goto l183
					l184:
						position, tokenIndex = position183, tokenIndex183
						if !_rules[ruledquote]() {
							goto l181
						}
					}
				l183:
					goto l182
				l181:
					position, tokenIndex = position181, tokenIndex181
				}
			l182:
				add(ruleStringLiteral, position167)
			}
			return true
		},
		/* 29 Escape <- <('\\' ('b' / 't' / 'n' / 'f' / 'r' / '"' / '\'' / '\\' / OctalEscape / UnicodeEscape))> */
		func() bool {
			position185, tokenIndex185 := position, tokenIndex
			{
				position186 := position
				if buffer[position] != rune('\\') {
					goto l185
				}
				position++
				{
					position187, tokenIndex187 := position, tokenIndex
					if buffer[position] != rune('b') {
						goto l188
					}
					position++
					goto l187
				l188:
					position, tokenIndex = position187, tokenIndex187
					if buffer[position] != rune('t') {
						goto l189
					}
					position++
					goto l187
				l189:
					position, tokenIndex = position187, tokenIndex187
					if buffer[position] != rune('n') {
						goto l190
					}
					position++
					goto l187
				l190:
					position, tokenIndex = position187, tokenIndex187
					if buffer[position] != rune('f') {
						goto l191
					}
					position++
					goto l187
				l191:
					position, tokenIndex = position187, tokenIndex187
					if buffer[position] != rune('r') {
						goto l192
					}
					position++
					goto l187
				l192:
					position, tokenIndex = position187, tokenIndex187
					if buffer[position] != rune('"') {
						goto l193
					}
					position++
					goto l187
				l193:
					position, tokenIndex = position187, tokenIndex187
					if buffer[position] != rune('\'') {
						goto l194
					}
					position++
					goto l187
				l194:
					position, tokenIndex = position187, tokenIndex187
					if buffer[position] != rune('\\') {
						goto l195
					}
					position++
					goto l187
				l195:
					position, tokenIndex = position187, tokenIndex187
					if !_rules[ruleOctalEscape]() {
						goto l196
					}
					goto l187
				l196:
					position, tokenIndex = position187, tokenIndex187
					if !_rules[ruleUnicodeEscape]() {
						goto l185
					}
				}
			l187:
				add(ruleEscape, position186)
			}
			return true
		l185:
			position, tokenIndex = position185, tokenIndex185
			return false
		},
		/* 30 OctalEscape <- <(([0-3] [0-7] [0-7]) / ([0-7] [0-7]) / [0-7])> */
		func() bool {
			position197, tokenIndex197 := position, tokenIndex
			{
				position198 := position
				{
					position199, tokenIndex199 := position, tokenIndex
					if c := buffer[position]; c < rune('0') || c > rune('3') {
						goto l200
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l200
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l200
					}
					position++
					goto l199
				l200:
					position, tokenIndex = position199, tokenIndex199
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l201
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l201
					}
					position++
					goto l199
				l201:
					position, tokenIndex = position199, tokenIndex199
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l197
					}
					position++
				}
			l199:
				add(ruleOctalEscape, position198)
			}
			return true
		l197:
			position, tokenIndex = position197, tokenIndex197
			return false
		},
		/* 31 UnicodeEscape <- <('u'+ HexDigit HexDigit HexDigit HexDigit)> */
		func() bool {
			position202, tokenIndex202 := position, tokenIndex
			{
				position203 := position
				if buffer[position] != rune('u') {
					goto l202
				}
				position++
			l204:
				{
					position205, tokenIndex205 := position, tokenIndex
					if buffer[position] != rune('u') {
						goto l205
					}
					position++
					goto l204
				l205:
					position, tokenIndex = position205, tokenIndex205
				}
				if !_rules[ruleHexDigit]() {
					goto l202
				}
				if !_rules[ruleHexDigit]() {
					goto l202
				}
				if !_rules[ruleHexDigit]() {
					goto l202
				}
				if !_rules[ruleHexDigit]() {
					goto l202
				}
				add(ruleUnicodeEscape, position203)
			}
			return true
		l202:
			position, tokenIndex = position202, tokenIndex202
			return false
		},
		/* 32 WordList <- <(Word (space ',' space? WordList)*)> */
		func() bool {
			position206, tokenIndex206 := position, tokenIndex
			{
				position207 := position
				if !_rules[ruleWord]() {
					goto l206
				}
			l208:
				{
					position209, tokenIndex209 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l209
					}
					if buffer[position] != rune(',') {
						goto l209
					}
					position++
					{
						position210, tokenIndex210 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l210
						}
						goto l211
					l210:
						position, tokenIndex = position210, tokenIndex210
					}
				l211:
					if !_rules[ruleWordList]() {
						goto l209
					}
					goto l208
				l209:
					position, tokenIndex = position209, tokenIndex209
				}
				add(ruleWordList, position207)
			}
			return true
		l206:
			position, tokenIndex = position206, tokenIndex206
			return false
		},
		/* 33 Word <- <((squote / dquote)? <(Letter LetterOrDigit*)> (squote / dquote)? Action0)> */
		func() bool {
			position212, tokenIndex212 := position, tokenIndex
			{
				position213 := position
				{
					position214, tokenIndex214 := position, tokenIndex
					{
						position216, tokenIndex216 := position, tokenIndex
						if !_rules[rulesquote]() {
							goto l217
						}
						goto l216
					l217:
						position, tokenIndex = position216, tokenIndex216
						if !_rules[ruledquote]() {
							goto l214
						}
					}
				l216:
					goto l215
				l214:
					position, tokenIndex = position214, tokenIndex214
				}
			l215:
				{
					position218 := position
					if !_rules[ruleLetter]() {
						goto l212
					}
				l219:
					{
						position220, tokenIndex220 := position, tokenIndex
						if !_rules[ruleLetterOrDigit]() {
							goto l220
						}
						goto l219
					l220:
						position, tokenIndex = position220, tokenIndex220
					}
					add(rulePegText, position218)
				}
				{
					position221, tokenIndex221 := position, tokenIndex
					{
						position223, tokenIndex223 := position, tokenIndex
						if !_rules[rulesquote]() {
							goto l224
						}
						goto l223
					l224:
						position, tokenIndex = position223, tokenIndex223
						if !_rules[ruledquote]() {
							goto l221
						}
					}
				l223:
					goto l222
				l221:
					position, tokenIndex = position221, tokenIndex221
				}
			l222:
				if !_rules[ruleAction0]() {
					goto l212
				}
				add(ruleWord, position213)
			}
			return true
		l212:
			position, tokenIndex = position212, tokenIndex212
			return false
		},
		/* 34 space <- <((' ' / '\t' / '\r' / '\n')+ / ('/' '*' (!('*' '/') .)* ('*' '/')) / ('/' '/' (!('\r' / '\n') .)* ('\r' / '\n')))*> */
		func() bool {
			{
				position226 := position
			l227:
				{
					position228, tokenIndex228 := position, tokenIndex
					{
						position229, tokenIndex229 := position, tokenIndex
						{
							position233, tokenIndex233 := position, tokenIndex
							if buffer[position] != rune(' ') {
								goto l234
							}
							position++
							goto l233
						l234:
							position, tokenIndex = position233, tokenIndex233
							if buffer[position] != rune('\t') {
								goto l235
							}
							position++
							goto l233
						l235:
							position, tokenIndex = position233, tokenIndex233
							if buffer[position] != rune('\r') {
								goto l236
							}
							position++
							goto l233
						l236:
							position, tokenIndex = position233, tokenIndex233
							if buffer[position] != rune('\n') {
								goto l230
							}
							position++
						}
					l233:
					l231:
						{
							position232, tokenIndex232 := position, tokenIndex
							{
								position237, tokenIndex237 := position, tokenIndex
								if buffer[position] != rune(' ') {
									goto l238
								}
								position++
								goto l237
							l238:
								position, tokenIndex = position237, tokenIndex237
								if buffer[position] != rune('\t') {
									goto l239
								}
								position++
								goto l237
							l239:
								position, tokenIndex = position237, tokenIndex237
								if buffer[position] != rune('\r') {
									goto l240
								}
								position++
								goto l237
							l240:
								position, tokenIndex = position237, tokenIndex237
								if buffer[position] != rune('\n') {
									goto l232
								}
								position++
							}
						l237:
							goto l231
						l232:
							position, tokenIndex = position232, tokenIndex232
						}
						goto l229
					l230:
						position, tokenIndex = position229, tokenIndex229
						if buffer[position] != rune('/') {
							goto l241
						}
						position++
						if buffer[position] != rune('*') {
							goto l241
						}
						position++
					l242:
						{
							position243, tokenIndex243 := position, tokenIndex
							{
								position244, tokenIndex244 := position, tokenIndex
								if buffer[position] != rune('*') {
									goto l244
								}
								position++
								if buffer[position] != rune('/') {
									goto l244
								}
								position++
								goto l243
							l244:
								position, tokenIndex = position244, tokenIndex244
							}
							if !matchDot() {
								goto l243
							}
							goto l242
						l243:
							position, tokenIndex = position243, tokenIndex243
						}
						if buffer[position] != rune('*') {
							goto l241
						}
						position++
						if buffer[position] != rune('/') {
							goto l241
						}
						position++
						goto l229
					l241:
						position, tokenIndex = position229, tokenIndex229
						if buffer[position] != rune('/') {
							goto l228
						}
						position++
						if buffer[position] != rune('/') {
							goto l228
						}
						position++
					l245:
						{
							position246, tokenIndex246 := position, tokenIndex
							{
								position247, tokenIndex247 := position, tokenIndex
								{
									position248, tokenIndex248 := position, tokenIndex
									if buffer[position] != rune('\r') {
										goto l249
									}
									position++
									goto l248
								l249:
									position, tokenIndex = position248, tokenIndex248
									if buffer[position] != rune('\n') {
										goto l247
									}
									position++
								}
							l248:
								goto l246
							l247:
								position, tokenIndex = position247, tokenIndex247
							}
							if !matchDot() {
								goto l246
							}
							goto l245
						l246:
							position, tokenIndex = position246, tokenIndex246
						}
						{
							position250, tokenIndex250 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l251
							}
							position++
							goto l250
						l251:
							position, tokenIndex = position250, tokenIndex250
							if buffer[position] != rune('\n') {
								goto l228
							}
							position++
						}
					l250:
					}
				l229:
					goto l227
				l228:
					position, tokenIndex = position228, tokenIndex228
				}
				add(rulespace, position226)
			}
			return true
		},
		nil,
		/* 37 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
	}
	p.rules = _rules
}
