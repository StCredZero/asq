package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/StCredZero/asq/pkg/asq"
	"github.com/alexflint/go-arg"
)

type TreeSitterCmd struct {
	File string `arg:"positional,required" help:"path to go file "`
}

type QueryCmd struct {
	File string `arg:"positional,required" help:"path to asq query file"`
}

type CLI struct {
	TreeSitter *TreeSitterCmd `arg:"subcommand:tree-sitter" help:"Generate a tree-sitter query from a Go file"`
	Query      *QueryCmd      `arg:"subcommand:query" help:"Search for matches using the tree-sitter query from a Go file"`
}

func main() {
	var cli CLI
	arg.MustParse(&cli)

	switch {
	case cli.TreeSitter != nil:
		// Generate tree-sitter query from file
		query, err := asq.ExtractTreeSitterQuery(cli.TreeSitter.File)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(query)

	case cli.Query != nil:
		// Generate tree-sitter query from file
		query, err := asq.ExtractTreeSitterQuery(cli.Query.File)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating query: %v\n", err)
			os.Exit(1)
		}

		// Walk through current directory recursively
		err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Skip non-Go files
			if !info.IsDir() && filepath.Ext(path) == ".go" {
				// Validate query against current file
				matches, err := asq.ValidateTreeSitterQuery(path, query)
				if err == nil {
					for _, match := range matches {
						fmt.Printf("//asq_match %s:%d:%d\n%s\n", path, match.Row, match.Col, match.Code)
					}
				}
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
			os.Exit(1)
		}
	}
}
