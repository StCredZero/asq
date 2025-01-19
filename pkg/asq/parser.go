package asq

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// ExtractTreeSitterQuery parses a Go file and extracts the code between //asq_start and //asq_end
// comments, then converts it to a tree-sitter query.
func ExtractTreeSitterQuery(filePath string) (string, error) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %v", err)
	}

	queryContext, startPos, endPos := NewQueryContext(astFile)
	if !startPos.IsValid() || !endPos.IsValid() {
		return "", fmt.Errorf("could not find //asq_start and //asq_end comments")
	}

	// Extract the AST nodes between the comments
	var foundNode ast.Node
	ast.Inspect(astFile, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		if n.Pos() >= startPos && n.End() <= endPos {
			switch node := n.(type) {
			case *ast.ExprStmt:
				foundNode = node.X
				return false
			case *ast.ReturnStmt, *ast.FuncDecl:
				foundNode = node
				return false
			}
		}
		return true
	})

	if foundNode == nil {
		return "", fmt.Errorf("no node found between comments")
	}

	// Convert to tree-sitter query
	return ConvertToTreeSitterQuery(foundNode, queryContext)
}
