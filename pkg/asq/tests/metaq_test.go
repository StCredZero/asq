package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StCredZero/asq/pkg/asq"
	sitter "github.com/smacker/go-tree-sitter"
)

// runTreeSitterValidation executes a tree-sitter query directly on the given file
// returns the line number and matched code, or error if validation fails
func runTreeSitterValidation(file, query string) (int, string, error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read file: %v", err)
	}

	lang, err := asq.GetTSLanguageFromEnry(file, contents)
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

	// Debug: Print the tree structure
	var printNode func(node *sitter.Node, level int)
	printNode = func(node *sitter.Node, level int) {
		indent := strings.Repeat("  ", level)
		fmt.Printf("%s%s\n", indent, node.Type())
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			printNode(child, level+1)
		}
	}
	fmt.Println("Tree structure:")
	printNode(root, 0)

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
				return row, strings.TrimSpace(code), nil
			}
		}
	}
	return 0, "", fmt.Errorf("no match found for capture @x")
}

func TestMetaqQuery(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "exact_match",
			code: `package example1
type Thingy1 struct{}
func (t Thingy1) Inst() Thingy1 { return t }
func (t Thingy1) Foo() bool { return true }
var e = new(Thingy1)
func asq_query2() {
	//asq_start
	e.Inst().Foo()
	//asq_end
}`,
			expected: `(call_expression function: (selector_expression operand: (call_expression function: (selector_expression operand: (identifier) (#eq? _ "e") field: (field_identifier) (#eq? _ "Inst")) arguments: (argument_list)) field: (field_identifier) (#eq? _ "Foo")) arguments: (argument_list)) @x`,
		},
		{
			name: "wildcard_match",
			code: `package example1
type Thingy1 struct{}
func (t Thingy1) Inst() Thingy1 { return t }
func (t Thingy1) Foo() bool { return true }
var e = new(Thingy1)
func asq_query2() {
	//asq_start
	/***/e.Inst().Foo()
	//asq_end
}`,
			expected: `(call_expression function: (selector_expression operand: (call_expression function: (selector_expression operand: (identifier) field: (field_identifier) (#eq? _ "Inst")) arguments: (argument_list)) field: (field_identifier) (#eq? _ "Foo")) arguments: (argument_list)) @x`,
		},
		{
			name: "exact_match_with_different_receiver",
			code: `package example1
type Thingy1 struct{}
func (t Thingy1) Inst() Thingy1 { return t }
func (t Thingy1) Foo() bool { return true }
var x = new(Thingy1)
func asq_query2() {
	//asq_start
	x.Inst().Foo()
	//asq_end
}`,
			expected: `(call_expression function: (selector_expression operand: (call_expression function: (selector_expression operand: (identifier) (#eq? _ "x") field: (field_identifier) (#eq? _ "Inst")) arguments: (argument_list)) field: (field_identifier) (#eq? _ "Foo")) arguments: (argument_list)) @x`,
		},
		{
			name: "negative_test_different_method",
			code: `package example1
type Thingy1 struct{}
func (t Thingy1) Inst2() Thingy1 { return t }
func (t Thingy1) Foo() bool { return true }
var e = new(Thingy1)
func asq_query2() {
	//asq_start
	e.Inst2().Foo()
	//asq_end
}`,
			expected: `(call_expression function: (selector_expression operand: (call_expression function: (selector_expression operand: (identifier) (#eq? _ "e") field: (field_identifier) (#eq? _ "Inst2")) arguments: (argument_list)) field: (field_identifier) (#eq? _ "Foo")) arguments: (argument_list)) @x`,
		},
		{
			name: "return_stmt_no_results",
			code: `package example1
func example() {
	//asq_start
	return
	//asq_end
}`,
			expected: `(return_statement) @x`,
		},
		{
			name: "return_stmt_with_result",
			code: `package example1
func example() bool {
	//asq_start
	return true
	//asq_end
}`,
			expected: `(return_statement (expression_list (true))) @x`,
		},
		{
			name: "function_declaration",
			code: `package example1
//asq_start
func Example() {
	return
}
//asq_end`,
			expected: `(function_declaration name: (identifier) (#eq? _ "Example") body: (block (return_statement))) @x`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "asq-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create test file
			testFile := filepath.Join(tmpDir, "test.go")
			err = os.WriteFile(testFile, []byte(tt.code), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Run asq
			query, err := asq.ExtractTreeSitterQuery(testFile)
			if err != nil {
				t.Fatalf("Failed to extract query: %v", err)
			}

			// Compare query output
			if got := strings.TrimSpace(query); got != strings.TrimSpace(tt.expected) {
				t.Errorf("\nExpected query:\n%s\nGot:\n%s", tt.expected, got)
			}

			// Validate query using tree-sitter
			lineNum, matchedCode, err := runTreeSitterValidation(testFile, tt.expected)
			if err != nil {
				t.Errorf("Failed to validate query: %v", err)
				return
			}

			// Find the code between asq_start/asq_end
			lines := strings.Split(tt.code, "\n")
			var targetLine int
			var targetCode []string
			inTarget := false
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "//asq_start" {
					targetLine = i + 2 // Add 2 because: 1 for 0-based to 1-based, and 1 for the line after asq_start
					inTarget = true
					continue
				}
				if trimmed == "//asq_end" {
					inTarget = false
					continue
				}
				if inTarget {
					targetCode = append(targetCode, line)
				}
			}

			// Join all target code lines
			fullTargetCode := strings.Join(targetCode, "\n")

			// Verify line number matches
			if lineNum != targetLine {
				t.Errorf("Line number mismatch: expected %d, got %d", targetLine, lineNum)
			}

			// Verify matched code is a substring of the target code
			matchedCode = strings.TrimSpace(matchedCode)
			fullTargetCode = strings.TrimSpace(fullTargetCode)
			if !strings.Contains(fullTargetCode, matchedCode) {
				t.Errorf("Code mismatch:\nExpected to contain:\n%s\nGot:\n%s", matchedCode, fullTargetCode)
			}
		})
	}
}
