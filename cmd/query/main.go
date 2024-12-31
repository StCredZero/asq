package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/stcredzero/asq/pkg/fs"
	"github.com/stcredzero/asq/pkg/query"
)

// For testing
var osExit = os.Exit

// Args holds the command line arguments
type Args struct {
	Find *FindCmd `arg:"subcommand:find" help:"finds code in the named file"`
}

// FindCmd represents the find subcommand
type FindCmd struct {
	Filepath string `arg:"positional" help:"path to the query file"`
}

func main() {
	var args Args
	parser, err := arg.NewParser(arg.Config{
		Program: "asq",
	}, &args)
	if err != nil {
		fmt.Printf("Error creating parser: %v\n", err)
		osExit(1)
	}
	if err := parser.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		osExit(1)
	}
	
	if args.Find == nil {
		parser.WriteHelp(os.Stdout)
		fmt.Println("\nerror: No subcommand specified")
		osExit(1)
	}



	// Create a new token.FileSet for position information
	fset := token.NewFileSet()

	// Extract query pattern from the specified file
	queryBody, err := query.ExtractQueryPattern(fset, args.Find.Filepath)
	if err != nil {
		fmt.Printf("Error extracting query pattern: %v\n", err)
		osExit(1)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		osExit(1)
	}

	// Build file tree from current directory
	rootLevel, err := fs.BuildFileTree(cwd, fset)
	if err != nil {
		fmt.Printf("Error building file tree: %v\n", err)
		osExit(1)
	}

	// Walk through the file tree and find matches
	var findMatches func(map[string]interface{})
	findMatches = func(level map[string]interface{}) {
		for _, v := range level {
			switch node := v.(type) {
			case fs.FileNode:
				if node.AST != nil {
					// Search for matches in this file
					ast.Inspect(node.AST, func(n ast.Node) bool {
						if n != nil && query.MatchPattern(queryBody, n) {
							pos := fset.Position(n.Pos())
							fmt.Printf("%s:%d\n", pos.Filename, pos.Line)
						}
						return true
					})
				}
			case map[string]interface{}:
				findMatches(node)
			}
		}
	}

	findMatches(rootLevel)
	fmt.Println("Done!")
}
