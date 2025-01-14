package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	"github.com/StCredZero/asq/pkg/asq"
)

type CLI struct {
	TreeSitter struct {
		File string `arg:"positional,required" help:"Path to Go file with //asq_start and //asq_end tags"`
	} `arg:"subcommand:tree-sitter" help:"Generate a tree-sitter query from a Go file"`

	Query struct {
		File string `arg:"positional,required" help:"Path to Go file with //asq_start and //asq_end tags"`
	} `arg:"subcommand:query" help:"Search for matches using the tree-sitter query from a Go file"`
}

func main() {
	var cli CLI
	arg.MustParse(&cli)

	switch {
	case cli.TreeSitter.File != "":
		// Generate tree-sitter query from file
		query, err := asq.ExtractTreeSitterQuery(cli.TreeSitter.File)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(query)

	case cli.Query.File != "":
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
				if row, matchedCode, err := asq.ValidateTreeSitterQuery(path, query); err == nil {
					fmt.Printf("//asq_match %s:%d:0\n%s\n", path, row, matchedCode)
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
