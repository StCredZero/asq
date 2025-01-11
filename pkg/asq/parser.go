package asq

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
	startOffset := fset.Position(startPos).Offset
	endOffset := fset.Position(endPos).Offset
	p := newPassOne(fset)
	
	if startOffset >= 0 && endOffset > startOffset && endOffset <= len(source) {
		codeBlock := source[startOffset:endOffset]
		
		// Find all /***/ tags and create intervals
		lines := strings.Split(codeBlock, "\n")
		baseOffset := startOffset
		
		for lineNum, line := range lines {
			pos := 0
			for {
				idx := strings.Index(line[pos:], "/***/")
				if idx == -1 {
					break
				}
				
				// Calculate absolute position of this /***/ tag
				tagStart := baseOffset + pos + idx
				
				// Find end of interval (next /***/ or end of line)
				nextTagIdx := strings.Index(line[pos+idx+5:], "/***/")
				intervalEnd := 0
				if nextTagIdx == -1 {
					// No more tags on this line, interval ends at end of line
					intervalEnd = baseOffset + len(line)
				} else {
					// Interval ends at start of next tag
					intervalEnd = tagStart + 5 + nextTagIdx
				}
				
				// Add the interval
				p.addInterval(fset.Position(startPos).Line+lineNum, tagStart, intervalEnd)
				
				// Move position past this tag
				pos += idx + 5
			}
			
			// Move to next line
			baseOffset += len(line) + 1 // +1 for newline
		}
	}

	// Extract the AST nodes between the comments
	var foundNode ast.Node
	ast.Inspect(file, func(n ast.Node) bool {
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
	return convertToTreeSitterQuery(foundNode, p)
}
