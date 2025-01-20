package asq

import (
	"errors"
	"fmt"
	"github.com/go-enry/go-enry/v2"
	"github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
)

// ConvertToTreeSitterQuery converts a Go AST node to a tree-sitter query string
func ConvertToTreeSitterQuery(node ast.Node, p *QueryContext) (string, error) {
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

var (
	ErrUnsupportedLang = errors.New("unsupported language")
)

// GetTSLanguageFromEnry detects the language of a file using go-enry and returns
// the corresponding tree-sitter language parser. Currently only supports Go.
func GetTSLanguageFromEnry(filename string, contents []byte) (*sitter.Language, error) {
	lang := enry.GetLanguage(filename, contents)
	if lang == "" {
		return nil, errors.New("could not detect language")
	}
	switch lang {
	case "Go":
		return golang.GetLanguage(), nil
	default:
		return nil, ErrUnsupportedLang
	}
}

// Match represents a single tree-sitter query match
type Match struct {
	Row  int
	Col  int
	Code string
}

// MatchGroup represents a group of matches that should be displayed together
type MatchGroup struct {
	FilePath    string
	StartLine   int
	EndLine     int
	Snippet     string
	IsFunction  bool
	FunctionPos token.Pos // Used for sorting function groups
}

// GroupMatchesForCursorDedup takes a list of matches for a file and groups them
// by their containing functions or root-level context. Returns groups sorted by position.
func GroupMatchesForCursorDedup(filePath string, matches []Match) ([]MatchGroup, error) {
	if len(matches) == 0 {
		return nil, nil
	}

	// Read and parse the file
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %v", err)
	}

	// Map to track matches by function
	functionGroups := make(map[*ast.FuncDecl][]Match)
	var rootMatches []Match

	// Find containing functions for all matches
	for _, match := range matches {
		var foundFunc *ast.FuncDecl
		ast.Inspect(astFile, func(n ast.Node) bool {
			if fd, ok := n.(*ast.FuncDecl); ok {
				startPos := fset.Position(fd.Pos())
				endPos := fset.Position(fd.End())
				if match.Row >= startPos.Line && match.Row <= endPos.Line {
					foundFunc = fd
					return false
				}
			}
			return true
		})

		if foundFunc != nil {
			functionGroups[foundFunc] = append(functionGroups[foundFunc], match)
		} else {
			rootMatches = append(rootMatches, match)
		}
	}

	// Create result groups
	var groups []MatchGroup
	lines := strings.Split(string(contents), "\n")

	// Handle root-level matches first
	if len(rootMatches) > 0 {
		// Sort root matches by line number
		sort.Slice(rootMatches, func(i, j int) bool {
			return rootMatches[i].Row < rootMatches[j].Row
		})

		// Find min and max lines
		minLine := rootMatches[0].Row
		maxLine := rootMatches[len(rootMatches)-1].Row

		// Add context lines
		startLine := minLine - 5
		if startLine < 1 {
			startLine = 1
		}
		endLine := maxLine + 5
		if endLine > len(lines) {
			endLine = len(lines)
		}

		// Extract lines, excluding any that are part of functions
		var contextLines []string
		for i := startLine - 1; i < endLine; i++ {
			isInFunction := false
			ast.Inspect(astFile, func(n ast.Node) bool {
				if fd, ok := n.(*ast.FuncDecl); ok {
					startPos := fset.Position(fd.Pos())
					endPos := fset.Position(fd.End())
					if i+1 >= startPos.Line && i+1 <= endPos.Line {
						isInFunction = true
						return false
					}
				}
				return true
			})
			if !isInFunction {
				contextLines = append(contextLines, lines[i])
			}
		}

		if len(contextLines) > 0 {
			// Trim leading and trailing empty lines while preserving internal spacing
			start := 0
			end := len(contextLines)
			
			// Find first non-empty line
			for start < end && strings.TrimSpace(contextLines[start]) == "" {
				start++
			}
			
			// Find last non-empty line
			for end > start && strings.TrimSpace(contextLines[end-1]) == "" {
				end--
			}
			
			if start < end {
				groups = append(groups, MatchGroup{
					FilePath:   filePath,
					StartLine:  startLine + start,
					EndLine:    startLine + end - 1,
					Snippet:    strings.Join(contextLines[start:end], "\n"),
					IsFunction: false,
				})
			}
		}
	}

	// Handle function groups
	for fd := range functionGroups {
		startPos := fset.Position(fd.Pos())
		endPos := fset.Position(fd.End())
		functionLines := lines[startPos.Line-1:endPos.Line]
		groups = append(groups, MatchGroup{
			FilePath:    filePath,
			StartLine:   startPos.Line,
			EndLine:     endPos.Line,
			Snippet:     strings.Join(functionLines, "\n"),
			IsFunction:  true,
			FunctionPos: fd.Pos(),
		})
	}

	// Sort groups by position (functions first, then root)
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].IsFunction && groups[j].IsFunction {
			return groups[i].FunctionPos < groups[j].FunctionPos
		}
		if groups[i].IsFunction {
			return true
		}
		if groups[j].IsFunction {
			return false
		}
		return groups[i].StartLine < groups[j].StartLine
	})

	return groups, nil
}

// ValidateTreeSitterQuery executes a tree-sitter query directly on the given file
// returns all matches with their line numbers, column numbers, and matched code
func ValidateTreeSitterQuery(file, query string) ([]Match, error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	lang, err := GetTSLanguageFromEnry(file, contents)
	if err != nil {
		return nil, fmt.Errorf("failed to get language: %v", err)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree := parser.Parse(nil, contents)
	root := tree.RootNode()

	q, err := sitter.NewQuery([]byte(query), lang)
	if err != nil {
		return nil, fmt.Errorf("invalid query: %v", err)
	}
	defer q.Close()

	qc := sitter.NewQueryCursor()
	defer qc.Close()
	qc.Exec(q, root)

	var matches []Match
	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}
		for _, c := range match.Captures {
			if q.CaptureNameForId(c.Index) == "x" {
				row := int(c.Node.StartPoint().Row) + 1
				col := int(c.Node.StartPoint().Column)
				// Get the complete node content
				nodeContent := string(contents[c.Node.StartByte():c.Node.EndByte()])

				// Get the line containing the node
				lineStart := 0
				for i := 0; i < int(c.Node.StartPoint().Row); i++ {
					lineStart = strings.IndexByte(string(contents[lineStart:]), '\n') + lineStart + 1
				}
				lineEnd := strings.IndexByte(string(contents[lineStart:]), '\n')
				if lineEnd == -1 {
					lineEnd = len(contents)
				} else {
					lineEnd += lineStart
				}

				// Extract the complete line for context
				fullLine := string(contents[lineStart:lineEnd])

				var finalCode string
				if strings.Contains(fullLine, "/***/") {
					// For lines with wildcards, use the complete line
					finalCode = strings.TrimSpace(fullLine)
				} else if strings.Contains(nodeContent, "\n") {
					// For multiline nodes (like function declarations), use the node content
					lines := strings.Split(nodeContent, "\n")
					for i := range lines {
						lines[i] = strings.TrimRight(lines[i], " \t\r\n")
					}
					finalCode = strings.Join(lines, "\n")
				} else {
					// For single-line nodes without wildcards, use the node content
					finalCode = strings.TrimSpace(nodeContent)
				}

				matches = append(matches, Match{
					Row:  row,
					Col:  col,
					Code: finalCode,
				})
			}
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no match found for capture @x")
	}

	return matches, nil
}

// GetSnippetForMatch returns the code snippet for a given match, including context.
// If the match is within a function, returns the entire function.
// Otherwise, returns 5 lines before and after the match.
func GetSnippetForMatch(filePath string, match Match) (string, error) {
	// Read the file contents
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Parse the file for AST analysis
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %v", err)
	}

	// Find if the match is within a function
	var containingFunc *ast.FuncDecl
	ast.Inspect(astFile, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			// Get the position information
			startPos := fset.Position(fd.Pos())
			endPos := fset.Position(fd.End())
			
			// Check if match.Row is within this function's lines
			if match.Row >= startPos.Line && match.Row <= endPos.Line {
				containingFunc = fd
				return false // Stop traversal
			}
		}
		return true
	})

	// Split content into lines for processing
	lines := strings.Split(string(contents), "\n")

	if containingFunc != nil {
		// Get the entire function text
		startPos := fset.Position(containingFunc.Pos())
		endPos := fset.Position(containingFunc.End())
		
		// Extract function lines (convert to 0-based index)
		functionLines := lines[startPos.Line-1:endPos.Line]
		return strings.Join(functionLines, "\n"), nil
	}

	// If not in a function, get 5 lines before and after
	startLine := match.Row - 5
	if startLine < 1 {
		startLine = 1
	}
	endLine := match.Row + 5
	if endLine > len(lines) {
		endLine = len(lines)
	}

	// Extract the lines (convert to 0-based index)
	contextLines := lines[startLine-1:endLine]
	return strings.Join(contextLines, "\n"), nil
}
