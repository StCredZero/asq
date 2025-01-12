package asq

import (
	"errors"
	"fmt"
	"github.com/go-enry/go-enry/v2"
	"github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"go/ast"
	"os"
	"strings"
)

// ConvertToTreeSitterQuery converts a Go AST node to a tree-sitter query string
func ConvertToTreeSitterQuery(node ast.Node, p *passOne) (string, error) {
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

// ValidateTreeSitterQuery executes a tree-sitter query directly on the given file
// returns the line number and matched code, or error if validation fails
func ValidateTreeSitterQuery(file, query string) (int, string, error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read file: %v", err)
	}

	lang, err := GetTSLanguageFromEnry(file, contents)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get language: %v", err)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree := parser.Parse(nil, contents)
	root := tree.RootNode()

	q, err := sitter.NewQuery([]byte(query), lang)
	if err != nil {
		return 0, "", fmt.Errorf("invalid query: %v", err)
	}
	defer q.Close()

	qc := sitter.NewQueryCursor()
	defer qc.Close()
	qc.Exec(q, root)

	// Only retrieve the first relevant capture with @x
	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}
		for _, c := range match.Captures {
			if q.CaptureNameForId(c.Index) == "x" {
				row := int(c.Node.StartPoint().Row) + 1
				code := string(contents[c.Node.StartByte():c.Node.EndByte()])
				lines := strings.Split(code, "\n")
				if len(lines) >= 1 {
					return row, strings.TrimSpace(lines[0]), nil
				}
				return row, "", nil
			}
		}
	}
	return 0, "", fmt.Errorf("no match found for capture @x")
}
