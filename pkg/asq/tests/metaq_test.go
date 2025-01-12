package cmd

import (
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

			// Compare query output
			if got := strings.TrimSpace(query); got != strings.TrimSpace(tt.expected) {
				t.Errorf("\nExpected query:\n%s\nGot:\n%s", tt.expected, got)
			}

			// Validate query using tree-sitter
			lineNum, matchedCode, err := asq.ValidateTreeSitterQuery(testFile, tt.expected)
			if err != nil {
				t.Errorf("Failed to validate query: %v", err)
				return
			}

			// Find the line number of code between asq_start/asq_end
			lines := strings.Split(tt.code, "\n")
			var targetLine int
			var targetCode string
			for i, line := range lines {
				if strings.TrimSpace(line) == "//asq_start" {
					// The interesting code is in the next line
					if i+1 < len(lines) {
						targetLine = i + 2 // Add 2 because: 1 for 0-based to 1-based, and 1 for the line after asq_start
						targetCode = strings.TrimSpace(lines[i+1])
						break
					}
				}
			}

			// Verify line number matches
			if lineNum != targetLine {
				t.Errorf("Line number mismatch: expected %d, got %d", targetLine, lineNum)
			}

			// Verify matched code is a substring of the target code
			matchedCode = strings.TrimSpace(matchedCode)
			if !strings.Contains(targetCode, matchedCode) {
				t.Errorf("Code mismatch:\nExpected to contain: %s\nGot: %s", matchedCode, targetCode)
			}
		})
	}
}
