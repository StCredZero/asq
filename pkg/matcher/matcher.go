package matcher

import (
	"go/ast"
)

// ASTMatch compares two AST nodes for structural equality.
// This function handles all AST node types including method chains (e.g., e.Inst().Foo())
// through its CallExpr and SelectorExpr cases. The exclusion of specific patterns
// (like asq_end() calls) should be handled during pattern extraction in the query
// package, not in the matcher itself.
func ASTMatch(pattern, target ast.Node) bool {
	// Handle nil cases
	if pattern == nil || target == nil {
		return pattern == target
	}

	switch p := pattern.(type) {
	case *ast.ExprStmt:
		if t, ok := target.(*ast.ExprStmt); ok {
			return ASTMatch(p.X, t.X)
		}
	case *ast.BlockStmt:
		if t, ok := target.(*ast.BlockStmt); ok {
			if len(p.List) != len(t.List) {
				return false
			}
			for i := range p.List {
				if !ASTMatch(p.List[i], t.List[i]) {
					return false
				}
			}
			return true
		}
	case *ast.AssignStmt:
		if t, ok := target.(*ast.AssignStmt); ok {
			if len(p.Lhs) != len(t.Lhs) || len(p.Rhs) != len(t.Rhs) || p.Tok != t.Tok {
				return false
			}
			for i := range p.Lhs {
				if !ASTMatch(p.Lhs[i], t.Lhs[i]) {
					return false
				}
			}
			for i := range p.Rhs {
				if !ASTMatch(p.Rhs[i], t.Rhs[i]) {
					return false
				}
			}
			return true
		}
	case *ast.CallExpr:
		if t, ok := target.(*ast.CallExpr); ok {
			if !ASTMatch(p.Fun, t.Fun) || len(p.Args) != len(t.Args) {
				return false
			}
			for i := range p.Args {
				if !ASTMatch(p.Args[i], t.Args[i]) {
					return false
				}
			}
			return true
		}
	case *ast.Ident:
		if t, ok := target.(*ast.Ident); ok {
			return p.Name == t.Name
		}
	case *ast.SelectorExpr:
		if t, ok := target.(*ast.SelectorExpr); ok {
			return ASTMatch(p.X, t.X) && ASTMatch(p.Sel, t.Sel)
		}
	case *ast.BasicLit:
		if t, ok := target.(*ast.BasicLit); ok {
			return p.Kind == t.Kind && p.Value == t.Value
		}
	case *ast.BinaryExpr:
		if t, ok := target.(*ast.BinaryExpr); ok {
			return p.Op == t.Op && ASTMatch(p.X, t.X) && ASTMatch(p.Y, t.Y)
		}
	case *ast.ReturnStmt:
		if t, ok := target.(*ast.ReturnStmt); ok {
			if len(p.Results) != len(t.Results) {
				return false
			}
			for i := range p.Results {
				if !ASTMatch(p.Results[i], t.Results[i]) {
					return false
				}
			}
			return true
		}
	}
	return false
}
