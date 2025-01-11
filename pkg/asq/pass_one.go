package asq

import (
	"go/ast"
	"go/token"
)

// passOne is an internal struct used during the first pass of AST processing
// to track which identifiers should be treated as wildcards.
type passOne struct {
	wildcardIdent map[*ast.Ident]bool
	fset          *token.FileSet
}

// newPassOne creates a new passOne instance
func newPassOne(fset *token.FileSet) *passOne {
	return &passOne{
		wildcardIdent: make(map[*ast.Ident]bool),
		fset:          fset,
	}
}

// isWildcard checks if the given node should be treated as a wildcard.
// Currently only supports ast.Ident nodes.
func (p *passOne) isWildcard(node ast.Node) bool {
	if ident, ok := node.(*ast.Ident); ok {
		return p.wildcardIdent[ident]
	}
	return false
}

// markWildcard marks an identifier as a wildcard
func (p *passOne) markWildcard(ident *ast.Ident) {
	p.wildcardIdent[ident] = true
}
