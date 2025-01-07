package metaq

import (
	"fmt"
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
		return convertCallExpr(n)
	case *ast.SelectorExpr:
		return convertSelectorExpr(n)
	case *ast.Ident:
		return "(identifier)"
	default:
		return fmt.Sprintf("(%T)", n)
	}
}

func convertCallExpr(call *ast.CallExpr) string {
	var sb strings.Builder
	sb.WriteString("(call_expression function: ")
	sb.WriteString(convertNode(call.Fun))
	sb.WriteString(" arguments: (argument_list))")
	return sb.String()
}

func convertSelectorExpr(sel *ast.SelectorExpr) string {
	var sb strings.Builder
	sb.WriteString("(selector_expression operand: ")
	sb.WriteString(convertNode(sel.X))
	sb.WriteString(" field: (field_identifier))")
	return sb.String()
}
