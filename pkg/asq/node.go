package asq

import (
	"fmt"
	"github.com/StCredZero/asq/pkg/slicex"
	"go/ast"
	"go/token"
	"io"
)

// BuildAsqNode converts an ast.Node to its corresponding asq.Node
func BuildAsqNode(node ast.Node, p *passOne) Node {
	if node == nil {
		return nil
	}

	switch astObj := node.(type) {
	case *ast.CallExpr:
		callExpr := &CallExpr{
			Ast: astObj,
			Fun: BuildAsqExpr(astObj.Fun, p),
			Args: slicex.Map(astObj.Args, func(arg ast.Expr) Expr {
				return BuildAsqExpr(arg, p)
			}),
		}
		callExpr.exprNode()
		return callExpr
	case *ast.SelectorExpr:
		selExpr := &SelectorExpr{
			Ast: astObj,
			X:   BuildAsqExpr(astObj.X, p),
		}
		selExpr.exprNode()
		return selExpr
	case *ast.Ident:
		ident := &Ident{
			Ast:      astObj,
			Wildcard: p.isWildcard(astObj),
		}
		ident.exprNode()
		return ident
	case *ast.ArrayType:
		var length Expr
		if astObj.Len != nil {
			if basicLit, ok := astObj.Len.(*ast.BasicLit); ok {
				length = &BasicLit{Ast: basicLit}
			} else {
				length = BuildAsqExpr(astObj.Len, p)
			}
		}
		return &ArrayType{
			Ast: astObj,
			Len: length,
			Elt: BuildAsqNode(astObj.Elt, p),
		}
	case *ast.BasicLit:
		return &BasicLit{
			Ast: astObj,
		}
	case *ast.ChanType:
		return &ChanType{
			Ast:   astObj,
			Value: BuildAsqNode(astObj.Value, p),
		}
	case *ast.CompositeLit:
		var elts []Expr
		for _, elt := range astObj.Elts {
			if basicLit, ok := elt.(*ast.BasicLit); ok {
				elts = append(elts, &BasicLit{Ast: basicLit})
			} else {
				elts = append(elts, BuildAsqExpr(elt, p))
			}
		}
		return &CompositeLit{
			Ast:  astObj,
			Type: BuildAsqNode(astObj.Type, p),
			Elts: elts,
		}
	case *ast.Field:
		var names []*Ident
		for _, name := range astObj.Names {
			if expr := BuildAsqExpr(name, p); expr != nil {
				if ident, ok := expr.(*Ident); ok {
					names = append(names, ident)
				}
			}
		}
		return &Field{
			Ast:   astObj,
			Names: names,
			Type:  BuildAsqNode(astObj.Type, p),
			Tag: func() *BasicLit {
				if astObj.Tag == nil {
					return nil
				}
				if expr := BuildAsqExpr(astObj.Tag, p); expr != nil {
					if tag, ok := expr.(*BasicLit); ok {
						return tag
					}
				}
				return nil
			}(),
		}
	case *ast.FieldList:
		var fields []*Field
		if astObj.List != nil {
			fields = slicex.Map(astObj.List, func(field *ast.Field) *Field {
				if node := BuildAsqNode(field, p); node != nil {
					return node.(*Field)
				}
				return nil
			})
		}
		return &FieldList{
			Ast:  astObj,
			List: fields,
		}
	case *ast.FuncLit:
		return &FuncLit{
			Ast:  astObj,
			Type: BuildAsqNode(astObj.Type, p).(*FuncType),
			Body: BuildAsqNode(astObj.Body, p),
		}
	case *ast.FuncType:
		var params, results *FieldList
		if astObj.Params != nil {
			if node := BuildAsqNode(astObj.Params, p); node != nil {
				params = node.(*FieldList)
			}
		}
		if astObj.Results != nil {
			if node := BuildAsqNode(astObj.Results, p); node != nil {
				results = node.(*FieldList)
			}
		}
		return &FuncType{
			Ast:     astObj,
			Params:  params,
			Results: results,
		}
	case *ast.MapType:
		return &MapType{
			Ast:   astObj,
			Key:   BuildAsqNode(astObj.Key, p),
			Value: BuildAsqNode(astObj.Value, p),
		}
	case *ast.StructType:
		return &StructType{
			Ast:    astObj,
			Fields: BuildAsqNode(astObj.Fields, p).(*FieldList),
		}
	case *ast.TypeSpec:
		return &TypeSpec{
			Ast:  astObj,
			Name: BuildAsqExpr(astObj.Name, p).(*Ident),
			Type: BuildAsqNode(astObj.Type, p),
		}
	case *ast.ValueSpec:
		var values []Expr
		for _, val := range astObj.Values {
			if basicLit, ok := val.(*ast.BasicLit); ok {
				values = append(values, &BasicLit{Ast: basicLit})
			} else {
				values = append(values, BuildAsqExpr(val, p))
			}
		}
		return &ValueSpec{
			Ast: astObj,
			Names: slicex.Map(astObj.Names, func(name *ast.Ident) *Ident {
				return BuildAsqExpr(name, p).(*Ident)
			}),
			Type:   BuildAsqNode(astObj.Type, p),
			Values: values,
		}
	case *ast.File:
		if astObj.Name == nil {
			return nil
		}
		return &Package{
			Ast:   &ast.Package{Name: astObj.Name.Name},
			Name:  BuildAsqNode(astObj.Name, p).(*Ident),
			Files: make(map[string]Node),
		}
	case *ast.FuncDecl:
		var name *Ident
		if astObj.Name != nil {
			name = BuildAsqExpr(astObj.Name, p).(*Ident)
		}
		var funcType *FuncType
		if astObj.Type != nil {
			if typeNode := BuildAsqNode(astObj.Type, p); typeNode != nil {
				funcType = typeNode.(*FuncType)
			}
		}
		var body Node
		if astObj.Body != nil && len(astObj.Body.List) == 1 {
			body = BuildAsqStmt(astObj.Body.List[0], p)
		}
		return &FuncDecl{
			Ast:  astObj,
			Name: name,
			Type: funcType,
			Body: body,
		}
	case ast.Stmt:
		return BuildAsqStmt(astObj, p)
	case ast.Decl:
		return BuildAsqDecl(astObj, p)
	default:
		return &DefaultNode{Node: astObj}
	}
}

// BuildAsqExpr converts an ast.Node to its corresponding asq.Node
func BuildAsqExpr(node ast.Node, p *passOne) Expr {
	if node == nil {
		return nil
	}

	switch astObj := node.(type) {
	case *ast.CallExpr:
		return &CallExpr{
			Ast: astObj,
			Fun: BuildAsqExpr(astObj.Fun, p),
			Args: slicex.Map(astObj.Args, func(arg ast.Expr) Expr {
				return BuildAsqExpr(arg, p)
			}),
		}
	case *ast.BinaryExpr:
		return &BinaryExpr{
			Ast: astObj,
			X:   BuildAsqExpr(astObj.X, p),
			Op:  astObj.Op,
			Y:   BuildAsqExpr(astObj.Y, p),
		}
	case *ast.SelectorExpr:
		return &SelectorExpr{
			Ast: astObj,
			X:   BuildAsqExpr(astObj.X, p),
		}
	case *ast.Ident:
		return &Ident{
			Ast:      astObj,
			Wildcard: p.isWildcard(astObj),
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

// ArrayType wraps an ast.ArrayType node
type ArrayType struct {
	Ast *ast.ArrayType
	Len Expr
	Elt Node
}

func (a *ArrayType) exprNode() {}

func (a *ArrayType) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(array_type")); err != nil {
		return err
	}
	if a.Len != nil {
		if _, err := w.Write([]byte(" length: ")); err != nil {
			return err
		}
		if basicLit, ok := a.Len.(*BasicLit); ok {
			if err := basicLit.WriteTreeSitterQuery(w); err != nil {
				return err
			}
		} else {
			if err := a.Len.WriteTreeSitterQuery(w); err != nil {
				return err
			}
		}
	}
	if a.Elt != nil {
		if _, err := w.Write([]byte(" element: ")); err != nil {
			return err
		}
		if err := a.Elt.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (a *ArrayType) AstNode() ast.Node {
	return a.Ast
}

// BadDecl wraps an ast.BadDecl node
type BadDecl struct {
	Ast *ast.BadDecl
}

func (b *BadDecl) declNode() {}

func (b *BadDecl) WriteTreeSitterQuery(w io.Writer) error {
	_, err := w.Write([]byte("(bad_declaration)"))
	return err
}

func (b *BadDecl) AstNode() ast.Node {
	return b.Ast
}

// BasicLit wraps an ast.BasicLit node
type BasicLit struct {
	Ast *ast.BasicLit
}

func (b *BasicLit) exprNode() {}

func (b *BasicLit) WriteTreeSitterQuery(w io.Writer) error {
	_, err := fmt.Fprintf(w, `(literal) @value (#eq? @value "%s")`, b.Ast.Value)
	return err
}

func (b *BasicLit) AstNode() ast.Node {
	return b.Ast
}

// ChanType wraps an ast.ChanType node
type ChanType struct {
	Ast   *ast.ChanType
	Value Node
}

func (c *ChanType) exprNode() {}

func (c *ChanType) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(channel_type")); err != nil {
		return err
	}
	if c.Value != nil {
		if _, err := w.Write([]byte(" value: ")); err != nil {
			return err
		}
		if err := c.Value.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (c *ChanType) AstNode() ast.Node {
	return c.Ast
}

// CompositeLit wraps an ast.CompositeLit node
type CompositeLit struct {
	Ast  *ast.CompositeLit
	Type Node
	Elts []Expr
}

func (c *CompositeLit) exprNode() {}

func (c *CompositeLit) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(composite_literal")); err != nil {
		return err
	}
	if c.Type != nil {
		if _, err := w.Write([]byte(" type: ")); err != nil {
			return err
		}
		if err := c.Type.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if len(c.Elts) > 0 {
		if _, err := w.Write([]byte(" elements: (")); err != nil {
			return err
		}
		for i, elt := range c.Elts {
			if i > 0 {
				if _, err := w.Write([]byte(" ")); err != nil {
					return err
				}
			}
			if err := elt.WriteTreeSitterQuery(w); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte(")")); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (c *CompositeLit) AstNode() ast.Node {
	return c.Ast
}

// Field wraps an ast.Field node
type Field struct {
	Ast   *ast.Field
	Names []*Ident
	Type  Node
	Tag   *BasicLit
}

func (f *Field) declNode() {}

func (f *Field) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(field_declaration")); err != nil {
		return err
	}
	if len(f.Names) > 0 {
		if _, err := w.Write([]byte(" names: (")); err != nil {
			return err
		}
		for i, name := range f.Names {
			if i > 0 {
				if _, err := w.Write([]byte(" ")); err != nil {
					return err
				}
			}
			if err := name.WriteTreeSitterQuery(w); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte(")")); err != nil {
			return err
		}
	}
	if f.Type != nil {
		if _, err := w.Write([]byte(" type: ")); err != nil {
			return err
		}
		if err := f.Type.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if f.Tag != nil {
		if _, err := w.Write([]byte(" tag: ")); err != nil {
			return err
		}
		if err := f.Tag.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (f *Field) AstNode() ast.Node {
	return f.Ast
}

// FieldList wraps an ast.FieldList node
type FieldList struct {
	Ast  *ast.FieldList
	List []*Field
}

func (f *FieldList) declNode() {}

func (f *FieldList) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(field_list")); err != nil {
		return err
	}
	for _, field := range f.List {
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
		if err := field.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (f *FieldList) AstNode() ast.Node {
	return f.Ast
}

// FuncLit wraps an ast.FuncLit node
type FuncLit struct {
	Ast  *ast.FuncLit
	Type *FuncType
	Body Node
}

func (f *FuncLit) exprNode() {}

func (f *FuncLit) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(function_literal")); err != nil {
		return err
	}
	if f.Type != nil {
		if _, err := w.Write([]byte(" type: ")); err != nil {
			return err
		}
		if err := f.Type.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (f *FuncLit) AstNode() ast.Node {
	return f.Ast
}

// FuncType wraps an ast.FuncType node
type FuncType struct {
	Ast     *ast.FuncType
	Params  *FieldList
	Results *FieldList
}

func (f *FuncType) exprNode() {}

func (f *FuncType) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(function_type")); err != nil {
		return err
	}
	if f.Params != nil {
		if _, err := w.Write([]byte(" parameters: ")); err != nil {
			return err
		}
		if err := f.Params.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if f.Results != nil {
		if _, err := w.Write([]byte(" results: ")); err != nil {
			return err
		}
		if err := f.Results.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (f *FuncType) AstNode() ast.Node {
	return f.Ast
}

// GenDecl wraps an ast.GenDecl node
type GenDecl struct {
	Ast   *ast.GenDecl
	Specs []Node
}

func (g *GenDecl) declNode() {}

func (g *GenDecl) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(generic_declaration ")); err != nil {
		return err
	}
	for i, spec := range g.Specs {
		if i > 0 {
			if _, err := w.Write([]byte(" ")); err != nil {
				return err
			}
		}
		if valueSpec, ok := spec.(*ValueSpec); ok {
			if err := valueSpec.WriteTreeSitterQuery(w); err != nil {
				return err
			}
		} else {
			if err := spec.WriteTreeSitterQuery(w); err != nil {
				return err
			}
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (g *GenDecl) AstNode() ast.Node {
	return g.Ast
}

// MapType wraps an ast.MapType node
type MapType struct {
	Ast   *ast.MapType
	Key   Node
	Value Node
}

func (m *MapType) exprNode() {}

func (m *MapType) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(map_type")); err != nil {
		return err
	}
	if m.Key != nil {
		if _, err := w.Write([]byte(" key: ")); err != nil {
			return err
		}
		if err := m.Key.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if m.Value != nil {
		if _, err := w.Write([]byte(" value: ")); err != nil {
			return err
		}
		if err := m.Value.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (m *MapType) AstNode() ast.Node {
	return m.Ast
}

// Package wraps an ast.Package node
type Package struct {
	Ast   *ast.Package
	Name  *Ident
	Files map[string]Node
}

func (p *Package) declNode() {}

func (p *Package) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(source_file package_name: ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, `(identifier) @name (#eq? @name "%s")`, p.Name.Ast.Name); err != nil {
		return err
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (p *Package) AstNode() ast.Node {
	return p.Ast
}

// StructType wraps an ast.StructType node
type StructType struct {
	Ast    *ast.StructType
	Fields *FieldList
}

func (s *StructType) exprNode() {}

func (s *StructType) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(struct_type")); err != nil {
		return err
	}
	if s.Fields != nil {
		if _, err := w.Write([]byte(" fields: ")); err != nil {
			return err
		}
		if err := s.Fields.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (s *StructType) AstNode() ast.Node {
	return s.Ast
}

// TypeSpec wraps an ast.TypeSpec node
type TypeSpec struct {
	Ast  *ast.TypeSpec
	Name *Ident
	Type Node
}

func (t *TypeSpec) declNode() {}

func (t *TypeSpec) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(type_spec")); err != nil {
		return err
	}
	if t.Name != nil {
		if _, err := w.Write([]byte(" name: ")); err != nil {
			return err
		}
		if err := t.Name.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if t.Type != nil {
		if _, err := w.Write([]byte(" type: ")); err != nil {
			return err
		}
		if err := t.Type.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (t *TypeSpec) AstNode() ast.Node {
	return t.Ast
}

// ValueSpec wraps an ast.ValueSpec node
type ValueSpec struct {
	Ast    *ast.ValueSpec
	Names  []*Ident
	Type   Node
	Values []Expr
}

func (v *ValueSpec) declNode() {}

func (v *ValueSpec) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(value_spec")); err != nil {
		return err
	}
	if len(v.Names) > 0 {
		if _, err := w.Write([]byte(" names: (")); err != nil {
			return err
		}
		for i, name := range v.Names {
			if i > 0 {
				if _, err := w.Write([]byte(" ")); err != nil {
					return err
				}
			}
			if err := name.WriteTreeSitterQuery(w); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte(")")); err != nil {
			return err
		}
	}
	if v.Type != nil {
		if _, err := w.Write([]byte(" type: ")); err != nil {
			return err
		}
		if err := v.Type.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if len(v.Values) > 0 {
		if _, err := w.Write([]byte(" values: (")); err != nil {
			return err
		}
		for i, value := range v.Values {
			if i > 0 {
				if _, err := w.Write([]byte(" ")); err != nil {
					return err
				}
			}
			if err := value.WriteTreeSitterQuery(w); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte(")")); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (v *ValueSpec) AstNode() ast.Node {
	return v.Ast
}

// FuncDecl wraps an ast.FuncDecl node
type FuncDecl struct {
	Ast  *ast.FuncDecl
	Name *Ident
	Type *FuncType
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
	if f.Type != nil && f.Type.Params != nil && len(f.Type.Params.List) > 0 {
		if _, err := w.Write([]byte(" parameters: ")); err != nil {
			return err
		}
		if err := f.Type.Params.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if f.Type != nil && f.Type.Results != nil && len(f.Type.Results.List) > 0 {
		if _, err := w.Write([]byte(" results: ")); err != nil {
			return err
		}
		if err := f.Type.Results.WriteTreeSitterQuery(w); err != nil {
			return err
		}
	}
	if f.Body != nil {
		if _, err := w.Write([]byte(" body: ")); err != nil {
			return err
		}
		// Check if body is a BlockStmt with a single ReturnStmt
		if blockStmt, ok := f.Body.(*BlockStmt); ok && len(blockStmt.List) == 1 {
			if returnStmt, ok := blockStmt.List[0].(*ReturnStmt); ok {
				// Write the ReturnStmt directly without block wrapper
				if err := returnStmt.WriteTreeSitterQuery(w); err != nil {
					return err
				}
			} else {
				// For all other single statements, wrap in block
				if _, err := w.Write([]byte("(block ")); err != nil {
					return err
				}
				if err := f.Body.WriteTreeSitterQuery(w); err != nil {
					return err
				}
				if _, err := w.Write([]byte(")")); err != nil {
					return err
				}
			}
		} else {
			// Not a BlockStmt or has multiple statements
			if _, err := w.Write([]byte("(block ")); err != nil {
				return err
			}
			if err := f.Body.WriteTreeSitterQuery(w); err != nil {
				return err
			}
			if _, err := w.Write([]byte(")")); err != nil {
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
type BinaryExpr struct {
	Ast *ast.BinaryExpr
	X   Expr
	Op  token.Token
	Y   Expr
}

func (b *BinaryExpr) exprNode() {}

func (b *BinaryExpr) WriteTreeSitterQuery(w io.Writer) error {
	if _, err := w.Write([]byte("(binary_expression left: ")); err != nil {
		return err
	}
	if err := b.X.WriteTreeSitterQuery(w); err != nil {
		return err
	}
	if _, err := w.Write([]byte(fmt.Sprintf(` operator: "%s" right: `, b.Op.String()))); err != nil {
		return err
	}
	if err := b.Y.WriteTreeSitterQuery(w); err != nil {
		return err
	}
	_, err := w.Write([]byte(")"))
	return err
}

func (b *BinaryExpr) AstNode() ast.Node {
	return b.Ast
}

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
		if _, err := w.Write([]byte(" (expression_list ")); err != nil {
			return err
		}
		for i, result := range r.Results {
			if i > 0 {
				if _, err := w.Write([]byte(" ")); err != nil {
					return err
				}
			}
			if ident, ok := result.(*Ident); ok {
				if ident.Ast.Name == "true" {
					if _, err := w.Write([]byte("(true)")); err != nil {
						return err
					}
				} else if _, err := fmt.Fprintf(w, `(identifier) @value (#eq? @value "%s")`, ident.Ast.Name); err != nil {
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

// Removed duplicate FuncDecl implementation

// BuildAsqStmt converts an ast.Stmt to its corresponding asq.Stmt
func BuildAsqStmt(stmt ast.Stmt, p *passOne) Stmt {
	if stmt == nil {
		return nil
	}

	switch s := stmt.(type) {
	case *ast.AssignStmt:
		return &AssignStmt{
			Ast: s,
			Lhs: slicex.Map(s.Lhs, func(lhs ast.Expr) Expr {
				return BuildAsqExpr(lhs, p)
			}),
			Rhs: slicex.Map(s.Rhs, func(rhs ast.Expr) Expr {
				return BuildAsqExpr(rhs, p)
			}),
		}
	case *ast.BadStmt:
		return &BadStmt{Ast: s}
	case *ast.BlockStmt:
		return &BlockStmt{
			Ast: s,
			List: slicex.Map(s.List, func(stmt ast.Stmt) Stmt {
				return BuildAsqStmt(stmt, p)
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
			Decl: BuildAsqDecl(s.Decl, p),
		}
	case *ast.DeferStmt:
		return &DeferStmt{
			Ast:  s,
			Call: BuildAsqExpr(s.Call, p),
		}
	case *ast.EmptyStmt:
		return &EmptyStmt{Ast: s}
	case *ast.ExprStmt:
		return &ExprStmt{
			Ast: s,
			X:   BuildAsqExpr(s.X, p),
		}
	case *ast.GoStmt:
		return &GoStmt{
			Ast:  s,
			Call: BuildAsqExpr(s.Call, p),
		}
	case *ast.IfStmt:
		return &IfStmt{
			Ast:  s,
			Init: BuildAsqStmt(s.Init, p),
			Cond: BuildAsqExpr(s.Cond, p),
			Body: BuildAsqStmt(s.Body, p).(*BlockStmt),
			Else: BuildAsqStmt(s.Else, p),
		}
	case *ast.IncDecStmt:
		return &IncDecStmt{
			Ast: s,
			X:   BuildAsqExpr(s.X, p),
		}
	case *ast.LabeledStmt:
		var label *Ident
		if s.Label != nil {
			label = &Ident{Ast: s.Label}
		}
		return &LabeledStmt{
			Ast:   s,
			Label: label,
			Stmt:  BuildAsqStmt(s.Stmt, p),
		}
	case *ast.RangeStmt:
		return &RangeStmt{
			Ast:   s,
			Key:   BuildAsqExpr(s.Key, p),
			Value: BuildAsqExpr(s.Value, p),
			X:     BuildAsqExpr(s.X, p),
			Body:  BuildAsqStmt(s.Body, p).(*BlockStmt),
		}
	case *ast.ReturnStmt:
		return &ReturnStmt{
			Ast: s,
			Results: slicex.Map(s.Results, func(result ast.Expr) Expr {
				return BuildAsqExpr(result, p)
			}),
		}
	case *ast.SelectStmt:
		return &SelectStmt{
			Ast:  s,
			Body: BuildAsqStmt(s.Body, p).(*BlockStmt),
		}
	case *ast.SendStmt:
		return &SendStmt{
			Ast:   s,
			Chan:  BuildAsqExpr(s.Chan, p),
			Value: BuildAsqExpr(s.Value, p),
		}
	case *ast.SwitchStmt:
		return &SwitchStmt{
			Ast:  s,
			Init: BuildAsqStmt(s.Init, p),
			Tag:  BuildAsqExpr(s.Tag, p),
			Body: BuildAsqStmt(s.Body, p).(*BlockStmt),
		}
	case *ast.TypeSwitchStmt:
		return &TypeSwitchStmt{
			Ast:    s,
			Init:   BuildAsqStmt(s.Init, p),
			Assign: BuildAsqStmt(s.Assign, p),
			Body:   BuildAsqStmt(s.Body, p).(*BlockStmt),
		}
	default:
		return &DefaultStmt{Ast: s}
	}
}

// BuildAsqDecl converts an ast.Decl to its corresponding asq.Decl
func BuildAsqDecl(decl ast.Decl, p *passOne) Decl {
	if decl == nil {
		return nil
	}

	switch d := decl.(type) {
	case *ast.BadDecl:
		return &BadDecl{Ast: d}
	case *ast.FuncDecl:
		var name *Ident
		if d.Name != nil {
			if expr := BuildAsqExpr(d.Name, p); expr != nil {
				name = expr.(*Ident)
			}
		}
		var funcType *FuncType
		if d.Type != nil && (d.Type.Params != nil && len(d.Type.Params.List) > 0 || d.Type.Results != nil && len(d.Type.Results.List) > 0) {
			if typeNode := BuildAsqNode(d.Type, p); typeNode != nil {
				funcType = typeNode.(*FuncType)
			}
		}
		var body Node
		if d.Body != nil {
			body = BuildAsqNode(d.Body, p)
			if blockStmt, ok := body.(*BlockStmt); ok && len(blockStmt.Ast.List) == 1 {
				if ret, ok := blockStmt.Ast.List[0].(*ast.ReturnStmt); ok {
					body = &ReturnStmt{Ast: ret}
				}
			}
		}
		return &FuncDecl{
			Ast:  d,
			Name: name,
			Type: funcType,
			Body: body,
		}
	case *ast.GenDecl:
		return &GenDecl{
			Ast: d,
			Specs: slicex.Map(d.Specs, func(spec ast.Spec) Node {
				return BuildAsqNode(spec, p)
			}),
		}

	default:
		return &DefaultDecl{Ast: d}
	}
}
