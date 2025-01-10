package metaq

import (
	"go/ast"
	"io"
)

type Pos int

// Node is the interface that all metaq nodes implement.
type Node interface {
	WriteTreeSitterQuery(w io.Writer) error
	AstNode() ast.Node
}

// Expr nodes contain expressions and implement the Expr interface.
type Expr interface {
	Node
	exprNode()
}

// Stmt nodes contain statements and implement the Stmt interface.
type Stmt interface {
	Node
	stmtNode()
}

// Decl nodes contain declarations and implement the Decl interface.
type Decl interface {
	Node
	declNode()
}
