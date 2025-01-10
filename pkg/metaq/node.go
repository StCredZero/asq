package metaq

import (
	"fmt"
	"github.com/StCredZero/asq/pkg/slicex"
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
	case ast.Stmt:
		return BuildAsqStmt(astObj, wildcardIdent)
	case ast.Decl:
		return BuildAsqDecl(astObj, wildcardIdent)
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
			Args: slicex.Map(astObj.Args, func(arg ast.Expr) Expr {
				return BuildAsqExpr(arg, wildcardIdent)
			}),
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
	Ast      *ast.CallExpr
	Fun      Expr
	Args     []Expr
	Wildcard bool
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
	Ast      *ast.SelectorExpr
	X        Expr
	Wildcard bool
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

// DefaultStmt wraps any ast.Stmt type that doesn't have a specific implementation
type DefaultStmt struct {
	Ast ast.Stmt
}

func (d *DefaultStmt) stmtNode() {}

func (d *DefaultStmt) Convert() string {
	return fmt.Sprintf("(%T)", d.Ast)
}

func (d *DefaultStmt) Pos() Pos {
	return Pos(d.Ast.Pos())
}

// DefaultDecl wraps any ast.Decl type that doesn't have a specific implementation
type DefaultDecl struct {
	Ast ast.Decl
}

func (d *DefaultDecl) declNode() {}

func (d *DefaultDecl) Convert() string {
	return fmt.Sprintf("(%T)", d.Ast)
}

func (d *DefaultDecl) Pos() Pos {
	return Pos(d.Ast.Pos())
}

// ReturnStmt wraps an ast.ReturnStmt node
type ReturnStmt struct {
	Ast     *ast.ReturnStmt
	Results []Expr
}

func (r *ReturnStmt) stmtNode() {}

func (r *ReturnStmt) Convert() string {
	var sb strings.Builder
	sb.WriteString("(return_statement")
	if len(r.Results) > 0 {
		sb.WriteString(" values: (expression_list")
		for _, result := range r.Results {
			sb.WriteString(" ")
			sb.WriteString(result.Convert())
		}
		sb.WriteString(")")
	}
	sb.WriteString(")")
	return sb.String()
}

func (r *ReturnStmt) Pos() Pos {
	return Pos(r.Ast.Pos())
}

// FuncDecl wraps an ast.FuncDecl node
type FuncDecl struct {
	Ast  *ast.FuncDecl
	Name *Ident
	Body Node
}

func (f *FuncDecl) declNode() {}

func (f *FuncDecl) Convert() string {
	var sb strings.Builder
	sb.WriteString("(function_declaration")
	if f.Name != nil {
		sb.WriteString(" name: ")
		sb.WriteString(f.Name.Convert())
	}
	if f.Body != nil {
		sb.WriteString(" body: ")
		sb.WriteString(f.Body.Convert())
	}
	sb.WriteString(")")
	return sb.String()
}

func (f *FuncDecl) Pos() Pos {
	return Pos(f.Ast.Pos())
}

// BuildAsqStmt converts an ast.Stmt to its corresponding metaq.Stmt
func BuildAsqStmt(stmt ast.Stmt, wildcardIdent map[*ast.Ident]bool) Stmt {
	if stmt == nil {
		return nil
	}

	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		return &ReturnStmt{
			Ast: s,
			Results: slicex.Map(s.Results, func(result ast.Expr) Expr {
				return BuildAsqExpr(result, wildcardIdent)
			}),
		}
	default:
		return &DefaultStmt{Ast: s}
	}
}

// BuildAsqDecl converts an ast.Decl to its corresponding metaq.Decl
func BuildAsqDecl(decl ast.Decl, wildcardIdent map[*ast.Ident]bool) Decl {
	if decl == nil {
		return nil
	}

	switch d := decl.(type) {
	case *ast.FuncDecl:
		return &FuncDecl{
			Ast:  d,
			Name: BuildAsqExpr(d.Name, wildcardIdent).(*Ident),
			Body: BuildAsqNode(d.Body, wildcardIdent),
		}
	default:
		return &DefaultDecl{Ast: d}
	}
}
