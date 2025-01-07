package metaq

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// ExtractTreeSitterQuery parses a Go file and extracts the code between //asq_start and //asq_end
// comments, then converts it to a tree-sitter query.
func ExtractTreeSitterQuery(filePath string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %v", err)
	}

	// Find the code block between comments
	var startPos, endPos token.Pos
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if strings.TrimSpace(c.Text) == "//asq_start" {
				startPos = c.End()
			} else if strings.TrimSpace(c.Text) == "//asq_end" {
				endPos = c.Pos()
				break
			}
		}
	}

	if !startPos.IsValid() || !endPos.IsValid() {
		return "", fmt.Errorf("could not find //asq_start and //asq_end comments")
	}

	// Extract the AST nodes between the comments
	var exprNode ast.Expr
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		if n.Pos() >= startPos && n.End() <= endPos {
			if stmt, ok := n.(*ast.ExprStmt); ok {
				exprNode = stmt.X
				return false
			}
		}
		return true
	})

	if exprNode == nil {
		return "", fmt.Errorf("no expression found between comments")
	}

	// Convert to tree-sitter query
	return convertToTreeSitterQuery(exprNode), nil
}
