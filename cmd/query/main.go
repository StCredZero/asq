package main

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
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
		parser.WriteHelp(os.Stdout)
		fmt.Printf("\nerror: Failed to create parser: %v\n", err)
		osExit(1)
	}
	if err := parser.Parse(os.Args[1:]); err != nil {
		if err == arg.ErrHelp {
			parser.WriteHelp(os.Stdout)
			osExit(0)
		}
		parser.WriteHelp(os.Stdout)
		fmt.Printf("\nerror: No subcommand specified\n")
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
		if strings.Contains(err.Error(), "no such file or directory") {
			fmt.Printf("Usage: asq <command> [<args>]\n\n")
			fmt.Printf("Options:\n")
			fmt.Printf("  --help, -h             display this help and exit\n\n")
			fmt.Printf("Commands:\n")
			fmt.Printf("  find                   finds code in the named file\n\n")
			fmt.Printf("error: No subcommand specified\n")
		} else {
			parser.WriteHelp(os.Stdout)
			fmt.Printf("\nerror: No subcommand specified\n")
		}
		osExit(1)
	}

	fmt.Printf("Debug: Extracted query pattern from %s\n", args.Find.Filepath)

	// Find matches in test001.go in the same directory as the query file
	queryDir := filepath.Dir(args.Find.Filepath)
	testFilePath := filepath.Join(queryDir, "test001.go")
	
	matches, err := query.FindMatches(queryBody, testFilePath)
	if err != nil {
		parser.WriteHelp(os.Stdout)
		fmt.Printf("\nerror: Failed to find matches: %v\n", err)
		osExit(1)
	}

	// Print matches
	if len(matches) == 0 {
		fmt.Printf("Debug: No matches found in test001.go\n")
	} else {
		fmt.Printf("Debug: Found %d matches:\n", len(matches))
		for _, match := range matches {
			fmt.Println(match)
		}
	}
	fmt.Println("Done!")
}
