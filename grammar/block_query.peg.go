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
	ruleequals
	rulesquote
	ruledquote
	ruleNumber
	ruleLetter
	ruleLetterOrDigit
	ruleWordList
	ruleWords
	ruleWord
	rulespace
	ruleAction0
	rulePegText
	ruleAction1
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
	"equals",
	"squote",
	"dquote",
	"Number",
	"Letter",
	"LetterOrDigit",
	"WordList",
	"Words",
	"Word",
	"space",
	"Action0",
	"PegText",
	"Action1",
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
	rules  [34]func() bool
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

		case ruleAction1:

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
		/* 6 limitStmt <- <(limit Number)> */
		func() bool {
			position22, tokenIndex22 := position, tokenIndex
			{
				position23 := position
				if !_rules[rulelimit]() {
					goto l22
				}
				if !_rules[ruleNumber]() {
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
		/* 10 whereExpr <- <(Word equals Word)> */
		func() bool {
			position34, tokenIndex34 := position, tokenIndex
			{
				position35 := position
				if !_rules[ruleWord]() {
					goto l34
				}
				if !_rules[ruleequals]() {
					goto l34
				}
				if !_rules[ruleWord]() {
					goto l34
				}
				add(rulewhereExpr, position35)
			}
			return true
		l34:
			position, tokenIndex = position34, tokenIndex34
			return false
		},
		/* 11 select <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') space)> */
		func() bool {
			position36, tokenIndex36 := position, tokenIndex
			{
				position37 := position
				{
					position38, tokenIndex38 := position, tokenIndex
					if buffer[position] != rune('s') {
						goto l39
					}
					position++
					goto l38
				l39:
					position, tokenIndex = position38, tokenIndex38
					if buffer[position] != rune('S') {
						goto l36
					}
					position++
				}
			l38:
				{
					position40, tokenIndex40 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l41
					}
					position++
					goto l40
				l41:
					position, tokenIndex = position40, tokenIndex40
					if buffer[position] != rune('E') {
						goto l36
					}
					position++
				}
			l40:
				{
					position42, tokenIndex42 := position, tokenIndex
					if buffer[position] != rune('l') {
						goto l43
					}
					position++
					goto l42
				l43:
					position, tokenIndex = position42, tokenIndex42
					if buffer[position] != rune('L') {
						goto l36
					}
					position++
				}
			l42:
				{
					position44, tokenIndex44 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l45
					}
					position++
					goto l44
				l45:
					position, tokenIndex = position44, tokenIndex44
					if buffer[position] != rune('E') {
						goto l36
					}
					position++
				}
			l44:
				{
					position46, tokenIndex46 := position, tokenIndex
					if buffer[position] != rune('c') {
						goto l47
					}
					position++
					goto l46
				l47:
					position, tokenIndex = position46, tokenIndex46
					if buffer[position] != rune('C') {
						goto l36
					}
					position++
				}
			l46:
				{
					position48, tokenIndex48 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l49
					}
					position++
					goto l48
				l49:
					position, tokenIndex = position48, tokenIndex48
					if buffer[position] != rune('T') {
						goto l36
					}
					position++
				}
			l48:
				if !_rules[rulespace]() {
					goto l36
				}
				add(ruleselect, position37)
			}
			return true
		l36:
			position, tokenIndex = position36, tokenIndex36
			return false
		},
		/* 12 from <- <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M') space)> */
		func() bool {
			position50, tokenIndex50 := position, tokenIndex
			{
				position51 := position
				{
					position52, tokenIndex52 := position, tokenIndex
					if buffer[position] != rune('f') {
						goto l53
					}
					position++
					goto l52
				l53:
					position, tokenIndex = position52, tokenIndex52
					if buffer[position] != rune('F') {
						goto l50
					}
					position++
				}
			l52:
				{
					position54, tokenIndex54 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l55
					}
					position++
					goto l54
				l55:
					position, tokenIndex = position54, tokenIndex54
					if buffer[position] != rune('R') {
						goto l50
					}
					position++
				}
			l54:
				{
					position56, tokenIndex56 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l57
					}
					position++
					goto l56
				l57:
					position, tokenIndex = position56, tokenIndex56
					if buffer[position] != rune('O') {
						goto l50
					}
					position++
				}
			l56:
				{
					position58, tokenIndex58 := position, tokenIndex
					if buffer[position] != rune('m') {
						goto l59
					}
					position++
					goto l58
				l59:
					position, tokenIndex = position58, tokenIndex58
					if buffer[position] != rune('M') {
						goto l50
					}
					position++
				}
			l58:
				if !_rules[rulespace]() {
					goto l50
				}
				add(rulefrom, position51)
			}
			return true
		l50:
			position, tokenIndex = position50, tokenIndex50
			return false
		},
		/* 13 order <- <(('o' / 'O') ('r' / 'R') ('d' / 'D') ('e' / 'E') ('r' / 'R') space)> */
		func() bool {
			position60, tokenIndex60 := position, tokenIndex
			{
				position61 := position
				{
					position62, tokenIndex62 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l63
					}
					position++
					goto l62
				l63:
					position, tokenIndex = position62, tokenIndex62
					if buffer[position] != rune('O') {
						goto l60
					}
					position++
				}
			l62:
				{
					position64, tokenIndex64 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l65
					}
					position++
					goto l64
				l65:
					position, tokenIndex = position64, tokenIndex64
					if buffer[position] != rune('R') {
						goto l60
					}
					position++
				}
			l64:
				{
					position66, tokenIndex66 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l67
					}
					position++
					goto l66
				l67:
					position, tokenIndex = position66, tokenIndex66
					if buffer[position] != rune('D') {
						goto l60
					}
					position++
				}
			l66:
				{
					position68, tokenIndex68 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l69
					}
					position++
					goto l68
				l69:
					position, tokenIndex = position68, tokenIndex68
					if buffer[position] != rune('E') {
						goto l60
					}
					position++
				}
			l68:
				{
					position70, tokenIndex70 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l71
					}
					position++
					goto l70
				l71:
					position, tokenIndex = position70, tokenIndex70
					if buffer[position] != rune('R') {
						goto l60
					}
					position++
				}
			l70:
				if !_rules[rulespace]() {
					goto l60
				}
				add(ruleorder, position61)
			}
			return true
		l60:
			position, tokenIndex = position60, tokenIndex60
			return false
		},
		/* 14 by <- <(('b' / 'B') ('y' / 'Y') space)> */
		func() bool {
			position72, tokenIndex72 := position, tokenIndex
			{
				position73 := position
				{
					position74, tokenIndex74 := position, tokenIndex
					if buffer[position] != rune('b') {
						goto l75
					}
					position++
					goto l74
				l75:
					position, tokenIndex = position74, tokenIndex74
					if buffer[position] != rune('B') {
						goto l72
					}
					position++
				}
			l74:
				{
					position76, tokenIndex76 := position, tokenIndex
					if buffer[position] != rune('y') {
						goto l77
					}
					position++
					goto l76
				l77:
					position, tokenIndex = position76, tokenIndex76
					if buffer[position] != rune('Y') {
						goto l72
					}
					position++
				}
			l76:
				if !_rules[rulespace]() {
					goto l72
				}
				add(ruleby, position73)
			}
			return true
		l72:
			position, tokenIndex = position72, tokenIndex72
			return false
		},
		/* 15 star <- <(('*' / (('a' / 'A') ('l' / 'L') ('l' / 'L'))) space)> */
		func() bool {
			position78, tokenIndex78 := position, tokenIndex
			{
				position79 := position
				{
					position80, tokenIndex80 := position, tokenIndex
					if buffer[position] != rune('*') {
						goto l81
					}
					position++
					goto l80
				l81:
					position, tokenIndex = position80, tokenIndex80
					{
						position82, tokenIndex82 := position, tokenIndex
						if buffer[position] != rune('a') {
							goto l83
						}
						position++
						goto l82
					l83:
						position, tokenIndex = position82, tokenIndex82
						if buffer[position] != rune('A') {
							goto l78
						}
						position++
					}
				l82:
					{
						position84, tokenIndex84 := position, tokenIndex
						if buffer[position] != rune('l') {
							goto l85
						}
						position++
						goto l84
					l85:
						position, tokenIndex = position84, tokenIndex84
						if buffer[position] != rune('L') {
							goto l78
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
							goto l78
						}
						position++
					}
				l86:
				}
			l80:
				if !_rules[rulespace]() {
					goto l78
				}
				add(rulestar, position79)
			}
			return true
		l78:
			position, tokenIndex = position78, tokenIndex78
			return false
		},
		/* 16 limit <- <(('l' / 'L') ('i' / 'I') ('m' / 'M') ('i' / 'I') ('t' / 'T') space)> */
		func() bool {
			position88, tokenIndex88 := position, tokenIndex
			{
				position89 := position
				{
					position90, tokenIndex90 := position, tokenIndex
					if buffer[position] != rune('l') {
						goto l91
					}
					position++
					goto l90
				l91:
					position, tokenIndex = position90, tokenIndex90
					if buffer[position] != rune('L') {
						goto l88
					}
					position++
				}
			l90:
				{
					position92, tokenIndex92 := position, tokenIndex
					if buffer[position] != rune('i') {
						goto l93
					}
					position++
					goto l92
				l93:
					position, tokenIndex = position92, tokenIndex92
					if buffer[position] != rune('I') {
						goto l88
					}
					position++
				}
			l92:
				{
					position94, tokenIndex94 := position, tokenIndex
					if buffer[position] != rune('m') {
						goto l95
					}
					position++
					goto l94
				l95:
					position, tokenIndex = position94, tokenIndex94
					if buffer[position] != rune('M') {
						goto l88
					}
					position++
				}
			l94:
				{
					position96, tokenIndex96 := position, tokenIndex
					if buffer[position] != rune('i') {
						goto l97
					}
					position++
					goto l96
				l97:
					position, tokenIndex = position96, tokenIndex96
					if buffer[position] != rune('I') {
						goto l88
					}
					position++
				}
			l96:
				{
					position98, tokenIndex98 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l99
					}
					position++
					goto l98
				l99:
					position, tokenIndex = position98, tokenIndex98
					if buffer[position] != rune('T') {
						goto l88
					}
					position++
				}
			l98:
				if !_rules[rulespace]() {
					goto l88
				}
				add(rulelimit, position89)
			}
			return true
		l88:
			position, tokenIndex = position88, tokenIndex88
			return false
		},
		/* 17 where <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') space)> */
		func() bool {
			position100, tokenIndex100 := position, tokenIndex
			{
				position101 := position
				{
					position102, tokenIndex102 := position, tokenIndex
					if buffer[position] != rune('w') {
						goto l103
					}
					position++
					goto l102
				l103:
					position, tokenIndex = position102, tokenIndex102
					if buffer[position] != rune('W') {
						goto l100
					}
					position++
				}
			l102:
				{
					position104, tokenIndex104 := position, tokenIndex
					if buffer[position] != rune('h') {
						goto l105
					}
					position++
					goto l104
				l105:
					position, tokenIndex = position104, tokenIndex104
					if buffer[position] != rune('H') {
						goto l100
					}
					position++
				}
			l104:
				{
					position106, tokenIndex106 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l107
					}
					position++
					goto l106
				l107:
					position, tokenIndex = position106, tokenIndex106
					if buffer[position] != rune('E') {
						goto l100
					}
					position++
				}
			l106:
				{
					position108, tokenIndex108 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l109
					}
					position++
					goto l108
				l109:
					position, tokenIndex = position108, tokenIndex108
					if buffer[position] != rune('R') {
						goto l100
					}
					position++
				}
			l108:
				{
					position110, tokenIndex110 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l111
					}
					position++
					goto l110
				l111:
					position, tokenIndex = position110, tokenIndex110
					if buffer[position] != rune('E') {
						goto l100
					}
					position++
				}
			l110:
				if !_rules[rulespace]() {
					goto l100
				}
				add(rulewhere, position101)
			}
			return true
		l100:
			position, tokenIndex = position100, tokenIndex100
			return false
		},
		/* 18 and <- <(('a' / 'A') ('n' / 'N') ('d' / 'D') space)> */
		func() bool {
			position112, tokenIndex112 := position, tokenIndex
			{
				position113 := position
				{
					position114, tokenIndex114 := position, tokenIndex
					if buffer[position] != rune('a') {
						goto l115
					}
					position++
					goto l114
				l115:
					position, tokenIndex = position114, tokenIndex114
					if buffer[position] != rune('A') {
						goto l112
					}
					position++
				}
			l114:
				{
					position116, tokenIndex116 := position, tokenIndex
					if buffer[position] != rune('n') {
						goto l117
					}
					position++
					goto l116
				l117:
					position, tokenIndex = position116, tokenIndex116
					if buffer[position] != rune('N') {
						goto l112
					}
					position++
				}
			l116:
				{
					position118, tokenIndex118 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l119
					}
					position++
					goto l118
				l119:
					position, tokenIndex = position118, tokenIndex118
					if buffer[position] != rune('D') {
						goto l112
					}
					position++
				}
			l118:
				if !_rules[rulespace]() {
					goto l112
				}
				add(ruleand, position113)
			}
			return true
		l112:
			position, tokenIndex = position112, tokenIndex112
			return false
		},
		/* 19 or <- <(('o' / 'O') ('r' / 'R') space)> */
		func() bool {
			position120, tokenIndex120 := position, tokenIndex
			{
				position121 := position
				{
					position122, tokenIndex122 := position, tokenIndex
					if buffer[position] != rune('o') {
						goto l123
					}
					position++
					goto l122
				l123:
					position, tokenIndex = position122, tokenIndex122
					if buffer[position] != rune('O') {
						goto l120
					}
					position++
				}
			l122:
				{
					position124, tokenIndex124 := position, tokenIndex
					if buffer[position] != rune('r') {
						goto l125
					}
					position++
					goto l124
				l125:
					position, tokenIndex = position124, tokenIndex124
					if buffer[position] != rune('R') {
						goto l120
					}
					position++
				}
			l124:
				if !_rules[rulespace]() {
					goto l120
				}
				add(ruleor, position121)
			}
			return true
		l120:
			position, tokenIndex = position120, tokenIndex120
			return false
		},
		/* 20 equals <- <('=' space)> */
		func() bool {
			position126, tokenIndex126 := position, tokenIndex
			{
				position127 := position
				if buffer[position] != rune('=') {
					goto l126
				}
				position++
				if !_rules[rulespace]() {
					goto l126
				}
				add(ruleequals, position127)
			}
			return true
		l126:
			position, tokenIndex = position126, tokenIndex126
			return false
		},
		/* 21 squote <- <'\''> */
		func() bool {
			position128, tokenIndex128 := position, tokenIndex
			{
				position129 := position
				if buffer[position] != rune('\'') {
					goto l128
				}
				position++
				add(rulesquote, position129)
			}
			return true
		l128:
			position, tokenIndex = position128, tokenIndex128
			return false
		},
		/* 22 dquote <- <'"'> */
		func() bool {
			position130, tokenIndex130 := position, tokenIndex
			{
				position131 := position
				if buffer[position] != rune('"') {
					goto l130
				}
				position++
				add(ruledquote, position131)
			}
			return true
		l130:
			position, tokenIndex = position130, tokenIndex130
			return false
		},
		/* 23 Number <- <([0-9]* Action0)> */
		func() bool {
			position132, tokenIndex132 := position, tokenIndex
			{
				position133 := position
			l134:
				{
					position135, tokenIndex135 := position, tokenIndex
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l135
					}
					position++
					goto l134
				l135:
					position, tokenIndex = position135, tokenIndex135
				}
				if !_rules[ruleAction0]() {
					goto l132
				}
				add(ruleNumber, position133)
			}
			return true
		l132:
			position, tokenIndex = position132, tokenIndex132
			return false
		},
		/* 24 Letter <- <([a-z] / [A-Z] / ('_' / '$'))> */
		func() bool {
			position136, tokenIndex136 := position, tokenIndex
			{
				position137 := position
				{
					position138, tokenIndex138 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l139
					}
					position++
					goto l138
				l139:
					position, tokenIndex = position138, tokenIndex138
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l140
					}
					position++
					goto l138
				l140:
					position, tokenIndex = position138, tokenIndex138
					{
						position141, tokenIndex141 := position, tokenIndex
						if buffer[position] != rune('_') {
							goto l142
						}
						position++
						goto l141
					l142:
						position, tokenIndex = position141, tokenIndex141
						if buffer[position] != rune('$') {
							goto l136
						}
						position++
					}
				l141:
				}
			l138:
				add(ruleLetter, position137)
			}
			return true
		l136:
			position, tokenIndex = position136, tokenIndex136
			return false
		},
		/* 25 LetterOrDigit <- <([a-z] / [A-Z] / [0-9] / ('_' / '$'))> */
		func() bool {
			position143, tokenIndex143 := position, tokenIndex
			{
				position144 := position
				{
					position145, tokenIndex145 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l146
					}
					position++
					goto l145
				l146:
					position, tokenIndex = position145, tokenIndex145
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l147
					}
					position++
					goto l145
				l147:
					position, tokenIndex = position145, tokenIndex145
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l148
					}
					position++
					goto l145
				l148:
					position, tokenIndex = position145, tokenIndex145
					{
						position149, tokenIndex149 := position, tokenIndex
						if buffer[position] != rune('_') {
							goto l150
						}
						position++
						goto l149
					l150:
						position, tokenIndex = position149, tokenIndex149
						if buffer[position] != rune('$') {
							goto l143
						}
						position++
					}
				l149:
				}
			l145:
				add(ruleLetterOrDigit, position144)
			}
			return true
		l143:
			position, tokenIndex = position143, tokenIndex143
			return false
		},
		/* 26 WordList <- <(Word (space ',' space? WordList)*)> */
		func() bool {
			position151, tokenIndex151 := position, tokenIndex
			{
				position152 := position
				if !_rules[ruleWord]() {
					goto l151
				}
			l153:
				{
					position154, tokenIndex154 := position, tokenIndex
					if !_rules[rulespace]() {
						goto l154
					}
					if buffer[position] != rune(',') {
						goto l154
					}
					position++
					{
						position155, tokenIndex155 := position, tokenIndex
						if !_rules[rulespace]() {
							goto l155
						}
						goto l156
					l155:
						position, tokenIndex = position155, tokenIndex155
					}
				l156:
					if !_rules[ruleWordList]() {
						goto l154
					}
					goto l153
				l154:
					position, tokenIndex = position154, tokenIndex154
				}
				add(ruleWordList, position152)
			}
			return true
		l151:
			position, tokenIndex = position151, tokenIndex151
			return false
		},
		/* 27 Words <- <(Word (space Word)*)> */
		nil,
		/* 28 Word <- <((squote / dquote)? <(Letter LetterOrDigit*)> (squote / dquote)? Action1)> */
		func() bool {
			position158, tokenIndex158 := position, tokenIndex
			{
				position159 := position
				{
					position160, tokenIndex160 := position, tokenIndex
					{
						position162, tokenIndex162 := position, tokenIndex
						if !_rules[rulesquote]() {
							goto l163
						}
						goto l162
					l163:
						position, tokenIndex = position162, tokenIndex162
						if !_rules[ruledquote]() {
							goto l160
						}
					}
				l162:
					goto l161
				l160:
					position, tokenIndex = position160, tokenIndex160
				}
			l161:
				{
					position164 := position
					if !_rules[ruleLetter]() {
						goto l158
					}
				l165:
					{
						position166, tokenIndex166 := position, tokenIndex
						if !_rules[ruleLetterOrDigit]() {
							goto l166
						}
						goto l165
					l166:
						position, tokenIndex = position166, tokenIndex166
					}
					add(rulePegText, position164)
				}
				{
					position167, tokenIndex167 := position, tokenIndex
					{
						position169, tokenIndex169 := position, tokenIndex
						if !_rules[rulesquote]() {
							goto l170
						}
						goto l169
					l170:
						position, tokenIndex = position169, tokenIndex169
						if !_rules[ruledquote]() {
							goto l167
						}
					}
				l169:
					goto l168
				l167:
					position, tokenIndex = position167, tokenIndex167
				}
			l168:
				if !_rules[ruleAction1]() {
					goto l158
				}
				add(ruleWord, position159)
			}
			return true
		l158:
			position, tokenIndex = position158, tokenIndex158
			return false
		},
		/* 29 space <- <((' ' / '\t' / '\r' / '\n')+ / ('/' '*' (!('*' '/') .)* ('*' '/')) / ('/' '/' (!('\r' / '\n') .)* ('\r' / '\n')))*> */
		func() bool {
			{
				position172 := position
			l173:
				{
					position174, tokenIndex174 := position, tokenIndex
					{
						position175, tokenIndex175 := position, tokenIndex
						{
							position179, tokenIndex179 := position, tokenIndex
							if buffer[position] != rune(' ') {
								goto l180
							}
							position++
							goto l179
						l180:
							position, tokenIndex = position179, tokenIndex179
							if buffer[position] != rune('\t') {
								goto l181
							}
							position++
							goto l179
						l181:
							position, tokenIndex = position179, tokenIndex179
							if buffer[position] != rune('\r') {
								goto l182
							}
							position++
							goto l179
						l182:
							position, tokenIndex = position179, tokenIndex179
							if buffer[position] != rune('\n') {
								goto l176
							}
							position++
						}
					l179:
					l177:
						{
							position178, tokenIndex178 := position, tokenIndex
							{
								position183, tokenIndex183 := position, tokenIndex
								if buffer[position] != rune(' ') {
									goto l184
								}
								position++
								goto l183
							l184:
								position, tokenIndex = position183, tokenIndex183
								if buffer[position] != rune('\t') {
									goto l185
								}
								position++
								goto l183
							l185:
								position, tokenIndex = position183, tokenIndex183
								if buffer[position] != rune('\r') {
									goto l186
								}
								position++
								goto l183
							l186:
								position, tokenIndex = position183, tokenIndex183
								if buffer[position] != rune('\n') {
									goto l178
								}
								position++
							}
						l183:
							goto l177
						l178:
							position, tokenIndex = position178, tokenIndex178
						}
						goto l175
					l176:
						position, tokenIndex = position175, tokenIndex175
						if buffer[position] != rune('/') {
							goto l187
						}
						position++
						if buffer[position] != rune('*') {
							goto l187
						}
						position++
					l188:
						{
							position189, tokenIndex189 := position, tokenIndex
							{
								position190, tokenIndex190 := position, tokenIndex
								if buffer[position] != rune('*') {
									goto l190
								}
								position++
								if buffer[position] != rune('/') {
									goto l190
								}
								position++
								goto l189
							l190:
								position, tokenIndex = position190, tokenIndex190
							}
							if !matchDot() {
								goto l189
							}
							goto l188
						l189:
							position, tokenIndex = position189, tokenIndex189
						}
						if buffer[position] != rune('*') {
							goto l187
						}
						position++
						if buffer[position] != rune('/') {
							goto l187
						}
						position++
						goto l175
					l187:
						position, tokenIndex = position175, tokenIndex175
						if buffer[position] != rune('/') {
							goto l174
						}
						position++
						if buffer[position] != rune('/') {
							goto l174
						}
						position++
					l191:
						{
							position192, tokenIndex192 := position, tokenIndex
							{
								position193, tokenIndex193 := position, tokenIndex
								{
									position194, tokenIndex194 := position, tokenIndex
									if buffer[position] != rune('\r') {
										goto l195
									}
									position++
									goto l194
								l195:
									position, tokenIndex = position194, tokenIndex194
									if buffer[position] != rune('\n') {
										goto l193
									}
									position++
								}
							l194:
								goto l192
							l193:
								position, tokenIndex = position193, tokenIndex193
							}
							if !matchDot() {
								goto l192
							}
							goto l191
						l192:
							position, tokenIndex = position192, tokenIndex192
						}
						{
							position196, tokenIndex196 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l197
							}
							position++
							goto l196
						l197:
							position, tokenIndex = position196, tokenIndex196
							if buffer[position] != rune('\n') {
								goto l174
							}
							position++
						}
					l196:
					}
				l175:
					goto l173
				l174:
					position, tokenIndex = position174, tokenIndex174
				}
				add(rulespace, position172)
			}
			return true
		},
		/* 31 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		nil,
		/* 33 Action1 <- <{}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
	}
	p.rules = _rules
}
