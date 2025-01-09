package metaq

import (
	"go/ast"
)

// convertToTreeSitterQuery converts a Go AST node to a tree-sitter query string
func convertToTreeSitterQuery(node ast.Node) string {
	metaqNode := BuildMetaqNode(node)
	return metaqNode.Convert() + " @x"
}
