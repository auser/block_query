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
	rulelessThan
	ruleequals
	rulesquote
	ruledquote
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
	"lessThan",
	"equals",
	"squote",
	"dquote",
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
	rules  [40]func() bool
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
		/* 2 queryExprs <- <(queryExpr (space queryExprs)*)> */
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
					if !_rules[rulequeryExprs]() {
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
		/* 9 whereExprs <- <(whereExpr (space? (and / or) space? whereExprs)*)> */
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
					{
						position32, tokenIndex32 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l32
						}
						goto l33
					l32:
						position, tokenIndex = position32, tokenIndex32
					}
				l33:
					{
						position34, tokenIndex34 := position, tokenIndex
						if !_rules[ruleand]() {
							goto l35
						}
						goto l34
					l35:
						position, tokenIndex = position34, tokenIndex34
						if !_rules[ruleor]() {
							goto l31
						}
					}
				l34:
					{
						position36, tokenIndex36 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l36
						}
						goto l37
					l36:
						position, tokenIndex = position36, tokenIndex36
					}
				l37:
					if !_rules[rulewhereExprs]() {
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
		/* 10 whereExpr <- <((Word equals StringLiteral) / (Word greaterThan DecimalNumeral) / (Word lessThan DecimalNumeral))> */
		func() bool {
			position38, tokenIndex38 := position, tokenIndex
			{
				position39 := position
				{
					position40, tokenIndex40 := position, tokenIndex
					if !_rules[ruleWord]() {
						goto l41
					}
					if !_rules[ruleequals]() {
						goto l41
					}
					if !_rules[ruleStringLiteral]() {
						goto l41
					}
					goto l40
				l41:
					position, tokenIndex = position40, tokenIndex40
					if !_rules[ruleWord]() {
						goto l42
					}
					if !_rules[rulegreaterThan]() {
						goto l42
					}
					if !_rules[ruleDecimalNumeral]() {
						goto l42
					}
					goto l40
				l42:
					position, tokenIndex = position40, tokenIndex40
					if !_rules[ruleWord]() {
						goto l38
					}
					if !_rules[rulelessThan]() {
						goto l38
					}
					if !_rules[ruleDecimalNumeral]() {
						goto l38
					}
				}
			l40:
				add(rulewhereExpr, position39)
			}
			return true
		l38:
			position, tokenIndex = position38, tokenIndex38
			return false
		},
		/* 11 select <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') space)> */
		func() bool {
			position43, tokenIndex43 := position, tokenIndex
			{
				position44 := position
				{
					position45, tokenIndex45 := position, tokenIndex
					if buffer[position] != rune('s') {
						goto l46
					}
					position++
					goto l45
				l46:
					position, tokenIndex = position45, tokenIndex45
					if buffer[position] != rune('S') {
						goto l43
					}
					position++
				}
			l45:
				{
					position47, tokenIndex47 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l48
					}
					position++
					goto l47
				l48:
					position, tokenIndex = position47, tokenIndex47
					if buffer[position] != rune('E') {
						goto l43
					}
					position++
				}
			l47:
				{
					position49, tokenIndex49 := position, tokenIndex
					if buffer[position] != rune('l') {
						goto l50
					}
					position++
					goto l49
				l50:
					position, tokenIndex = position49, tokenIndex49
					if buffer[position] != rune('L') {
						goto l43
					}
					position++
				}
			l49:
				{
					position51, tokenIndex51 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l52
					}
					position++
					goto l51
				l52:
					position, tokenIndex = position51, tokenIndex51
					if buffer[position] != rune('E') {
						goto l43
					}
					position++
				}
			l51:
				{
					position53, tokenIndex53 := position, tokenIndex
					if buffer[position] != rune('c') {
						goto l54
					}
					position++
					goto l53
				l54:
					position, tokenIndex = position53, tokenIndex53
					if buffer[position] != rune('C') {
						goto l43
					}
					position++
				}
			l53:
				{
					position55, tokenIndex55 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l56
					}
					position++
					goto l55
				l56:
					position, tokenIndex = position55, tokenIndex55
					if buffer[position] != rune('T') {
						goto l43
					}
					position++
				}
			l55:
				if !_rules[rulespace]() {
					goto l43
				}
				add(ruleselect, position44)
			}
			return true
		l43:
			position, tokenIndex = position43, tokenIndex43
			return false
		},
		/* 12 from <- <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M') space)> */
		func() bool {
			position57, tokenIndex57 := position, tokenIndex
			{
				position58 := position
				{
					position59, tokenIndex59 := position, tokenIndex
					if buffer[position] != rune('f') {
						goto l60
					}
					position++
					goto l59
				l60:
					position, tokenIndex = position59, tokenIndex59
					if buffer[position] != rune('F') {
						goto l57
					}
					position++
				}
			l59:
				{
					position61, tokenIndex61 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l62
					}
					position++
					goto l61
				l62:
					position, tokenIndex = position61, tokenIndex61
					if buffer[position] != rune('R') {
						goto l57
					}
					position++
				}
			l61:
				{
					position63, tokenIndex63 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l64
					}
					position++
					goto l63
				l64:
					position, tokenIndex = position63, tokenIndex63
					if buffer[position] != rune('O') {
						goto l57
					}
					position++
				}
			l63:
				{
					position65, tokenIndex65 := position, tokenIndex
					if buffer[position] != rune('m') {
						goto l66
					}
					position++
					goto l65
				l66:
					position, tokenIndex = position65, tokenIndex65
					if buffer[position] != rune('M') {
						goto l57
					}
					position++
				}
			l65:
				if !_rules[rulespace]() {
					goto l57
				}
				add(rulefrom, position58)
			}
			return true
		l57:
			position, tokenIndex = position57, tokenIndex57
			return false
		},
		/* 13 order <- <(('o' / 'O') ('r' / 'R') ('d' / 'D') ('e' / 'E') ('r' / 'R') space)> */
		func() bool {
			position67, tokenIndex67 := position, tokenIndex
			{
				position68 := position
				{
					position69, tokenIndex69 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l70
					}
					position++
					goto l69
				l70:
					position, tokenIndex = position69, tokenIndex69
					if buffer[position] != rune('O') {
						goto l67
					}
					position++
				}
			l69:
				{
					position71, tokenIndex71 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l72
					}
					position++
					goto l71
				l72:
					position, tokenIndex = position71, tokenIndex71
					if buffer[position] != rune('R') {
						goto l67
					}
					position++
				}
			l71:
				{
					position73, tokenIndex73 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l74
					}
					position++
					goto l73
				l74:
					position, tokenIndex = position73, tokenIndex73
					if buffer[position] != rune('D') {
						goto l67
					}
					position++
				}
			l73:
				{
					position75, tokenIndex75 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l76
					}
					position++
					goto l75
				l76:
					position, tokenIndex = position75, tokenIndex75
					if buffer[position] != rune('E') {
						goto l67
					}
					position++
				}
			l75:
				{
					position77, tokenIndex77 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l78
					}
					position++
					goto l77
				l78:
					position, tokenIndex = position77, tokenIndex77
					if buffer[position] != rune('R') {
						goto l67
					}
					position++
				}
			l77:
				if !_rules[rulespace]() {
					goto l67
				}
				add(ruleorder, position68)
			}
			return true
		l67:
			position, tokenIndex = position67, tokenIndex67
			return false
		},
		/* 14 by <- <(('b' / 'B') ('y' / 'Y') space)> */
		func() bool {
			position79, tokenIndex79 := position, tokenIndex
			{
				position80 := position
				{
					position81, tokenIndex81 := position, tokenIndex
					if buffer[position] != rune('b') {
						goto l82
					}
					position++
					goto l81
				l82:
					position, tokenIndex = position81, tokenIndex81
					if buffer[position] != rune('B') {
						goto l79
					}
					position++
				}
			l81:
				{
					position83, tokenIndex83 := position, tokenIndex
					if buffer[position] != rune('y') {
						goto l84
					}
					position++
					goto l83
				l84:
					position, tokenIndex = position83, tokenIndex83
					if buffer[position] != rune('Y') {
						goto l79
					}
					position++
				}
			l83:
				if !_rules[rulespace]() {
					goto l79
				}
				add(ruleby, position80)
			}
			return true
		l79:
			position, tokenIndex = position79, tokenIndex79
			return false
		},
		/* 15 star <- <(('*' / (('a' / 'A') ('l' / 'L') ('l' / 'L'))) space)> */
		func() bool {
			position85, tokenIndex85 := position, tokenIndex
			{
				position86 := position
				{
					position87, tokenIndex87 := position, tokenIndex
					if buffer[position] != rune('*') {
						goto l88
					}
					position++
					goto l87
				l88:
					position, tokenIndex = position87, tokenIndex87
					{
						position89, tokenIndex89 := position, tokenIndex
						if buffer[position] != rune('a') {
							goto l90
						}
						position++
						goto l89
					l90:
						position, tokenIndex = position89, tokenIndex89
						if buffer[position] != rune('A') {
							goto l85
						}
						position++
					}
				l89:
					{
						position91, tokenIndex91 := position, tokenIndex
						if buffer[position] != rune('l') {
							goto l92
						}
						position++
						goto l91
					l92:
						position, tokenIndex = position91, tokenIndex91
						if buffer[position] != rune('L') {
							goto l85
						}
						position++
					}
				l91:
					{
						position93, tokenIndex93 := position, tokenIndex
						if buffer[position] != rune('l') {
							goto l94
						}
						position++
						goto l93
					l94:
						position, tokenIndex = position93, tokenIndex93
						if buffer[position] != rune('L') {
							goto l85
						}
						position++
					}
				l93:
				}
			l87:
				if !_rules[rulespace]() {
					goto l85
				}
				add(rulestar, position86)
			}
			return true
		l85:
			position, tokenIndex = position85, tokenIndex85
			return false
		},
		/* 16 limit <- <(('l' / 'L') ('i' / 'I') ('m' / 'M') ('i' / 'I') ('t' / 'T') space)> */
		func() bool {
			position95, tokenIndex95 := position, tokenIndex
			{
				position96 := position
				{
					position97, tokenIndex97 := position, tokenIndex
					if buffer[position] != rune('l') {
						goto l98
					}
					position++
					goto l97
				l98:
					position, tokenIndex = position97, tokenIndex97
					if buffer[position] != rune('L') {
						goto l95
					}
					position++
				}
			l97:
				{
					position99, tokenIndex99 := position, tokenIndex
					if buffer[position] != rune('i') {
						goto l100
					}
					position++
					goto l99
				l100:
					position, tokenIndex = position99, tokenIndex99
					if buffer[position] != rune('I') {
						goto l95
					}
					position++
				}
			l99:
				{
					position101, tokenIndex101 := position, tokenIndex
					if buffer[position] != rune('m') {
						goto l102
					}
					position++
					goto l101
				l102:
					position, tokenIndex = position101, tokenIndex101
					if buffer[position] != rune('M') {
						goto l95
					}
					position++
				}
			l101:
				{
					position103, tokenIndex103 := position, tokenIndex
					if buffer[position] != rune('i') {
						goto l104
					}
					position++
					goto l103
				l104:
					position, tokenIndex = position103, tokenIndex103
					if buffer[position] != rune('I') {
						goto l95
					}
					position++
				}
			l103:
				{
					position105, tokenIndex105 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l106
					}
					position++
					goto l105
				l106:
					position, tokenIndex = position105, tokenIndex105
					if buffer[position] != rune('T') {
						goto l95
					}
					position++
				}
			l105:
				if !_rules[rulespace]() {
					goto l95
				}
				add(rulelimit, position96)
			}
			return true
		l95:
			position, tokenIndex = position95, tokenIndex95
			return false
		},
		/* 17 where <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') space)> */
		func() bool {
			position107, tokenIndex107 := position, tokenIndex
			{
				position108 := position
				{
					position109, tokenIndex109 := position, tokenIndex
					if buffer[position] != rune('w') {
						goto l110
					}
					position++
					goto l109
				l110:
					position, tokenIndex = position109, tokenIndex109
					if buffer[position] != rune('W') {
						goto l107
					}
					position++
				}
			l109:
				{
					position111, tokenIndex111 := position, tokenIndex
					if buffer[position] != rune('h') {
						goto l112
					}
					position++
					goto l111
				l112:
					position, tokenIndex = position111, tokenIndex111
					if buffer[position] != rune('H') {
						goto l107
					}
					position++
				}
			l111:
				{
					position113, tokenIndex113 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l114
					}
					position++
					goto l113
				l114:
					position, tokenIndex = position113, tokenIndex113
					if buffer[position] != rune('E') {
						goto l107
					}
					position++
				}
			l113:
				{
					position115, tokenIndex115 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l116
					}
					position++
					goto l115
				l116:
					position, tokenIndex = position115, tokenIndex115
					if buffer[position] != rune('R') {
						goto l107
					}
					position++
				}
			l115:
				{
					position117, tokenIndex117 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l118
					}
					position++
					goto l117
				l118:
					position, tokenIndex = position117, tokenIndex117
					if buffer[position] != rune('E') {
						goto l107
					}
					position++
				}
			l117:
				if !_rules[rulespace]() {
					goto l107
				}
				add(rulewhere, position108)
			}
			return true
		l107:
			position, tokenIndex = position107, tokenIndex107
			return false
		},
		/* 18 and <- <(('a' / 'A') ('n' / 'N') ('d' / 'D') space)> */
		func() bool {
			position119, tokenIndex119 := position, tokenIndex
			{
				position120 := position
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
						goto l119
					}
					position++
				}
			l121:
				{
					position123, tokenIndex123 := position, tokenIndex
					if buffer[position] != rune('n') {
						goto l124
					}
					position++
					goto l123
				l124:
					position, tokenIndex = position123, tokenIndex123
					if buffer[position] != rune('N') {
						goto l119
					}
					position++
				}
			l123:
				{
					position125, tokenIndex125 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l126
					}
					position++
					goto l125
				l126:
					position, tokenIndex = position125, tokenIndex125
					if buffer[position] != rune('D') {
						goto l119
					}
					position++
				}
			l125:
				if !_rules[rulespace]() {
					goto l119
				}
				add(ruleand, position120)
			}
			return true
		l119:
			position, tokenIndex = position119, tokenIndex119
			return false
		},
		/* 19 or <- <(('o' / 'O') ('r' / 'R') space)> */
		func() bool {
			position127, tokenIndex127 := position, tokenIndex
			{
				position128 := position
				{
					position129, tokenIndex129 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l130
					}
					position++
					goto l129
				l130:
					position, tokenIndex = position129, tokenIndex129
					if buffer[position] != rune('O') {
						goto l127
					}
					position++
				}
			l129:
				{
					position131, tokenIndex131 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l132
					}
					position++
					goto l131
				l132:
					position, tokenIndex = position131, tokenIndex131
					if buffer[position] != rune('R') {
						goto l127
					}
					position++
				}
			l131:
				if !_rules[rulespace]() {
					goto l127
				}
				add(ruleor, position128)
			}
			return true
		l127:
			position, tokenIndex = position127, tokenIndex127
			return false
		},
		/* 20 greaterThan <- <(space? '>' space?)> */
		func() bool {
			position133, tokenIndex133 := position, tokenIndex
			{
				position134 := position
				{
					position135, tokenIndex135 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l135
					}
					goto l136
				l135:
					position, tokenIndex = position135, tokenIndex135
				}
			l136:
				if buffer[position] != rune('>') {
					goto l133
				}
				position++
				{
					position137, tokenIndex137 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l137
					}
					goto l138
				l137:
					position, tokenIndex = position137, tokenIndex137
				}
			l138:
				add(rulegreaterThan, position134)
			}
			return true
		l133:
			position, tokenIndex = position133, tokenIndex133
			return false
		},
		/* 21 lessThan <- <(space? '<' space?)> */
		func() bool {
			position139, tokenIndex139 := position, tokenIndex
			{
				position140 := position
				{
					position141, tokenIndex141 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l141
					}
					goto l142
				l141:
					position, tokenIndex = position141, tokenIndex141
				}
			l142:
				if buffer[position] != rune('<') {
					goto l139
				}
				position++
				{
					position143, tokenIndex143 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l143
					}
					goto l144
				l143:
					position, tokenIndex = position143, tokenIndex143
				}
			l144:
				add(rulelessThan, position140)
			}
			return true
		l139:
			position, tokenIndex = position139, tokenIndex139
			return false
		},
		/* 22 equals <- <(space? '=' space?)> */
		func() bool {
			position145, tokenIndex145 := position, tokenIndex
			{
				position146 := position
				{
					position147, tokenIndex147 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l147
					}
					goto l148
				l147:
					position, tokenIndex = position147, tokenIndex147
				}
			l148:
				if buffer[position] != rune('=') {
					goto l145
				}
				position++
				{
					position149, tokenIndex149 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l149
					}
					goto l150
				l149:
					position, tokenIndex = position149, tokenIndex149
				}
			l150:
				add(ruleequals, position146)
			}
			return true
		l145:
			position, tokenIndex = position145, tokenIndex145
			return false
		},
		/* 23 squote <- <'\''> */
		func() bool {
			position151, tokenIndex151 := position, tokenIndex
			{
				position152 := position
				if buffer[position] != rune('\'') {
					goto l151
				}
				position++
				add(rulesquote, position152)
			}
			return true
		l151:
			position, tokenIndex = position151, tokenIndex151
			return false
		},
		/* 24 dquote <- <'"'> */
		func() bool {
			position153, tokenIndex153 := position, tokenIndex
			{
				position154 := position
				if buffer[position] != rune('"') {
					goto l153
				}
				position++
				add(ruledquote, position154)
			}
			return true
		l153:
			position, tokenIndex = position153, tokenIndex153
			return false
		},
		/* 25 Letter <- <([a-z] / [A-Z])> */
		func() bool {
			position155, tokenIndex155 := position, tokenIndex
			{
				position156 := position
				{
					position157, tokenIndex157 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l158
					}
					position++
					goto l157
				l158:
					position, tokenIndex = position157, tokenIndex157
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l155
					}
					position++
				}
			l157:
				add(ruleLetter, position156)
			}
			return true
		l155:
			position, tokenIndex = position155, tokenIndex155
			return false
		},
		/* 26 LetterOrDigit <- <([a-z] / [A-Z] / [0-9])> */
		func() bool {
			position159, tokenIndex159 := position, tokenIndex
			{
				position160 := position
				{
					position161, tokenIndex161 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l162
					}
					position++
					goto l161
				l162:
					position, tokenIndex = position161, tokenIndex161
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l163
					}
					position++
					goto l161
				l163:
					position, tokenIndex = position161, tokenIndex161
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l159
					}
					position++
				}
			l161:
				add(ruleLetterOrDigit, position160)
			}
			return true
		l159:
			position, tokenIndex = position159, tokenIndex159
			return false
		},
		/* 27 HexDigit <- <([a-f] / [A-F] / [0-9])> */
		func() bool {
			position164, tokenIndex164 := position, tokenIndex
			{
				position165 := position
				{
					position166, tokenIndex166 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('f') {
						goto l167
					}
					position++
					goto l166
				l167:
					position, tokenIndex = position166, tokenIndex166
					if c := buffer[position]; c < rune('A') || c > rune('F') {
						goto l168
					}
					position++
					goto l166
				l168:
					position, tokenIndex = position166, tokenIndex166
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l164
					}
					position++
				}
			l166:
				add(ruleHexDigit, position165)
			}
			return true
		l164:
			position, tokenIndex = position164, tokenIndex164
			return false
		},
		/* 28 DecimalNumeral <- <('0' / ([0-9] ('_'* [0-9])*))> */
		func() bool {
			position169, tokenIndex169 := position, tokenIndex
			{
				position170 := position
				{
					position171, tokenIndex171 := position, tokenIndex
					if buffer[position] != rune('0') {
						goto l172
					}
					position++
					goto l171
				l172:
					position, tokenIndex = position171, tokenIndex171
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l169
					}
					position++
				l173:
					{
						position174, tokenIndex174 := position, tokenIndex
					l175:
						{
							position176, tokenIndex176 := position, tokenIndex
							if buffer[position] != rune('_') {
								goto l176
							}
							position++
							goto l175
						l176:
							position, tokenIndex = position176, tokenIndex176
						}
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l174
						}
						position++
						goto l173
					l174:
						position, tokenIndex = position174, tokenIndex174
					}
				}
			l171:
				add(ruleDecimalNumeral, position170)
			}
			return true
		l169:
			position, tokenIndex = position169, tokenIndex169
			return false
		},
		/* 29 StringChar <- <(Escape / (!(' ' / '\'' / '"' / '\n' / '\\') .))> */
		func() bool {
			position177, tokenIndex177 := position, tokenIndex
			{
				position178 := position
				{
					position179, tokenIndex179 := position, tokenIndex
					if !_rules[ruleEscape]() {
						goto l180
					}
					goto l179
				l180:
					position, tokenIndex = position179, tokenIndex179
					{
						position181, tokenIndex181 := position, tokenIndex
						{
							position182, tokenIndex182 := position, tokenIndex
							if buffer[position] != rune(' ') {
								goto l183
							}
							position++
							goto l182
						l183:
							position, tokenIndex = position182, tokenIndex182
							if buffer[position] != rune('\'') {
								goto l184
							}
							position++
							goto l182
						l184:
							position, tokenIndex = position182, tokenIndex182
							if buffer[position] != rune('"') {
								goto l185
							}
							position++
							goto l182
						l185:
							position, tokenIndex = position182, tokenIndex182
							if buffer[position] != rune('\n') {
								goto l186
							}
							position++
							goto l182
						l186:
							position, tokenIndex = position182, tokenIndex182
							if buffer[position] != rune('\\') {
								goto l181
							}
							position++
						}
					l182:
						goto l177
					l181:
						position, tokenIndex = position181, tokenIndex181
					}
					if !matchDot() {
						goto l177
					}
				}
			l179:
				add(ruleStringChar, position178)
			}
			return true
		l177:
			position, tokenIndex = position177, tokenIndex177
			return false
		},
		/* 30 StringLiteral <- <((squote / dquote)? StringChar* (squote / dquote)?)> */
		func() bool {
			{
				position188 := position
				{
					position189, tokenIndex189 := position, tokenIndex
					{
						position191, tokenIndex191 := position, tokenIndex
						if !_rules[rulesquote]() {
							goto l192
						}
						goto l191
					l192:
						position, tokenIndex = position191, tokenIndex191
						if !_rules[ruledquote]() {
							goto l189
						}
					}
				l191:
					goto l190
				l189:
					position, tokenIndex = position189, tokenIndex189
				}
			l190:
			l193:
				{
					position194, tokenIndex194 := position, tokenIndex
					if !_rules[ruleStringChar]() {
						goto l194
					}
					goto l193
				l194:
					position, tokenIndex = position194, tokenIndex194
				}
				{
					position195, tokenIndex195 := position, tokenIndex
					{
						position197, tokenIndex197 := position, tokenIndex
						if !_rules[rulesquote]() {
							goto l198
						}
						goto l197
					l198:
						position, tokenIndex = position197, tokenIndex197
						if !_rules[ruledquote]() {
							goto l195
						}
					}
				l197:
					goto l196
				l195:
					position, tokenIndex = position195, tokenIndex195
				}
			l196:
				add(ruleStringLiteral, position188)
			}
			return true
		},
		/* 31 Escape <- <('\\' ('b' / 't' / 'n' / 'f' / 'r' / '"' / '\'' / '\\' / OctalEscape / UnicodeEscape))> */
		func() bool {
			position199, tokenIndex199 := position, tokenIndex
			{
				position200 := position
				if buffer[position] != rune('\\') {
					goto l199
				}
				position++
				{
					position201, tokenIndex201 := position, tokenIndex
					if buffer[position] != rune('b') {
						goto l202
					}
					position++
					goto l201
				l202:
					position, tokenIndex = position201, tokenIndex201
					if buffer[position] != rune('t') {
						goto l203
					}
					position++
					goto l201
				l203:
					position, tokenIndex = position201, tokenIndex201
					if buffer[position] != rune('n') {
						goto l204
					}
					position++
					goto l201
				l204:
					position, tokenIndex = position201, tokenIndex201
					if buffer[position] != rune('f') {
						goto l205
					}
					position++
					goto l201
				l205:
					position, tokenIndex = position201, tokenIndex201
					if buffer[position] != rune('r') {
						goto l206
					}
					position++
					goto l201
				l206:
					position, tokenIndex = position201, tokenIndex201
					if buffer[position] != rune('"') {
						goto l207
					}
					position++
					goto l201
				l207:
					position, tokenIndex = position201, tokenIndex201
					if buffer[position] != rune('\'') {
						goto l208
					}
					position++
					goto l201
				l208:
					position, tokenIndex = position201, tokenIndex201
					if buffer[position] != rune('\\') {
						goto l209
					}
					position++
					goto l201
				l209:
					position, tokenIndex = position201, tokenIndex201
					if !_rules[ruleOctalEscape]() {
						goto l210
					}
					goto l201
				l210:
					position, tokenIndex = position201, tokenIndex201
					if !_rules[ruleUnicodeEscape]() {
						goto l199
					}
				}
			l201:
				add(ruleEscape, position200)
			}
			return true
		l199:
			position, tokenIndex = position199, tokenIndex199
			return false
		},
		/* 32 OctalEscape <- <(([0-3] [0-7] [0-7]) / ([0-7] [0-7]) / [0-7])> */
		func() bool {
			position211, tokenIndex211 := position, tokenIndex
			{
				position212 := position
				{
					position213, tokenIndex213 := position, tokenIndex
					if c := buffer[position]; c < rune('0') || c > rune('3') {
						goto l214
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l214
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l214
					}
					position++
					goto l213
				l214:
					position, tokenIndex = position213, tokenIndex213
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l215
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l215
					}
					position++
					goto l213
				l215:
					position, tokenIndex = position213, tokenIndex213
					if c := buffer[position]; c < rune('0') || c > rune('7') {
						goto l211
					}
					position++
				}
			l213:
				add(ruleOctalEscape, position212)
			}
			return true
		l211:
			position, tokenIndex = position211, tokenIndex211
			return false
		},
		/* 33 UnicodeEscape <- <('u'+ HexDigit HexDigit HexDigit HexDigit)> */
		func() bool {
			position216, tokenIndex216 := position, tokenIndex
			{
				position217 := position
				if buffer[position] != rune('u') {
					goto l216
				}
				position++
			l218:
				{
					position219, tokenIndex219 := position, tokenIndex
					if buffer[position] != rune('u') {
						goto l219
					}
					position++
					goto l218
				l219:
					position, tokenIndex = position219, tokenIndex219
				}
				if !_rules[ruleHexDigit]() {
					goto l216
				}
				if !_rules[ruleHexDigit]() {
					goto l216
				}
				if !_rules[ruleHexDigit]() {
					goto l216
				}
				if !_rules[ruleHexDigit]() {
					goto l216
				}
				add(ruleUnicodeEscape, position217)
			}
			return true
		l216:
			position, tokenIndex = position216, tokenIndex216
			return false
		},
		/* 34 WordList <- <(Word (space ',' space? WordList)*)> */
		func() bool {
			position220, tokenIndex220 := position, tokenIndex
			{
				position221 := position
				if !_rules[ruleWord]() {
					goto l220
				}
			l222:
				{
					position223, tokenIndex223 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l223
					}
					if buffer[position] != rune(',') {
						goto l223
					}
					position++
					{
						position224, tokenIndex224 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l224
						}
						goto l225
					l224:
						position, tokenIndex = position224, tokenIndex224
					}
				l225:
					if !_rules[ruleWordList]() {
						goto l223
					}
					goto l222
				l223:
					position, tokenIndex = position223, tokenIndex223
				}
				add(ruleWordList, position221)
			}
			return true
		l220:
			position, tokenIndex = position220, tokenIndex220
			return false
		},
		/* 35 Word <- <(<(Letter LetterOrDigit*)> Action0)> */
		func() bool {
			position226, tokenIndex226 := position, tokenIndex
			{
				position227 := position
				{
					position228 := position
					if !_rules[ruleLetter]() {
						goto l226
					}
				l229:
					{
						position230, tokenIndex230 := position, tokenIndex
						if !_rules[ruleLetterOrDigit]() {
							goto l230
						}
						goto l229
					l230:
						position, tokenIndex = position230, tokenIndex230
					}
					add(rulePegText, position228)
				}
				if !_rules[ruleAction0]() {
					goto l226
				}
				add(ruleWord, position227)
			}
			return true
		l226:
			position, tokenIndex = position226, tokenIndex226
			return false
		},
		/* 36 space <- <((' ' / '\t' / '\r' / '\n')+ / ('/' '*' (!('*' '/') .)* ('*' '/')) / ('/' '/' (!('\r' / '\n') .)* ('\r' / '\n')))*> */
		func() bool {
			{
				position232 := position
			l233:
				{
					position234, tokenIndex234 := position, tokenIndex
					{
						position235, tokenIndex235 := position, tokenIndex
						{
							position239, tokenIndex239 := position, tokenIndex
							if buffer[position] != rune(' ') {
								goto l240
							}
							position++
							goto l239
						l240:
							position, tokenIndex = position239, tokenIndex239
							if buffer[position] != rune('\t') {
								goto l241
							}
							position++
							goto l239
						l241:
							position, tokenIndex = position239, tokenIndex239
							if buffer[position] != rune('\r') {
								goto l242
							}
							position++
							goto l239
						l242:
							position, tokenIndex = position239, tokenIndex239
							if buffer[position] != rune('\n') {
								goto l236
							}
							position++
						}
					l239:
					l237:
						{
							position238, tokenIndex238 := position, tokenIndex
							{
								position243, tokenIndex243 := position, tokenIndex
								if buffer[position] != rune(' ') {
									goto l244
								}
								position++
								goto l243
							l244:
								position, tokenIndex = position243, tokenIndex243
								if buffer[position] != rune('\t') {
									goto l245
								}
								position++
								goto l243
							l245:
								position, tokenIndex = position243, tokenIndex243
								if buffer[position] != rune('\r') {
									goto l246
								}
								position++
								goto l243
							l246:
								position, tokenIndex = position243, tokenIndex243
								if buffer[position] != rune('\n') {
									goto l238
								}
								position++
							}
						l243:
							goto l237
						l238:
							position, tokenIndex = position238, tokenIndex238
						}
						goto l235
					l236:
						position, tokenIndex = position235, tokenIndex235
						if buffer[position] != rune('/') {
							goto l247
						}
						position++
						if buffer[position] != rune('*') {
							goto l247
						}
						position++
					l248:
						{
							position249, tokenIndex249 := position, tokenIndex
							{
								position250, tokenIndex250 := position, tokenIndex
								if buffer[position] != rune('*') {
									goto l250
								}
								position++
								if buffer[position] != rune('/') {
									goto l250
								}
								position++
								goto l249
							l250:
								position, tokenIndex = position250, tokenIndex250
							}
							if !matchDot() {
								goto l249
							}
							goto l248
						l249:
							position, tokenIndex = position249, tokenIndex249
						}
						if buffer[position] != rune('*') {
							goto l247
						}
						position++
						if buffer[position] != rune('/') {
							goto l247
						}
						position++
						goto l235
					l247:
						position, tokenIndex = position235, tokenIndex235
						if buffer[position] != rune('/') {
							goto l234
						}
						position++
						if buffer[position] != rune('/') {
							goto l234
						}
						position++
					l251:
						{
							position252, tokenIndex252 := position, tokenIndex
							{
								position253, tokenIndex253 := position, tokenIndex
								{
									position254, tokenIndex254 := position, tokenIndex
									if buffer[position] != rune('\r') {
										goto l255
									}
									position++
									goto l254
								l255:
									position, tokenIndex = position254, tokenIndex254
									if buffer[position] != rune('\n') {
										goto l253
									}
									position++
								}
							l254:
								goto l252
							l253:
								position, tokenIndex = position253, tokenIndex253
							}
							if !matchDot() {
								goto l252
							}
							goto l251
						l252:
							position, tokenIndex = position252, tokenIndex252
						}
						{
							position256, tokenIndex256 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l257
							}
							position++
							goto l256
						l257:
							position, tokenIndex = position256, tokenIndex256
							if buffer[position] != rune('\n') {
								goto l234
							}
							position++
						}
					l256:
					}
				l235:
					goto l233
				l234:
					position, tokenIndex = position234, tokenIndex234
				}
				add(rulespace, position232)
			}
			return true
		},
		nil,
		/* 39 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
	}
	p.rules = _rules
}
