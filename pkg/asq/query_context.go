package asq

import (
	"go/ast"
	"go/token"
	"strings"
)

// RangeInterval represents an active interval for a wildcard tag
type RangeInterval struct {
	comment  *ast.Comment
	Start    token.Pos // Start offset in the line
	TokenEnd token.Pos // End offset in the line
	End      token.Pos
}

// QueryContext is an internal struct used during the first pass of AST processing
// to track which identifiers should be treated as wildcards based on active intervals.
type QueryContext struct {
	wildcardRanges []RangeInterval // Active intervals for wildcard tags
}

// NewQueryContext creates a new QueryContext instance
func NewQueryContext(file *ast.File) (*QueryContext, token.Pos, token.Pos) {
	queryContext := &QueryContext{
		wildcardRanges: make([]RangeInterval, 0),
	}

	// Find the code block between comments
	var collectingComments bool
	var startPos, endPos token.Pos
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			// Normalize comment text by removing comment markers and whitespace
			trimmed := strings.TrimSpace(c.Text)
			trimmed = strings.TrimPrefix(trimmed, "//")
			trimmed = strings.TrimPrefix(trimmed, "/*")
			trimmed = strings.TrimSuffix(trimmed, "*/")
			trimmed = strings.TrimSpace(trimmed)

			if trimmed == "asq_start" {
				startPos = c.End()
				collectingComments = true
			} else if trimmed == "asq_end" {
				endPos = c.Pos()
				break
			}
			if collectingComments && c.Text == "/***/" {
				queryContext.AddInterval(c)
			}
		}
	}
	// pad out the last interval to the //asq_end boundary
	queryContext.SetLastIntervalEnd(endPos)
	return queryContext, startPos, endPos
}

// SetLastIntervalEnd sets the end of the last wildcard interval
func (p *QueryContext) SetLastIntervalEnd(start token.Pos) {
	if len(p.wildcardRanges) > 0 {
		lastInterval := p.wildcardRanges[len(p.wildcardRanges)-1]
		Debug("setting last wildcard range %d-%d to %d", lastInterval.Start, lastInterval.End, start-1)
		lastInterval.End = start
		p.wildcardRanges[len(p.wildcardRanges)-1] = lastInterval
	}
}

// AddInterval adds a new wildcard interval
func (p *QueryContext) AddInterval(c *ast.Comment) {
	start, end := c.Pos(), c.End()
	Debug("=== start add %d-%d", start, end)
	if len(p.wildcardRanges) > 0 {
		p.SetLastIntervalEnd(start - 1)
	}
	newInterval := RangeInterval{
		comment:  c,
		Start:    start,
		End:      end,
		TokenEnd: end,
	}
	Debug("=== adding wildcard range %d-%d", newInterval.Start, newInterval.End)
	p.wildcardRanges = append(p.wildcardRanges, newInterval)
}

// IsWildcard checks if the given node should be treated as a wildcard.
// A node is considered a wildcard if either:
// 1. It is an ast.Ident node with a name prefixed by "_asq_"
// 2. It is the first syntactic entity in an active interval
func (p *QueryContext) IsWildcard(node ast.Node) bool {
	// Check for _asq_ prefix in identifiers
	if ident, isIdent := node.(*ast.Ident); isIdent {
		if strings.HasPrefix(ident.Name, "_asq_") {
			return true
		}
	}

	// Check for interval-based wildcards
	for {
		if len(p.wildcardRanges) == 0 {
			return false
		}
		firstRange := p.wildcardRanges[0]
		nodePos, nodeEnd := node.Pos(), node.End()
		if firstRange.Start < nodePos && firstRange.End >= nodeEnd {
			p.wildcardRanges = p.wildcardRanges[1:]
			_, isIdent := node.(*ast.Ident)
			return isIdent
		} else if firstRange.End < nodePos {
			p.wildcardRanges = p.wildcardRanges[1:]
			continue
		} else if firstRange.End > nodeEnd {
			break
		}
	}
	return false
}
