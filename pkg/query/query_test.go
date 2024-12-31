package query

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractQueryPattern(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "asq-query-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		content  string
		wantErr  bool
		errMsg   string
		stmtLen  int // Expected number of statements after extraction
	}{
		{
			name: "Valid query with asq_end",
			content: `package test
func asq_query() {
	e.Inst().Foo()
	asq_end()
}`,
			wantErr: false,
			stmtLen: 1, // Only e.Inst().Foo() should remain
		},
		{
			name: "Valid query without asq_end",
			content: `package test
func asq_query() {
	e.Inst().Foo()
}`,
			wantErr: false,
			stmtLen: 1,
		},
		{
			name: "Multiple statements with asq_end",
			content: `package test
func asq_query() {
	x := 42
	e.Inst().Foo()
	asq_end()
}`,
			wantErr: false,
			stmtLen: 2, // x := 42 and e.Inst().Foo() should remain
		},
		{
			name: "Missing asq_query function",
			content: `package test
func other() {
	e.Inst().Foo()
}`,
			wantErr: true,
			errMsg:  "asq_query function not found",
		},
		{
			name:    "Invalid Go file",
			content: "invalid go syntax",
			wantErr: true,
			errMsg:  "error parsing query file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			filename := filepath.Join(tmpDir, "test.go")
			err := os.WriteFile(filename, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Test ExtractQueryPattern
			fset := token.NewFileSet()
			got, err := ExtractQueryPattern(fset, filename)

			// Check error cases
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %v", tt.errMsg, err)
				}
				return
			}

			// Check success cases
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if got == nil {
				t.Fatal("Expected non-nil BlockStmt")
			}

			if len(got.List) != tt.stmtLen {
				t.Errorf("Expected %d statements, got %d", tt.stmtLen, len(got.List))
			}

			// Verify asq_end() is removed
			for _, stmt := range got.List {
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
						if ident, ok := callExpr.Fun.(*ast.Ident); ok {
							if ident.Name == "asq_end" {
								t.Error("Found asq_end() call that should have been removed")
							}
						}
					}
				}
			}
		})
	}
}

// contains checks if substr is contained in s
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}
