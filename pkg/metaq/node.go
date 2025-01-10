package metaq

import (
	"fmt"
	"go/ast"
	"strings"
)

// BuildAsqNode converts an ast.Node to its corresponding metaq.Node
func BuildAsqNode(node ast.Node, wildcardIdent map[*ast.Ident]bool) Node {
	if node == nil {
		return nil
	}

	switch astObj := node.(type) {
	case *ast.CallExpr:
		return &CallExpr{
			Ast: astObj,
			Fun: BuildAsqExpr(astObj.Fun, wildcardIdent),
		}
	case *ast.SelectorExpr:
		return &SelectorExpr{
			Ast: astObj,
			X:   BuildAsqExpr(astObj.X, wildcardIdent),
		}
	case *ast.Ident:
		_, isWildcard := wildcardIdent[astObj]
		return &Ident{
			Ast:      astObj,
			Wildcard: isWildcard,
		}
	default:
		return &DefaultNode{Node: astObj}
	}
}

// BuildAsqExpr converts an ast.Node to its corresponding metaq.Node
func BuildAsqExpr(node ast.Node, wildcardIdent map[*ast.Ident]bool) Expr {
	if node == nil {
		return nil
	}

	switch astObj := node.(type) {
	case *ast.CallExpr:
		return &CallExpr{
			Ast: astObj,
			Fun: BuildAsqExpr(astObj.Fun, wildcardIdent),
		}
	case *ast.SelectorExpr:
		return &SelectorExpr{
			Ast: astObj,
			X:   BuildAsqExpr(astObj.X, wildcardIdent),
		}
	default:
		return &DefaultExpr{Node: astObj}
	}
}

// CallExpr wraps an ast.CallExpr node
type CallExpr struct {
	Ast *ast.CallExpr
	Fun Expr
}

func (c *CallExpr) exprNode() {}

func (c *CallExpr) Convert() string {
	var sb strings.Builder
	sb.WriteString("(call_expression function: ")
	sb.WriteString(c.Fun.Convert())
	sb.WriteString(" arguments: (argument_list))")
	return sb.String()
}

func (c *CallExpr) Pos() Pos {
	return Pos(c.Ast.Pos())
}

// SelectorExpr wraps an ast.SelectorExpr node
type SelectorExpr struct {
	Ast *ast.SelectorExpr
	X   Expr
}

func (s *SelectorExpr) exprNode() {}

func (s *SelectorExpr) Convert() string {
	var sb strings.Builder
	sb.WriteString("(selector_expression operand: ")
	sb.WriteString(s.X.Convert())
	sb.WriteString(fmt.Sprintf(` field: (field_identifier) @field (#eq? @field "%s"))`, s.Ast.Sel.Name))
	return sb.String()
}

func (s *SelectorExpr) Pos() Pos {
	return Pos(s.Ast.Pos())
}

// Ident wraps an ast.Ident node with an additional Wildcard field
type Ident struct {
	Ast      *ast.Ident
	Wildcard bool
}

func (i *Ident) Convert() string {
	if i.Wildcard {
		return "(identifier)"
	}
	return fmt.Sprintf(`(identifier) @name (#eq? @name "%s")`, i.Ast.Name)
}

func (i *Ident) Pos() Pos {
	return Pos(i.Ast.Pos())
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

func (d *DefaultNode) Pos() Pos {
	return Pos(0)
}

// DefaultExpr wraps any ast.Expr type that doesn't have a specific implementation
type DefaultExpr struct {
	Node ast.Node
}

func (d *DefaultExpr) exprNode() {}

func (d *DefaultExpr) Convert() string {
	return fmt.Sprintf("(%T)", d.Node)
}

func (d *DefaultExpr) Pos() Pos {
	return Pos(0)
}
