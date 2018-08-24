package grammar

import "fmt"

type Type uint8

// Query structure
type Query struct {
	Select
	Order
	Limit
	ExprStack
}

// Select structure
type Select struct {
	Table  string
	Fields []string
}

const (
	ASC  = 0
	DESC = 1
)

type Order struct {
	Field    string
	Ordering int
}

type Limit struct {
	ShouldLimit bool
	Count       int
}

// Stack is a basic LIFO stack that resizes as needed.
type ExprStack struct {
	nodes []NodeExpr
	count int
}

// Push adds a node to the stack.
func (s *ExprStack) Push(n NodeExpr) {
	s.nodes = append(s.nodes[:s.count], n)
	s.count++
}

// Pop removes and returns a node from the stack in last to first order.
func (s *ExprStack) Pop() NodeExpr {
	if s.count == 0 {
		return nil
	}
	s.count--
	return s.nodes[s.count]
}

func (s *ExprStack) Len() int {
	return s.count
}

func (s *ExprStack) String() string {
	var str string
	for _, expr := range s.nodes {
		str = fmt.Sprintf("%s %s", str, expr)
	}
	return str
}

type NodeExpr interface {
	String() string
}

type NodeEquals struct {
	left  string
	right string
}

func (n *NodeEquals) String() string {
	return fmt.Sprintf("%s == %s", n.left, n.right)
}

type NodeAnd struct {
	left  NodeExpr
	right NodeExpr
}

func (n *NodeAnd) String() string {
	return fmt.Sprintf("%s AND %s", n.left.String(), n.right.String())
}

type NodeOr struct {
	left  NodeExpr
	right NodeExpr
}

func (n *NodeOr) String() string {
	return fmt.Sprintf("%s OR %s", n.left.String(), n.right.String())
}

type NodeGreaterThan struct {
	left  string
	right int
}

func (n *NodeGreaterThan) String() string {
	return fmt.Sprintf("%s > %s", n.left, n.right)
}

type NodeGreaterThanOrEqual struct {
	left  string
	right int
}

func (n *NodeGreaterThanOrEqual) String() string {
	return fmt.Sprintf("%s >= %s", n.left, n.right)
}

type NodeLessThan struct {
	left  string
	right int
}

func (n *NodeLessThan) String() string {
	return fmt.Sprintf("%s < %d", n.left, n.right)
}

type NodeLessThanOrEqual struct {
	left  string
	right int
}

func (n *NodeLessThanOrEqual) String() string {
	return fmt.Sprintf("%s <= %d", n.left, n.right)
}

type NodeNotEqual struct {
	left  string
	right string
}

func (n *NodeNotEqual) String() string {
	return fmt.Sprintf("%s != %s", n.left, n.right)
}

// type NodeStrEquals struct {
// 	NodeExpr
// 	left  NodeExpr
// 	right NodeExpr
// }

// func (n *NodeStrEquals) Assertion() bool {
// 	return n.left == n.right
// }

// type NodeAnd struct {
// 	NodeExpr
// 	left  NodeExpr
// 	right NodeExpr
// }

// func (n *NodeAnd) Assertion() bool {
// 	return n.left.Assertion() && n.right.Assertion()
// }

// type NodeOr struct {
// 	NodeExpr
// 	left  NodeExpr
// 	right NodeExpr
// }

// func (n *NodeOr) Assertion() bool {
// 	return n.left.Assertion() || n.right.Assertion()
// }
