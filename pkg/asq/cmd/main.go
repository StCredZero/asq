package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/StCredZero/asq/pkg/asq"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <go-file>\n", os.Args[0])
		os.Exit(1)
	}

	filePath := flag.Arg(0)
	query, err := asq.ExtractTreeSitterQuery(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(query)
}
