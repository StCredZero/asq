package query

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// ExtractQueryPattern extracts the pattern from an asq_query function,
// excluding any asq_end() call at the end of the function body.
func ExtractQueryPattern(fset *token.FileSet, filepath string) (*ast.BlockStmt, error) {
	// Parse the query file
	queryFile, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing query file: %v", err)
	}

	// Find the asq_query function and extract its body
	var queryBody *ast.BlockStmt
	ast.Inspect(queryFile, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "asq_query" {
			queryBody = fn.Body
			// Filter out asq_end() call from queryBody
			for i, stmt := range queryBody.List {
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
						if funIdent, ok := callExpr.Fun.(*ast.Ident); ok && funIdent.Name == "asq_end" {
							queryBody.List = append(queryBody.List[:i], queryBody.List[i+1:]...)
							break
						}
					}
				}
			}
			return false
		}
		return true
	})

	if queryBody == nil {
		return nil, fmt.Errorf("asq_query function not found in %s", filepath)
	}

	return queryBody, nil
}

// MatchPattern checks if a given AST node matches the pattern
func MatchPattern(pattern *ast.BlockStmt, node ast.Node) bool {
	if pattern == nil || node == nil {
		return false
	}

	switch n := node.(type) {
	case *ast.BlockStmt:
		// Check if any subsequence of statements in the block matches our pattern
		for i := 0; i <= len(n.List)-len(pattern.List); i++ {
			match := true
			for j := 0; j < len(pattern.List); j++ {
				if !matchStatement(pattern.List[j], n.List[i+j]) {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
		// If no match found in this block, continue searching in child nodes
		for _, stmt := range n.List {
			if MatchPattern(pattern, stmt) {
				return true
			}
		}
	case *ast.FuncDecl:
		if n.Body != nil {
			return MatchPattern(pattern, n.Body)
		}
	case *ast.File:
		// For source files, check all function bodies
		for _, decl := range n.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if funcDecl.Body != nil && MatchPattern(pattern, funcDecl.Body) {
					return true
				}
			}
		}
	case *ast.IfStmt:
		// Check the body of if statements
		if n.Body != nil && MatchPattern(pattern, n.Body) {
			return true
		}
		if n.Else != nil && MatchPattern(pattern, n.Else) {
			return true
		}
	case *ast.ForStmt:
		// Check the body of for loops
		if n.Body != nil && MatchPattern(pattern, n.Body) {
			return true
		}
	case *ast.RangeStmt:
		// Check the body of range statements
		if n.Body != nil && MatchPattern(pattern, n.Body) {
			return true
		}
	}
	return false
}

// matchStatement checks if two AST statements match
func matchStatement(pattern, node ast.Stmt) bool {
	if pattern == nil || node == nil {
		return pattern == node
	}

	// Check if the types match
	if reflect.TypeOf(pattern) != reflect.TypeOf(node) {
		return false
	}

	// Handle different types of statements
	switch p := pattern.(type) {
	case *ast.ExprStmt:
		n, ok := node.(*ast.ExprStmt)
		if !ok {
			return false
		}
		return matchExpr(p.X, n.X)
	case *ast.AssignStmt:
		n, ok := node.(*ast.AssignStmt)
		if !ok {
			return false
		}
		return p.Tok == n.Tok && matchExprList(p.Lhs, n.Lhs) && matchExprList(p.Rhs, n.Rhs)
	default:
		// For other types, compare the entire statement
		return reflect.DeepEqual(pattern, node)
	}
}

// matchExpr checks if two expressions match
func matchExpr(pattern, node ast.Expr) bool {
	if pattern == nil || node == nil {
		return pattern == node
	}

	// Check if the types match
	if reflect.TypeOf(pattern) != reflect.TypeOf(node) {
		return false
	}

	switch p := pattern.(type) {
	case *ast.CallExpr:
		n, ok := node.(*ast.CallExpr)
		if !ok {
			return false
		}
		return matchExpr(p.Fun, n.Fun) && matchExprList(p.Args, n.Args)
	case *ast.SelectorExpr:
		n, ok := node.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		return matchExpr(p.X, n.X) && p.Sel.Name == n.Sel.Name
	case *ast.Ident:
		n, ok := node.(*ast.Ident)
		if !ok {
			return false
		}
		return p.Name == n.Name
	case *ast.UnaryExpr:
		n, ok := node.(*ast.UnaryExpr)
		if !ok {
			return false
		}
		return p.Op == n.Op && matchExpr(p.X, n.X)
	case *ast.CompositeLit:
		n, ok := node.(*ast.CompositeLit)
		if !ok {
			return false
		}
		return matchExpr(p.Type, n.Type) && matchExprList(p.Elts, n.Elts)
	default:
		// For other types, compare the entire expression
		return reflect.DeepEqual(pattern, node)
	}
}

// matchExprList checks if two lists of expressions match
func matchExprList(pattern, node []ast.Expr) bool {
	if len(pattern) != len(node) {
		return false
	}
	for i := range pattern {
		if !matchExpr(pattern[i], node[i]) {
			return false
		}
	}
	return true
}

// FindMatches finds all matches of the pattern in the given file
func FindMatches(pattern *ast.BlockStmt, filePath string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing file: %v", err)
	}

	var matches []string

	// Extract pattern statements, excluding asq_end() call
	var patternStmts []ast.Stmt
	for _, stmt := range pattern.List {
		// Skip asq_end() call
		if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
			if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
				if ident, ok := callExpr.Fun.(*ast.Ident); ok && ident.Name == "asq_end" {
					continue
				}
			}
		}
		patternStmts = append(patternStmts, stmt)
	}

	// Find all function declarations first
	var funcs []*ast.FuncDecl
	ast.Inspect(node, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			// Skip asq_ prefixed functions and functions without bodies
			if !strings.HasPrefix(fd.Name.Name, "asq_") && fd.Body != nil {
				funcs = append(funcs, fd)
			}
			return false
		}
		return true
	})

	// Process each function body
	for _, fn := range funcs {
		body := fn.Body
		funcStart := fset.Position(fn.Body.Lbrace).Line

		// Check each position in the block for a potential match
		for i := 0; i < len(body.List); i++ {
			// Get the line number for the first statement
			firstStmtLine := fset.Position(body.List[i].Pos()).Line
			if firstStmtLine <= funcStart {
				continue
			}

			// Try to find the first statement match
			if !matchStatement(patternStmts[0], body.List[i]) {
				continue
			}

			// Look for the second statement after the first match
			for j := i + 1; j < len(body.List); j++ {
				if matchStatement(patternStmts[1], body.List[j]) {
					
					match := fmt.Sprintf("%s:%d", filePath, firstStmtLine)
					
					// Only add unique matches
					found := false
					for _, existing := range matches {
						if existing == match {
							found = true
							break
						}
					}
					if !found && firstStmtLine > funcStart {
						matches = append(matches, match)
					}
					break // Found a match, move to next potential first statement
				}
			}
		}
	}

	// Sort matches by line number for consistent output
	sort.Slice(matches, func(i, j int) bool {
		iLine := strings.Split(matches[i], ":")[1]
		jLine := strings.Split(matches[j], ":")[1]
		iNum, _ := strconv.Atoi(iLine)
		jNum, _ := strconv.Atoi(jLine)
		return iNum < jNum
	})
	return matches, nil
}
