package metaq

import (
	"go/ast"
)

// convertToTreeSitterQuery converts a Go AST node to a tree-sitter query string
func convertToTreeSitterQuery(node ast.Node, wildcardIdent map[*ast.Ident]bool) string {
	metaqNode := BuildAsqNode(node, wildcardIdent)
	return metaqNode.Convert() + " @x"
}
