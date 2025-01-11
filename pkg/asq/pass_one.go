package asq

import (
	"go/ast"
	"go/token"
	"strings"
)

// RangeInterval represents an active interval for a wildcard tag
type RangeInterval struct {
	Start    token.Pos // Start offset in the line
	TokenEnd token.Pos // End offset in the line
	End      token.Pos
}

// passOne is an internal struct used during the first pass of AST processing
// to track which identifiers should be treated as wildcards based on active intervals.
type passOne struct {
	wildcardRanges []RangeInterval // Active intervals for wildcard tags
}

// newPassOne creates a new passOne instance
func newPassOne(file *ast.File) *passOne {
	passOneData := &passOne{
		wildcardRanges: make([]RangeInterval, 0),
	}

	// Find the code block between comments
	var collectingComments bool
	var endPos token.Pos
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if strings.TrimSpace(c.Text) == "//asq_start" {
				collectingComments = true
			} else if strings.TrimSpace(c.Text) == "//asq_end" {
				endPos = c.Pos()
				break
			}
			if collectingComments && c.Text == "/***/" {
				passOneData.addInterval(cg.Pos(), cg.End())
			}
		}
	}
	// pad out the last interval to the //asq_end boundary
	passOneData.setLastIntervalEnd(endPos)
	return passOneData
}

// addInterval adds a new wildcard interval
func (p *passOne) setLastIntervalEnd(start token.Pos) {
	if len(p.wildcardRanges) > 0 {
		lastInterval := p.wildcardRanges[len(p.wildcardRanges)-1]
		lastInterval.End = start - 1
		p.wildcardRanges[len(p.wildcardRanges)-1] = lastInterval
	}
}

// addInterval adds a new wildcard interval
func (p *passOne) addInterval(start, end token.Pos) {
	if len(p.wildcardRanges) > 0 {
		p.setLastIntervalEnd(start - 1)
	}
	p.wildcardRanges = append(p.wildcardRanges, RangeInterval{
		Start:    start,
		End:      end,
		TokenEnd: end,
	})
}

// isWildcard checks if the given node should be treated as a wildcard.
// Currently only supports ast.Ident nodes and checks if the node is the first
// syntactic entity in an active interval.
func (p *passOne) isWildcard(node ast.Node) bool {
	for {
		if len(p.wildcardRanges) == 0 {
			return false
		}
		firstRange := p.wildcardRanges[0]
		if firstRange.Start < node.Pos() && firstRange.End >= node.End() {
			p.wildcardRanges = p.wildcardRanges[1:]
			_, isIdent := node.(*ast.Ident)
			return isIdent
		} else if firstRange.End < node.Pos() {
			p.wildcardRanges = p.wildcardRanges[1:]
			continue
		} else if firstRange.End > node.End() {
			break
		}
	}
	return false
}
