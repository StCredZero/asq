package fs

import (
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildFileTree(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "asq-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file structure
	files := map[string]string{
		"test.go":           "package test\nfunc main() {}\n",
		"asq_query.go":      "package test\nfunc asq_query() {}\n",
		"test1.go":          "package test\nfunc Test1() {}\n",
		"test2.go":          "package test\nfunc Test2() {}\n",
		"dir1/test1.go":     "package test\nfunc Test1() {}\n",
		"dir1/test2.go":     "package test\nfunc Test2() {}\n",
		"asq/skip/test.go":  "package test\nfunc Skip() {}\n",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", filepath.Dir(fullPath), err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Test BuildFileTree
	fset := token.NewFileSet()
	tree, err := BuildFileTree(tmpDir, fset)
	if err != nil {
		t.Fatalf("BuildFileTree failed: %v", err)
	}

	// Verify tree structure
	tests := []struct {
		path     string
		expected bool
		isDir    bool
	}{
		{"test.go", true, false},
		{"asq_query.go", false, false}, // Should be skipped
		{"test1.go", true, false},
		{"test2.go", true, false},
		{"asq/skip/test.go", false, false}, // Should be skipped
		{"dir1", true, true},
		{"asq", false, true}, // Should be skipped
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			components := filepath.SplitList(tt.path)
			current := tree
			found := true

			for i, comp := range components {
				if i == len(components)-1 && !tt.isDir {
					if node, ok := current[comp].(FileNode); ok {
						if !tt.expected {
							t.Errorf("File %s should have been skipped", tt.path)
						}
						if node.Type != "file" {
							t.Errorf("Expected file type for %s, got %s", tt.path, node.Type)
						}
						if tt.path == "test.go" && node.AST == nil {
							t.Error("AST should not be nil for Go file")
						}
					} else {
						found = false
					}
				} else {
					var ok bool
					current, ok = current[comp].(map[string]interface{})
					if !ok {
						found = false
						break
					}
				}
			}

			if found != tt.expected {
				if tt.expected {
					t.Errorf("Expected to find %s in tree", tt.path)
				} else {
					t.Errorf("Expected %s to be skipped", tt.path)
				}
			}
		})
	}
}
