package metaq

import (
	"fmt"
	"github.com/StCredZero/asq/pkg/slicex"
	"go/ast"
	"io"
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

func (c *CallExpr) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(call_expression function: ")); err != nil {
		return err
	}
	if err := c.Fun.WriteTreeSitterQuery(w); err != nil {
		return err
	}
	_, err := w.Write([]byte(" arguments: (argument_list))"))
	return err
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

func (s *SelectorExpr) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(selector_expression operand: ")); err != nil {
		return err
	}
	if err := s.X.WriteTreeSitterQuery(w); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, ` field: (field_identifier) @field (#eq? @field "%s"))`, s.Ast.Sel.Name)
	return err
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

func (i *Ident) WriteTreeSitterQuery(w io.Writer) error {
	if i.Wildcard {
		_, err := w.Write([]byte("(identifier)"))
		return err
	}
	_, err := fmt.Fprintf(w, `(identifier) @name (#eq? @name "%s")`, i.Ast.Name)
	return err
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

func (d *DefaultNode) WriteTreeSitterQuery(w io.Writer) error {
	_, err := fmt.Fprintf(w, "(%T)", d.Node)
	return err
}

func (d *DefaultNode) AstNode() ast.Node {
	return nil
}

// DefaultExpr wraps any ast.Expr type that doesn't have a specific implementation
type DefaultExpr struct {
	Node ast.Node
}

func (d *DefaultExpr) exprNode() {}

func (d *DefaultExpr) WriteTreeSitterQuery(w io.Writer) error {
	_, err := fmt.Fprintf(w, "(%T)", d.Node)
	return err
}

func (d *DefaultExpr) AstNode() ast.Node {
	return nil
}

// DefaultStmt wraps any ast.Stmt type that doesn't have a specific implementation
type DefaultStmt struct {
	Ast ast.Stmt
}

func (d *DefaultStmt) stmtNode() {}

func (d *DefaultStmt) WriteTreeSitterQuery(w io.Writer) error {
	_, err := fmt.Fprintf(w, "(%T)", d.Ast)
	return err
}

func (d *DefaultStmt) AstNode() ast.Node {
	return d.Ast
}

// DefaultDecl wraps any ast.Decl type that doesn't have a specific implementation
type DefaultDecl struct {
	Ast ast.Decl
}

func (d *DefaultDecl) declNode() {}

func (d *DefaultDecl) WriteTreeSitterQuery(w io.Writer) error {
	_, err := fmt.Fprintf(w, "(%T)", d.Ast)
	return err
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

func (a *AssignStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(assignment_expression")); err != nil {
		return err
	}
	if _, err := w.Write([]byte(" left: ")); err != nil {
		return err
	}
	for i, lhs := range a.Lhs {
		if i > 0 {
			if _, err := w.Write([]byte(", ")); err != nil {
				return err
			}
		}
		if err := lhs.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte(" right: ")); err != nil {
		return err
	}
	for i, rhs := range a.Rhs {
		if i > 0 {
			if _, err := w.Write([]byte(", ")); err != nil {
				return err
			}
		}
		if err := rhs.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (a *AssignStmt) AstNode() ast.Node {
	return a.Ast
}

// BadStmt wraps an ast.BadStmt node
type BadStmt struct {
	Ast *ast.BadStmt
}

func (b *BadStmt) stmtNode() {}

func (b *BadStmt) WriteTreeSitterQuery(w io.Writer) error {
	_, err := w.Write([]byte("(bad_statement)"))
	return err
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

func (b *BlockStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(block")); err != nil {
		return err
	}
	for _, stmt := range b.List {
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
		if err := stmt.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (b *BranchStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(branch_statement")); err != nil {
		return err
	}
	if b.Label != nil {
		if _, err := w.Write([]byte(" label: ")); err != nil {
			return err
		}
		if err := b.Label.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (d *DeclStmt) WriteTreeSitterQuery(w io.Writer) error {
	return d.Decl.WriteTreeSitterQuery(w)
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

func (d *DeferStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(defer_statement expression: ")); err != nil {
		return err
	}
	if err := d.Call.WriteTreeSitterQuery(w); err != nil {
		return err
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (d *DeferStmt) AstNode() ast.Node {
	return d.Ast
}

// EmptyStmt wraps an ast.EmptyStmt node
type EmptyStmt struct {
	Ast *ast.EmptyStmt
}

func (e *EmptyStmt) stmtNode() {}

func (e *EmptyStmt) WriteTreeSitterQuery(w io.Writer) error {
	_, err := w.Write([]byte("(empty_statement)"))
	return err
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

func (e *ExprStmt) WriteTreeSitterQuery(w io.Writer) error {
	return e.X.WriteTreeSitterQuery(w)
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

func (g *GoStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(go_statement expression: ")); err != nil {
		return err
	}
	if err := g.Call.WriteTreeSitterQuery(w); err != nil {
		return err
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (i *IfStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(if_statement")); err != nil {
		return err
	}
	if i.Init != nil {
		if _, err := w.Write([]byte(" initializer: ")); err != nil {
			return err
		}
		if err := i.Init.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if i.Cond != nil {
		if _, err := w.Write([]byte(" condition: ")); err != nil {
			return err
		}
		if err := i.Cond.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if i.Body != nil {
		if _, err := w.Write([]byte(" consequence: ")); err != nil {
			return err
		}
		if err := i.Body.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if i.Else != nil {
		if _, err := w.Write([]byte(" alternative: ")); err != nil {
			return err
		}
		if err := i.Else.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (i *IncDecStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(inc_dec_statement expression: ")); err != nil {
		return err
	}
	if err := i.X.WriteTreeSitterQuery(w); err != nil {
		return err
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (l *LabeledStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(labeled_statement")); err != nil {
		return err
	}
	if l.Label != nil {
		if _, err := w.Write([]byte(" label: ")); err != nil {
			return err
		}
		if err := l.Label.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if l.Stmt != nil {
		if _, err := w.Write([]byte(" statement: ")); err != nil {
			return err
		}
		if err := l.Stmt.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (r *RangeStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(range_statement")); err != nil {
		return err
	}
	if r.Key != nil {
		if _, err := w.Write([]byte(" key: ")); err != nil {
			return err
		}
		if err := r.Key.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if r.Value != nil {
		if _, err := w.Write([]byte(" value: ")); err != nil {
			return err
		}
		if err := r.Value.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if r.X != nil {
		if _, err := w.Write([]byte(" expression: ")); err != nil {
			return err
		}
		if err := r.X.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if r.Body != nil {
		if _, err := w.Write([]byte(" body: ")); err != nil {
			return err
		}
		if err := r.Body.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (s *SelectStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(select_statement")); err != nil {
		return err
	}
	if s.Body != nil {
		if _, err := w.Write([]byte(" body: ")); err != nil {
			return err
		}
		if err := s.Body.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (s *SendStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(send_statement")); err != nil {
		return err
	}
	if s.Chan != nil {
		if _, err := w.Write([]byte(" channel: ")); err != nil {
			return err
		}
		if err := s.Chan.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if s.Value != nil {
		if _, err := w.Write([]byte(" value: ")); err != nil {
			return err
		}
		if err := s.Value.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (s *SwitchStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(switch_statement")); err != nil {
		return err
	}
	if s.Init != nil {
		if _, err := w.Write([]byte(" initializer: ")); err != nil {
			return err
		}
		if err := s.Init.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if s.Tag != nil {
		if _, err := w.Write([]byte(" value: ")); err != nil {
			return err
		}
		if err := s.Tag.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if s.Body != nil {
		if _, err := w.Write([]byte(" body: ")); err != nil {
			return err
		}
		if err := s.Body.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (t *TypeSwitchStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(type_switch_statement")); err != nil {
		return err
	}
	if t.Init != nil {
		if _, err := w.Write([]byte(" initializer: ")); err != nil {
			return err
		}
		if err := t.Init.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if t.Assign != nil {
		if _, err := w.Write([]byte(" assign: ")); err != nil {
			return err
		}
		if err := t.Assign.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if t.Body != nil {
		if _, err := w.Write([]byte(" body: ")); err != nil {
			return err
		}
		if err := t.Body.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (r *ReturnStmt) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(return_statement")); err != nil {
		return err
	}
	if len(r.Results) > 0 {
		if _, err := w.Write([]byte(" values: (expression_list")); err != nil {
			return err
		}
		for _, result := range r.Results {
			if _, err := w.Write([]byte(" ")); err != nil {
				return err
			}
			if ident, ok := result.(*Ident); ok {
				if _, err := fmt.Fprintf(w, `(identifier) @value (#eq? @value "%s")`, ident.Ast.Name); err != nil {
					return err
				}
			} else {
				if err := result.WriteTreeSitterQuery(w); err != nil {
					return err
				}
			}
		}
		if _, err := w.Write([]byte(")")); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
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

func (f *FuncDecl) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(function_declaration")); err != nil {
		return err
	}
	if f.Name != nil {
		if _, err := w.Write([]byte(" name: ")); err != nil {
			return err
		}
		if err := f.Name.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if f.Body != nil {
		if block, ok := f.Body.(*BlockStmt); ok && len(block.List) == 1 {
			if _, err := w.Write([]byte(" body: ")); err != nil {
				return err
			}
			if err := block.List[0].WriteTreeSitterQuery(w); err != nil {
				return err
			}
		} else {
			if _, err := w.Write([]byte(" body: ")); err != nil {
				return err
			}
			if err := f.Body.WriteTreeSitterQuery(w); err != nil {
				return err
			}
		}
	}
	_, err := w.Write([]byte(")"))
	return err
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
