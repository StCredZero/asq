package asq

import (
	"go/ast"
	"strings"
)

// convertToTreeSitterQuery converts a Go AST node to a tree-sitter query string
func convertToTreeSitterQuery(node ast.Node, p *passOne) (string, error) {
	var sb strings.Builder
	metaqNode := BuildAsqNode(node, p)
	if err := metaqNode.WriteTreeSitterQuery(&sb); err != nil {
		return "", err
	}
	if _, err := sb.WriteString(" @x"); err != nil {
		return "", err
	}
	return sb.String(), nil
}