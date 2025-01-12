package cmd

import (
	"fmt"
	sitter "github.com/smacker/go-tree-sitter"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StCredZero/asq/pkg/asq"
)

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
			expected: `(call_expression function: (selector_expression operand: (call_expression function: (selector_expression operand: (identifier) @name (#eq? @name "e") field: (field_identifier) @field (#eq? @field "Inst")) arguments: (argument_list)) field: (field_identifier) @field (#eq? @field "Foo")) arguments: (argument_list)) @x`,
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
			expected: `(call_expression function: (selector_expression operand: (call_expression function: (selector_expression operand: (identifier) field: (field_identifier) @field (#eq? @field "Inst")) arguments: (argument_list)) field: (field_identifier) @field (#eq? @field "Foo")) arguments: (argument_list)) @x`,
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
			expected: `(call_expression function: (selector_expression operand: (call_expression function: (selector_expression operand: (identifier) @name (#eq? @name "x") field: (field_identifier) @field (#eq? @field "Inst")) arguments: (argument_list)) field: (field_identifier) @field (#eq? @field "Foo")) arguments: (argument_list)) @x`,
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
			expected: `(call_expression function: (selector_expression operand: (call_expression function: (selector_expression operand: (identifier) @name (#eq? @name "e") field: (field_identifier) @field (#eq? @field "Inst2")) arguments: (argument_list)) field: (field_identifier) @field (#eq? @field "Foo")) arguments: (argument_list)) @x`,
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
			expected: `(return_statement values: (expression_list (identifier) @value (#eq? @value "true"))) @x`,
		},
		{
			name: "function_declaration",
			code: `package example1
//asq_start
func Example() {
	return
}
//asq_end`,
			expected: `(function_declaration name: (identifier) @name (#eq? @name "Example") body: (block (return_statement))) @x`,
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

			// Compare output
			if got := strings.TrimSpace(query); got != strings.TrimSpace(tt.expected) {
				t.Errorf("\nExpected:\n%s\nGot:\n%s", tt.expected, got)
			}
		})
	}
}

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
