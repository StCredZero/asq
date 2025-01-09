package metaq

import (
	"go/ast"
	"strings"
)

// convertToTreeSitterQuery converts a Go AST node to a tree-sitter query string
func convertToTreeSitterQuery(node ast.Node) string {
	return convertNode(node) + " @x"
}

func convertNode(node ast.Node) string {
	if node == nil {
		return ""
	}

	switch n := node.(type) {
	case *ast.CallExpr:
		return (&CallExpr{Call: n}).Convert()
	case *ast.SelectorExpr:
		return (&SelectorExpr{Sel: n}).Convert()
	case *ast.Ident:
		ident := &Ident{Id: n}
		if strings.HasPrefix(n.Name, "wildcarded_") {
			ident.Wildcard = true
			ident.Id.Name = strings.TrimPrefix(n.Name, "wildcarded_")
		}
		return ident.Convert()
	default:
		return (&DefaultNode{Node: n}).Convert()
	}
}
