// Package matcher provides AST pattern matching functionality for Go code.
//
// The matcher package is responsible for comparing AST nodes for structural equality,
// supporting various node types including method chains (e.g., e.Inst().Foo()).
// It is designed to work with the query package, which handles pattern extraction
// and preprocessing (such as excluding asq_end() calls) before pattern matching.
//
// Example usage:
//
//	if matcher.ASTMatch(pattern, target) {
//	    // AST nodes match structurally
//	}
package matcher
