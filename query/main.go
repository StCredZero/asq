// Implementation Plan for parse utility
/*
Package Structure:
- main package in asq/parse/main.go
- will use:
  - go/ast: for parsing Go source files
  - go/parser: for parsing source into AST
  - go/token: for position information
  - path/filepath: for walking directory tree
  - strings: for path manipulation
  - fmt: for output

Data Structure Design:
type FileNode struct {
    AST  *ast.File        // AST for Go files
    Path string           // Full path to file
    Type string           // "file" or "dir"
}

Main Implementation Steps:
1. Create FileSystem walker
   - Use filepath.Walk
   - Skip directories named "asq"
   - Build nested map structure

2. Parse Go files
   - Use parser.ParseFile for .go files
   - Store AST in map

3. Store in rootLevel variable
   - Create nested map[string]interface{}
   - Keys are path elements
   - Values are either FileNode or nested maps

4. Add breakpoint
   - fmt.Println("Done!") at end

Example structure of rootLevel:
rootLevel = {
    "src": {
        "package": {
            "file.go": FileNode{
                AST: *ast.File,
                Path: "src/package/file.go",
                Type: "file"
            }
        }
    }
}
*/

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	arg "github.com/alexflint/go-arg"
)

// Args defines the command line arguments
type Args struct {
	Find *FindCmd `arg:"subcommand:find" help:"finds code in the named file"`
}

// FindCmd represents the find subcommand
type FindCmd struct {
	Filepath string `arg:"positional,required" help:"path to the golang source file containing the query"`
}

// astMatch compares two AST nodes for structural equality
func astMatch(pattern, target ast.Node) bool {
	// Handle nil cases
	if pattern == nil || target == nil {
		return pattern == target
	}

	switch p := pattern.(type) {
	case *ast.ExprStmt:
		if t, ok := target.(*ast.ExprStmt); ok {
			return astMatch(p.X, t.X)
		}
	case *ast.BlockStmt:
		if t, ok := target.(*ast.BlockStmt); ok {
			if len(p.List) != len(t.List) {
				return false
			}
			for i := range p.List {
				if !astMatch(p.List[i], t.List[i]) {
					return false
				}
			}
			return true
		}
	case *ast.AssignStmt:
		if t, ok := target.(*ast.AssignStmt); ok {
			if len(p.Lhs) != len(t.Lhs) || len(p.Rhs) != len(t.Rhs) || p.Tok != t.Tok {
				return false
			}
			for i := range p.Lhs {
				if !astMatch(p.Lhs[i], t.Lhs[i]) {
					return false
				}
			}
			for i := range p.Rhs {
				if !astMatch(p.Rhs[i], t.Rhs[i]) {
					return false
				}
			}
			return true
		}
	case *ast.CallExpr:
		if t, ok := target.(*ast.CallExpr); ok {
			if !astMatch(p.Fun, t.Fun) || len(p.Args) != len(t.Args) {
				return false
			}
			for i := range p.Args {
				if !astMatch(p.Args[i], t.Args[i]) {
					return false
				}
			}
			return true
		}
	case *ast.Ident:
		if t, ok := target.(*ast.Ident); ok {
			return p.Name == t.Name
		}
	case *ast.SelectorExpr:
		if t, ok := target.(*ast.SelectorExpr); ok {
			return astMatch(p.X, t.X) && astMatch(p.Sel, t.Sel)
		}
	case *ast.BasicLit:
		if t, ok := target.(*ast.BasicLit); ok {
			return p.Kind == t.Kind && p.Value == t.Value
		}
	case *ast.BinaryExpr:
		if t, ok := target.(*ast.BinaryExpr); ok {
			return p.Op == t.Op && astMatch(p.X, t.X) && astMatch(p.Y, t.Y)
		}
	case *ast.ReturnStmt:
		if t, ok := target.(*ast.ReturnStmt); ok {
			if len(p.Results) != len(t.Results) {
				return false
			}
			for i := range p.Results {
				if !astMatch(p.Results[i], t.Results[i]) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// FileNode represents a file in the filesystem
type FileNode struct {
	AST  *ast.File
	Path string
	Type string
}

// buildFileTree builds a nested map representing the file hierarchy
func buildFileTree(root string, fset *token.FileSet) (map[string]interface{}, error) {
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
			// Skip files with asq_ prefix
			if strings.HasPrefix(lastComp, "asq_") {
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

func main() {
	var args Args
	arg.MustParse(&args)

	if args.Find != nil {
		// Create a single FileSet for all operations
		fset := token.NewFileSet()

		// Parse the query file to extract the asq_query function body
		queryFile, err := parser.ParseFile(fset, args.Find.Filepath, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing query file: %v\n", err)
			os.Exit(1)
		}

		// Find the asq_query function and extract its body
		var queryBody *ast.BlockStmt
		ast.Inspect(queryFile, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "asq_query" {
				queryBody = fn.Body
				// Filter out asq_end() call from queryBody
				for i, stmt := range queryBody.List {
					if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
						if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
							if funIdent, ok := callExpr.Fun.(*ast.Ident); ok && funIdent.Name == "asq_end" {
								queryBody.List = append(queryBody.List[:i], queryBody.List[i+1:]...)
								break
							}
						}
					}
				}
				return false
			}
			return true
		})

		if queryBody == nil {
			fmt.Printf("Error: asq_query function not found in %s\n", args.Find.Filepath)
			os.Exit(1)
		}

		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		// Build the file tree using the same FileSet
		rootLevel, err := buildFileTree(cwd, fset)
		if err != nil {
			fmt.Printf("Error building file tree: %v\n", err)
			os.Exit(1)
		}

		// Search through all Go files in the tree for matching AST patterns
		matchCount := 0

		// Helper function to recursively search through the file tree
		var searchTree func(m map[string]interface{})
		searchTree = func(m map[string]interface{}) {
			for _, v := range m {
				switch node := v.(type) {
				case map[string]interface{}:
					searchTree(node)
				case FileNode:
					// Skip files with asq_ prefix and the query file itself
					if strings.HasPrefix(filepath.Base(node.Path), "asq_") || node.Path == args.Find.Filepath {
						continue
					}

					if node.AST != nil {
						// Compare each statement in queryBody against all statements in the file
						for _, queryStmt := range queryBody.List {
							ast.Inspect(node.AST, func(n ast.Node) bool {
								if stmt, ok := n.(*ast.ExprStmt); ok {
									if astMatch(queryStmt, stmt) {
										pos := fset.Position(stmt.Pos())
										fmt.Printf("%s:%d\n", pos.Filename, pos.Line)
										matchCount++
									}
								}
								return true
							})
						}
					}
				}
			}
		}

		// Start the recursive search
		searchTree(rootLevel)

		if matchCount == 0 {
			fmt.Printf("No matches found\n")
		}
	}

	fmt.Println("Done!")
}
