package metaq

import (
	"fmt"
	"github.com/StCredZero/asq/pkg/slicex"
	"go/ast"
	"io"
	"strings"
)

// BuildAsqNode converts an ast.Node to its corresponding metaq.Node
func BuildAsqNode(node ast.Node, wildcardIdent map[*ast.Ident]bool) Node {
	if node == nil {
		return nil
	}

	switch astObj := node.(type) {
	case *ast.CallExpr:
		callExpr := &CallExpr{
			Ast: astObj,
			Fun: BuildAsqExpr(astObj.Fun, wildcardIdent),
			Args: slicex.Map(astObj.Args, func(arg ast.Expr) Expr {
				return BuildAsqExpr(arg, wildcardIdent)
			}),
		}
		callExpr.exprNode()
		return callExpr
	case *ast.SelectorExpr:
		selExpr := &SelectorExpr{
			Ast: astObj,
			X:   BuildAsqExpr(astObj.X, wildcardIdent),
		}
		selExpr.exprNode()
		return selExpr
	case *ast.Ident:
		ident := &Ident{
			Ast:      astObj,
			Wildcard: wildcardIdent[astObj],
		}
		ident.exprNode()
		return ident
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
	case *ast.Ident:
		return &Ident{
			Ast:      astObj,
			Wildcard: wildcardIdent[astObj],
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

func (c *CallExpr) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(call_expression function: ")
	sb.WriteString(c.Fun.WriteTreeSitterQuery(nil))
	sb.WriteString(" arguments: (argument_list))")
	return sb.String()
}

func (c *CallExpr) AstNode() ast.Node {
	return c.Ast
}

// SelectorExpr wraps an ast.SelectorExpr node
type SelectorExpr struct {
	Ast      *ast.SelectorExpr
	X        Expr
	Wildcard bool
}

func (s *SelectorExpr) exprNode() {}

func (s *SelectorExpr) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(selector_expression operand: ")
	sb.WriteString(s.X.WriteTreeSitterQuery(nil))
	sb.WriteString(fmt.Sprintf(` field: (field_identifier) @field (#eq? @field "%s"))`, s.Ast.Sel.Name))
	return sb.String()
}

func (s *SelectorExpr) AstNode() ast.Node {
	return s.Ast
}

// Ident wraps an ast.Ident node with an additional Wildcard field
type Ident struct {
	Ast      *ast.Ident
	Wildcard bool
}

func (i *Ident) exprNode() {}

func (i *Ident) WriteTreeSitterQuery(io.Writer) string {
	if i.Wildcard {
		return "(identifier)"
	}
	return fmt.Sprintf(`(identifier) @name (#eq? @name "%s")`, i.Ast.Name)
}

func (i *Ident) AstNode() ast.Node {
	return i.Ast
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

func (d *DefaultNode) WriteTreeSitterQuery(io.Writer) string {
	return fmt.Sprintf("(%T)", d.Node)
}

func (d *DefaultNode) AstNode() ast.Node {
	return nil
}

// DefaultExpr wraps any ast.Expr type that doesn't have a specific implementation
type DefaultExpr struct {
	Node ast.Node
}

func (d *DefaultExpr) exprNode() {}

func (d *DefaultExpr) WriteTreeSitterQuery(io.Writer) string {
	return fmt.Sprintf("(%T)", d.Node)
}

func (d *DefaultExpr) AstNode() ast.Node {
	return nil
}

// DefaultStmt wraps any ast.Stmt type that doesn't have a specific implementation
type DefaultStmt struct {
	Ast ast.Stmt
}

func (d *DefaultStmt) stmtNode() {}

func (d *DefaultStmt) WriteTreeSitterQuery(io.Writer) string {
	return fmt.Sprintf("(%T)", d.Ast)
}

func (d *DefaultStmt) AstNode() ast.Node {
	return d.Ast
}

// DefaultDecl wraps any ast.Decl type that doesn't have a specific implementation
type DefaultDecl struct {
	Ast ast.Decl
}

func (d *DefaultDecl) declNode() {}

func (d *DefaultDecl) WriteTreeSitterQuery(io.Writer) string {
	return fmt.Sprintf("(%T)", d.Ast)
}

func (d *DefaultDecl) AstNode() ast.Node {
	return d.Ast
}

// AssignStmt wraps an ast.AssignStmt node
type AssignStmt struct {
	Ast *ast.AssignStmt
	Lhs []Expr
	Rhs []Expr
}

func (a *AssignStmt) stmtNode() {}

func (a *AssignStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(assignment_expression")
	sb.WriteString(" left: ")
	for i, lhs := range a.Lhs {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(lhs.WriteTreeSitterQuery(nil))
	}
	sb.WriteString(" right: ")
	for i, rhs := range a.Rhs {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(rhs.WriteTreeSitterQuery(nil))
	}
	sb.WriteString(")")
	return sb.String()
}

func (a *AssignStmt) AstNode() ast.Node {
	return a.Ast
}

// BadStmt wraps an ast.BadStmt node
type BadStmt struct {
	Ast *ast.BadStmt
}

func (b *BadStmt) stmtNode() {}

func (b *BadStmt) WriteTreeSitterQuery(io.Writer) string {
	return "(bad_statement)"
}

func (b *BadStmt) AstNode() ast.Node {
	return b.Ast
}

// BlockStmt wraps an ast.BlockStmt node
type BlockStmt struct {
	Ast  *ast.BlockStmt
	List []Stmt
}

func (b *BlockStmt) stmtNode() {}

func (b *BlockStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(block")
	for _, stmt := range b.List {
		sb.WriteString(" ")
		sb.WriteString(stmt.WriteTreeSitterQuery(nil))
	}
	sb.WriteString(")")
	return sb.String()
}

func (b *BlockStmt) AstNode() ast.Node {
	return b.Ast
}

// BranchStmt wraps an ast.BranchStmt node
type BranchStmt struct {
	Ast   *ast.BranchStmt
	Label *Ident
}

func (b *BranchStmt) stmtNode() {}

func (b *BranchStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(branch_statement")
	if b.Label != nil {
		sb.WriteString(" label: ")
		sb.WriteString(b.Label.WriteTreeSitterQuery(nil))
	}
	sb.WriteString(")")
	return sb.String()
}

func (b *BranchStmt) AstNode() ast.Node {
	return b.Ast
}

// DeclStmt wraps an ast.DeclStmt node
type DeclStmt struct {
	Ast  *ast.DeclStmt
	Decl Decl
}

func (d *DeclStmt) stmtNode() {}

func (d *DeclStmt) WriteTreeSitterQuery(io.Writer) string {
	return d.Decl.WriteTreeSitterQuery(nil)
}

func (d *DeclStmt) AstNode() ast.Node {
	return d.Ast
}

// DeferStmt wraps an ast.DeferStmt node
type DeferStmt struct {
	Ast  *ast.DeferStmt
	Call Expr
}

func (d *DeferStmt) stmtNode() {}

func (d *DeferStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(defer_statement expression: ")
	sb.WriteString(d.Call.WriteTreeSitterQuery(nil))
	sb.WriteString(")")
	return sb.String()
}

func (d *DeferStmt) AstNode() ast.Node {
	return d.Ast
}

// EmptyStmt wraps an ast.EmptyStmt node
type EmptyStmt struct {
	Ast *ast.EmptyStmt
}

func (e *EmptyStmt) stmtNode() {}

func (e *EmptyStmt) WriteTreeSitterQuery(io.Writer) string {
	return "(empty_statement)"
}

func (e *EmptyStmt) AstNode() ast.Node {
	return e.Ast
}

// ExprStmt wraps an ast.ExprStmt node
type ExprStmt struct {
	Ast *ast.ExprStmt
	X   Expr
}

func (e *ExprStmt) stmtNode() {}

func (e *ExprStmt) WriteTreeSitterQuery(io.Writer) string {
	return e.X.WriteTreeSitterQuery(nil)
}

func (e *ExprStmt) AstNode() ast.Node {
	return e.Ast
}

// GoStmt wraps an ast.GoStmt node
type GoStmt struct {
	Ast  *ast.GoStmt
	Call Expr
}

func (g *GoStmt) stmtNode() {}

func (g *GoStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(go_statement expression: ")
	sb.WriteString(g.Call.WriteTreeSitterQuery(nil))
	sb.WriteString(")")
	return sb.String()
}

func (g *GoStmt) AstNode() ast.Node {
	return g.Ast
}

// IfStmt wraps an ast.IfStmt node
type IfStmt struct {
	Ast  *ast.IfStmt
	Init Stmt
	Cond Expr
	Body *BlockStmt
	Else Stmt
}

func (i *IfStmt) stmtNode() {}

func (i *IfStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(if_statement")
	if i.Init != nil {
		sb.WriteString(" initializer: ")
		sb.WriteString(i.Init.WriteTreeSitterQuery(nil))
	}
	if i.Cond != nil {
		sb.WriteString(" condition: ")
		sb.WriteString(i.Cond.WriteTreeSitterQuery(nil))
	}
	if i.Body != nil {
		sb.WriteString(" consequence: ")
		sb.WriteString(i.Body.WriteTreeSitterQuery(nil))
	}
	if i.Else != nil {
		sb.WriteString(" alternative: ")
		sb.WriteString(i.Else.WriteTreeSitterQuery(nil))
	}
	sb.WriteString(")")
	return sb.String()
}

func (i *IfStmt) AstNode() ast.Node {
	return i.Ast
}

// IncDecStmt wraps an ast.IncDecStmt node
type IncDecStmt struct {
	Ast *ast.IncDecStmt
	X   Expr
}

func (i *IncDecStmt) stmtNode() {}

func (i *IncDecStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(inc_dec_statement expression: ")
	sb.WriteString(i.X.WriteTreeSitterQuery(nil))
	sb.WriteString(")")
	return sb.String()
}

func (i *IncDecStmt) AstNode() ast.Node {
	return i.Ast
}

// LabeledStmt wraps an ast.LabeledStmt node
type LabeledStmt struct {
	Ast   *ast.LabeledStmt
	Label *Ident
	Stmt  Stmt
}

func (l *LabeledStmt) stmtNode() {}

func (l *LabeledStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(labeled_statement")
	if l.Label != nil {
		sb.WriteString(" label: ")
		sb.WriteString(l.Label.WriteTreeSitterQuery(nil))
	}
	if l.Stmt != nil {
		sb.WriteString(" statement: ")
		sb.WriteString(l.Stmt.WriteTreeSitterQuery(nil))
	}
	sb.WriteString(")")
	return sb.String()
}

func (l *LabeledStmt) AstNode() ast.Node {
	return l.Ast
}

// RangeStmt wraps an ast.RangeStmt node
type RangeStmt struct {
	Ast   *ast.RangeStmt
	Key   Expr
	Value Expr
	X     Expr
	Body  *BlockStmt
}

func (r *RangeStmt) stmtNode() {}

func (r *RangeStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(range_statement")
	if r.Key != nil {
		sb.WriteString(" key: ")
		sb.WriteString(r.Key.WriteTreeSitterQuery(nil))
	}
	if r.Value != nil {
		sb.WriteString(" value: ")
		sb.WriteString(r.Value.WriteTreeSitterQuery(nil))
	}
	if r.X != nil {
		sb.WriteString(" expression: ")
		sb.WriteString(r.X.WriteTreeSitterQuery(nil))
	}
	if r.Body != nil {
		sb.WriteString(" body: ")
		sb.WriteString(r.Body.WriteTreeSitterQuery(nil))
	}
	sb.WriteString(")")
	return sb.String()
}

func (r *RangeStmt) AstNode() ast.Node {
	return r.Ast
}

// SelectStmt wraps an ast.SelectStmt node
type SelectStmt struct {
	Ast  *ast.SelectStmt
	Body *BlockStmt
}

func (s *SelectStmt) stmtNode() {}

func (s *SelectStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(select_statement")
	if s.Body != nil {
		sb.WriteString(" body: ")
		sb.WriteString(s.Body.WriteTreeSitterQuery(nil))
	}
	sb.WriteString(")")
	return sb.String()
}

func (s *SelectStmt) AstNode() ast.Node {
	return s.Ast
}

// SendStmt wraps an ast.SendStmt node
type SendStmt struct {
	Ast   *ast.SendStmt
	Chan  Expr
	Value Expr
}

func (s *SendStmt) stmtNode() {}

func (s *SendStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(send_statement")
	if s.Chan != nil {
		sb.WriteString(" channel: ")
		sb.WriteString(s.Chan.WriteTreeSitterQuery(nil))
	}
	if s.Value != nil {
		sb.WriteString(" value: ")
		sb.WriteString(s.Value.WriteTreeSitterQuery(nil))
	}
	sb.WriteString(")")
	return sb.String()
}

func (s *SendStmt) AstNode() ast.Node {
	return s.Ast
}

// SwitchStmt wraps an ast.SwitchStmt node
type SwitchStmt struct {
	Ast  *ast.SwitchStmt
	Init Stmt
	Tag  Expr
	Body *BlockStmt
}

func (s *SwitchStmt) stmtNode() {}

func (s *SwitchStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(switch_statement")
	if s.Init != nil {
		sb.WriteString(" initializer: ")
		sb.WriteString(s.Init.WriteTreeSitterQuery(nil))
	}
	if s.Tag != nil {
		sb.WriteString(" value: ")
		sb.WriteString(s.Tag.WriteTreeSitterQuery(nil))
	}
	if s.Body != nil {
		sb.WriteString(" body: ")
		sb.WriteString(s.Body.WriteTreeSitterQuery(nil))
	}
	sb.WriteString(")")
	return sb.String()
}

func (s *SwitchStmt) AstNode() ast.Node {
	return s.Ast
}

// TypeSwitchStmt wraps an ast.TypeSwitchStmt node
type TypeSwitchStmt struct {
	Ast    *ast.TypeSwitchStmt
	Init   Stmt
	Assign Stmt
	Body   *BlockStmt
}

func (t *TypeSwitchStmt) stmtNode() {}

func (t *TypeSwitchStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(type_switch_statement")
	if t.Init != nil {
		sb.WriteString(" initializer: ")
		sb.WriteString(t.Init.WriteTreeSitterQuery(nil))
	}
	if t.Assign != nil {
		sb.WriteString(" assign: ")
		sb.WriteString(t.Assign.WriteTreeSitterQuery(nil))
	}
	if t.Body != nil {
		sb.WriteString(" body: ")
		sb.WriteString(t.Body.WriteTreeSitterQuery(nil))
	}
	sb.WriteString(")")
	return sb.String()
}

func (t *TypeSwitchStmt) AstNode() ast.Node {
	return t.Ast
}

// ReturnStmt wraps an ast.ReturnStmt node
type ReturnStmt struct {
	Ast     *ast.ReturnStmt
	Results []Expr
}

func (r *ReturnStmt) stmtNode() {}

func (r *ReturnStmt) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(return_statement")
	if len(r.Results) > 0 {
		sb.WriteString(" values: (expression_list")
		for _, result := range r.Results {
			sb.WriteString(" ")
			if ident, ok := result.(*Ident); ok {
				sb.WriteString(fmt.Sprintf(`(identifier) @value (#eq? @value "%s")`, ident.Ast.Name))
			} else {
				sb.WriteString(result.WriteTreeSitterQuery(nil))
			}
		}
		sb.WriteString(")")
	}
	sb.WriteString(")")
	return sb.String()
}

func (r *ReturnStmt) AstNode() ast.Node {
	return r.Ast
}

// FuncDecl wraps an ast.FuncDecl node
type FuncDecl struct {
	Ast  *ast.FuncDecl
	Name *Ident
	Body Node
}

func (f *FuncDecl) declNode() {}

func (f *FuncDecl) WriteTreeSitterQuery(io.Writer) string {
	var sb strings.Builder
	sb.WriteString("(function_declaration")
	if f.Name != nil {
		sb.WriteString(" name: ")
		sb.WriteString(f.Name.WriteTreeSitterQuery(nil))
	}
	if f.Body != nil {
		if block, ok := f.Body.(*BlockStmt); ok && len(block.List) == 1 {
			sb.WriteString(" body: ")
			sb.WriteString(block.List[0].WriteTreeSitterQuery(nil))
		} else {
			sb.WriteString(" body: ")
			sb.WriteString(f.Body.WriteTreeSitterQuery(nil))
		}
	}
	sb.WriteString(")")
	return sb.String()
}

func (f *FuncDecl) AstNode() ast.Node {
	return f.Ast
}

// BuildAsqStmt converts an ast.Stmt to its corresponding metaq.Stmt
func BuildAsqStmt(stmt ast.Stmt, wildcardIdent map[*ast.Ident]bool) Stmt {
	if stmt == nil {
		return nil
	}

	switch s := stmt.(type) {
	case *ast.AssignStmt:
		return &AssignStmt{
			Ast: s,
			Lhs: slicex.Map(s.Lhs, func(lhs ast.Expr) Expr {
				return BuildAsqExpr(lhs, wildcardIdent)
			}),
			Rhs: slicex.Map(s.Rhs, func(rhs ast.Expr) Expr {
				return BuildAsqExpr(rhs, wildcardIdent)
			}),
		}
	case *ast.BadStmt:
		return &BadStmt{Ast: s}
	case *ast.BlockStmt:
		return &BlockStmt{
			Ast: s,
			List: slicex.Map(s.List, func(stmt ast.Stmt) Stmt {
				return BuildAsqStmt(stmt, wildcardIdent)
			}),
		}
	case *ast.BranchStmt:
		var label *Ident
		if s.Label != nil {
			label = &Ident{Ast: s.Label}
		}
		return &BranchStmt{
			Ast:   s,
			Label: label,
		}
	case *ast.DeclStmt:
		return &DeclStmt{
			Ast:  s,
			Decl: BuildAsqDecl(s.Decl, wildcardIdent),
		}
	case *ast.DeferStmt:
		return &DeferStmt{
			Ast:  s,
			Call: BuildAsqExpr(s.Call, wildcardIdent),
		}
	case *ast.EmptyStmt:
		return &EmptyStmt{Ast: s}
	case *ast.ExprStmt:
		return &ExprStmt{
			Ast: s,
			X:   BuildAsqExpr(s.X, wildcardIdent),
		}
	case *ast.GoStmt:
		return &GoStmt{
			Ast:  s,
			Call: BuildAsqExpr(s.Call, wildcardIdent),
		}
	case *ast.IfStmt:
		return &IfStmt{
			Ast:  s,
			Init: BuildAsqStmt(s.Init, wildcardIdent),
			Cond: BuildAsqExpr(s.Cond, wildcardIdent),
			Body: BuildAsqStmt(s.Body, wildcardIdent).(*BlockStmt),
			Else: BuildAsqStmt(s.Else, wildcardIdent),
		}
	case *ast.IncDecStmt:
		return &IncDecStmt{
			Ast: s,
			X:   BuildAsqExpr(s.X, wildcardIdent),
		}
	case *ast.LabeledStmt:
		var label *Ident
		if s.Label != nil {
			label = &Ident{Ast: s.Label}
		}
		return &LabeledStmt{
			Ast:   s,
			Label: label,
			Stmt:  BuildAsqStmt(s.Stmt, wildcardIdent),
		}
	case *ast.RangeStmt:
		return &RangeStmt{
			Ast:   s,
			Key:   BuildAsqExpr(s.Key, wildcardIdent),
			Value: BuildAsqExpr(s.Value, wildcardIdent),
			X:     BuildAsqExpr(s.X, wildcardIdent),
			Body:  BuildAsqStmt(s.Body, wildcardIdent).(*BlockStmt),
		}
	case *ast.ReturnStmt:
		return &ReturnStmt{
			Ast: s,
			Results: slicex.Map(s.Results, func(result ast.Expr) Expr {
				return BuildAsqExpr(result, wildcardIdent)
			}),
		}
	case *ast.SelectStmt:
		return &SelectStmt{
			Ast:  s,
			Body: BuildAsqStmt(s.Body, wildcardIdent).(*BlockStmt),
		}
	case *ast.SendStmt:
		return &SendStmt{
			Ast:   s,
			Chan:  BuildAsqExpr(s.Chan, wildcardIdent),
			Value: BuildAsqExpr(s.Value, wildcardIdent),
		}
	case *ast.SwitchStmt:
		return &SwitchStmt{
			Ast:  s,
			Init: BuildAsqStmt(s.Init, wildcardIdent),
			Tag:  BuildAsqExpr(s.Tag, wildcardIdent),
			Body: BuildAsqStmt(s.Body, wildcardIdent).(*BlockStmt),
		}
	case *ast.TypeSwitchStmt:
		return &TypeSwitchStmt{
			Ast:    s,
			Init:   BuildAsqStmt(s.Init, wildcardIdent),
			Assign: BuildAsqStmt(s.Assign, wildcardIdent),
			Body:   BuildAsqStmt(s.Body, wildcardIdent).(*BlockStmt),
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
		var name *Ident
		if d.Name != nil {
			name = &Ident{Ast: d.Name}
		}
		return &FuncDecl{
			Ast:  d,
			Name: name,
			Body: BuildAsqNode(d.Body, wildcardIdent),
		}
	default:
		return &DefaultDecl{Ast: d}
	}
}
