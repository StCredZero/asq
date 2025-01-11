package asq

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
	astFile, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %v", err)
	}

	passOneData := newPassOne(astFile)

	// Find the code block between comments
	var collectingComments bool
	var startPos, endPos token.Pos
	for _, cg := range astFile.Comments {
		for _, c := range cg.List {
			if strings.TrimSpace(c.Text) == "//asq_start" {
				collectingComments = true
				startPos = c.End()
			} else if strings.TrimSpace(c.Text) == "//asq_end" {
				endPos = c.Pos()
				break
			}
			if collectingComments && c.Text == "/***/" {
				passOneData.addInterval(cg.Pos(), cg.End())
			}
		}
	}
	// pad out the last interval to the //asq_end boundary
	passOneData.setLastIntervalEnd(endPos)

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
	return convertToTreeSitterQuery(foundNode, passOneData)
}
