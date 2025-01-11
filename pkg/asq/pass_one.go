package asq

import (
	"go/ast"
	"go/token"
)

// RangeInterval represents an active interval for a wildcard tag
type RangeInterval struct {
	Line  int // Line number where interval occurs
	Start int // Start offset in the line
	End   int // End offset in the line
	Used  bool // Whether this interval has been used
}

// passOne is an internal struct used during the first pass of AST processing
// to track which identifiers should be treated as wildcards based on active intervals.
type passOne struct {
	wildcardRanges []RangeInterval // Active intervals for wildcard tags
	fset          *token.FileSet
}

// newPassOne creates a new passOne instance
func newPassOne(fset *token.FileSet) *passOne {
	return &passOne{
		wildcardRanges: make([]RangeInterval, 0),
		fset:          fset,
	}
}

// addInterval adds a new wildcard interval
func (p *passOne) addInterval(line, start, end int) {
	p.wildcardRanges = append(p.wildcardRanges, RangeInterval{
		Line:  line,
		Start: start,
		End:   end,
		Used:  false,
	})
}

// isWildcard checks if the given node should be treated as a wildcard.
// Currently only supports ast.Ident nodes and checks if the node is the first
// syntactic entity in an active interval.
func (p *passOne) isWildcard(node ast.Node) bool {
	ident, ok := node.(*ast.Ident)
	if !ok {
		return false
	}

	pos := p.fset.Position(ident.Pos())
	offset := pos.Offset
	line := pos.Line

	// Check if this identifier is the first syntactic entity in any active interval
	for i, interval := range p.wildcardRanges {
		if interval.Line == line && !interval.Used && offset >= interval.Start && offset < interval.End {
			// Mark this interval as used since we found its first syntactic entity
			p.wildcardRanges[i].Used = true
			return true
		}
	}
	return false
}

// markWildcard is deprecated and will be removed.
// Use addInterval instead to define wildcard ranges.
func (p *passOne) markWildcard(ident *ast.Ident) {
	// No-op - we now use intervals instead of direct marking
}
