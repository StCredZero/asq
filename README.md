# asq - Active Semantic Query Tool

A tool for semantic code querying using tree-sitter patterns. Extract and match code patterns in Go files using semantic queries.

## Installation

### Using Go

```bash
go install github.com/StCredZero/asq/cmd/asq@latest
```

### Using Homebrew

*Coming soon with first release*

After the first release, you'll be able to install using:
```bash
brew tap StCredZero/asq
brew install asq
```

## Usage

### Generate a Tree-sitter Query

To generate a tree-sitter query from a Go file that contains code marked with `//asq_start` and `//asq_end` comments:

```bash
asq tree-sitter path/to/file.go
```

Example input file:
```go
package example

func Example() {
    //asq_start
    e.Inst().Foo()
    //asq_end
}
```

### Search Using Generated Query

To search for matches of the generated query in all Go files recursively from the current directory:

```bash
asq query path/to/file.go
```

The output will show matches with their file paths, line numbers, and column numbers:
```
//asq_match path/to/match1.go:10:4
e.Inst().Foo()
//asq_match path/to/match2.go:15:2
e.Inst().Foo()
```

## License

MIT License - see [LICENSE](LICENSE) for details.
