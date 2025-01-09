package metaq

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
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
	var hasWildcard bool
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

	// Get the source code between comments
	sourceBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %v", err)
	}
	source := string(sourceBytes)

	// Find the code between comments in the source
	startOffset := int(fset.Position(startPos).Offset)
	endOffset := int(fset.Position(endPos).Offset)
	if startOffset >= 0 && endOffset > startOffset && endOffset <= len(source) {
		codeBlock := source[startOffset:endOffset]
		hasWildcard = strings.Contains(codeBlock, "/***/")
		if hasWildcard {
			// Store the position of /***/ for processing
			wildcardPos := strings.Index(codeBlock, "/***/") + startOffset
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return true
				}
				// Mark identifiers that appear right after /***/ as wildcards
				if ident, ok := n.(*ast.Ident); ok {
					identPos := int(fset.Position(ident.Pos()).Offset)
					if identPos > wildcardPos && identPos-wildcardPos <= 5 { // 5 is length of /***/
						ident.Name = "wildcarded_" + ident.Name
					}
				}
				return true
			})
		}
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
