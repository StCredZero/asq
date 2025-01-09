package metaq

import (
	"fmt"
	"go/ast"
	"strings"
)

// Node is the interface that all metaq nodes implement.
type Node interface {
	Convert() string
}

// Inspect traverses a metaq Node in depth-first order.
// This parallels ast.Inspect(node ast.Node, fn func(ast.Node) bool).
func Inspect(n Node, fn func(Node) bool) {
	if n == nil {
		return
	}
	if !fn(n) {
		return
	}
	// TODO: We'll handle children once we define them in each Node struct
	// stub
}

// BuildMetaqNode converts an ast.Node to its corresponding metaq.Node
func BuildMetaqNode(node ast.Node) Node {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.CallExpr:
		return &CallExpr{
			Call: n,
			Fun:  BuildMetaqNode(n.Fun),
		}
	case *ast.SelectorExpr:
		return &SelectorExpr{
			Sel: n,
			X:   BuildMetaqNode(n.X),
		}
	case *ast.Ident:
		return &Ident{
			Id:       n,
			Wildcard: strings.HasPrefix(n.Name, "wildcarded_"),
		}
	default:
		return &DefaultNode{Node: n}
	}
}

// CallExpr wraps an ast.CallExpr node
type CallExpr struct {
	Call *ast.CallExpr
	Fun  Node
}

func (c *CallExpr) Convert() string {
	var sb strings.Builder
	sb.WriteString("(call_expression function: ")
	sb.WriteString(c.Fun.Convert())
	sb.WriteString(" arguments: (argument_list))")
	return sb.String()
}

// SelectorExpr wraps an ast.SelectorExpr node
type SelectorExpr struct {
	Sel *ast.SelectorExpr
	X   Node
}

func (s *SelectorExpr) Convert() string {
	var sb strings.Builder
	sb.WriteString("(selector_expression operand: ")
	sb.WriteString(s.X.Convert())
	sb.WriteString(fmt.Sprintf(` field: (field_identifier) @field (#eq? @field "%s"))`, s.Sel.Sel.Name))
	return sb.String()
}

// Ident wraps an ast.Ident node with an additional Wildcard field
type Ident struct {
	Id       *ast.Ident
	Wildcard bool
}

func (i *Ident) Convert() string {
	if i.Wildcard {
		return "(identifier)"
	}
	return fmt.Sprintf(`(identifier) @name (#eq? @name "%s")`, i.Id.Name)
}

// The following ast.Node types have not been fully implemented yet:
/*
- *ast.ArrayType
- *ast.AssignStmt
- *ast.BadDecl
- *ast.BadExpr
- *ast.BadStmt
- *ast.BasicLit
- *ast.BinaryExpr
- *ast.BlockStmt
- *ast.BranchStmt
- *ast.CaseClause
- *ast.ChanType
- *ast.CommClause
- *ast.CompositeLit
- *ast.DeclStmt
- *ast.DeferStmt
- *ast.Ellipsis
- *ast.EmptyStmt
- *ast.ExprStmt
- *ast.Field
- *ast.FieldList
- *ast.File
- *ast.ForStmt
- *ast.FuncDecl
- *ast.FuncLit
- *ast.FuncType
- *ast.GenDecl
- *ast.GoStmt
- *ast.IfStmt
- *ast.ImportSpec
- *ast.IncDecStmt
- *ast.IndexExpr
- *ast.InterfaceType
- *ast.KeyValueExpr
- *ast.LabeledStmt
- *ast.MapType
- *ast.Package
- *ast.ParenExpr
- *ast.RangeStmt
- *ast.ReturnStmt
- *ast.SelectStmt
- *ast.SendStmt
- *ast.SliceExpr
- *ast.StarExpr
- *ast.StructType
- *ast.SwitchStmt
- *ast.TypeAssertExpr
- *ast.TypeSpec
- *ast.TypeSwitchStmt
- *ast.UnaryExpr
- *ast.ValueSpec
*/

// DefaultNode wraps any ast.Node type that doesn't have a specific implementation
type DefaultNode struct {
	Node ast.Node
}

func (d *DefaultNode) Convert() string {
	return fmt.Sprintf("(%T)", d.Node)
}
