package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/stcredzero/asq/pkg/metaq"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <go-file>\n", os.Args[0])
		os.Exit(1)
	}

	filePath := flag.Arg(0)
	query, err := metaq.ExtractTreeSitterQuery(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(query)
}
