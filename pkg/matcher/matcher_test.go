package matcher

import (
	"go/ast"
	"go/parser"
	"testing"
)

func TestASTMatch(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		target   string
		expected bool
	}{
		{
			name:     "Simple method chain",
			pattern:  `e.Inst().Foo()`,
			target:   `e.Inst().Foo()`,
			expected: true,
		},
		{
			name:     "Different method chain",
			pattern:  `e.Inst().Foo()`,
			target:   `e.Inst().Bar()`,
			expected: false,
		},
		{
			name:     "Different receiver",
			pattern:  `e.Inst().Foo()`,
			target:   `x.Inst().Foo()`,
			expected: false,
		},
		{
			name:     "Simple assignment",
			pattern:  `42`,
			target:   `42`,
			expected: true,
		},
		{
			name:     "Different assignment",
			pattern:  `42`,
			target:   `43`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse pattern
			pattern, err := parser.ParseExpr(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to parse pattern: %v", err)
			}
			patternStmt := &ast.ExprStmt{X: pattern}

			// Parse target
			target, err := parser.ParseExpr(tt.target)
			if err != nil {
				t.Fatalf("Failed to parse target: %v", err)
			}
			targetStmt := &ast.ExprStmt{X: target}

			// Test matching
			result := ASTMatch(patternStmt, targetStmt)
			if result != tt.expected {
				t.Errorf("ASTMatch(%q, %q) = %v, want %v",
					tt.pattern, tt.target, result, tt.expected)
			}
		})
	}
}
