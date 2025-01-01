package fs

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// FileNode represents a file in the filesystem
type FileNode struct {
	AST  *ast.File
	Path string
	Type string
}

// BuildFileTree builds a nested map representing the file hierarchy
func BuildFileTree(root string, fset *token.FileSet) (map[string]interface{}, error) {
	rootLevel := make(map[string]interface{})

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip asq directories
		if info.IsDir() && info.Name() == "asq" {
			return filepath.SkipDir
		}

		// Get relative path from root
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Split path into components
		components := strings.Split(relPath, string(os.PathSeparator))

		// Navigate/create the nested map structure
		current := rootLevel
		for _, comp := range components[:len(components)-1] {
			if _, exists := current[comp]; !exists {
				current[comp] = make(map[string]interface{})
			}
			current = current[comp].(map[string]interface{})
		}

		// Create FileNode for the actual file
		lastComp := components[len(components)-1]
		if info.IsDir() {
			current[lastComp] = make(map[string]interface{})
		} else {
			// Skip files with asq_ or lanq_ prefix
			if strings.HasPrefix(lastComp, "lanq_") || strings.HasPrefix(lastComp, "asq_") {
				return nil
			}

			node := FileNode{
				Path: path,
				Type: "file",
			}

			// Parse Go files
			if strings.HasSuffix(path, ".go") {
				astFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
				if err != nil {
					// Log error but continue processing other files
					fmt.Printf("Error parsing %s: %v\n", path, err)
				} else {
					node.AST = astFile
				}
			}

			current[lastComp] = node
		}

		return nil
	})

	return rootLevel, err
}
