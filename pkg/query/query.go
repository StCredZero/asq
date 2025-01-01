package query

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
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
			queryBody = &ast.BlockStmt{List: make([]ast.Stmt, 0)}
			// Extract statements until asq_end() call
			for _, stmt := range fn.Body.List {
				// Skip asq_end() calls
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
						if ident, ok := callExpr.Fun.(*ast.Ident); ok && ident.Name == "asq_end" {
							continue
						}
					}
				}

				// For method calls, ensure we capture the full chain
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					if call, ok := exprStmt.X.(*ast.CallExpr); ok {
						if _, ok := call.Fun.(*ast.SelectorExpr); ok {
							// Found a method call, add it to the pattern
							queryBody.List = append(queryBody.List, stmt)
							fmt.Printf("Debug: Found method call statement: %T\n", stmt)
							continue
						}
					}
				}

				// Add all other statements
				queryBody.List = append(queryBody.List, stmt)
				fmt.Printf("Debug: Found statement: %T\n", stmt)
			}
			fmt.Printf("Debug: Extracted %d statements from asq_query\n", len(queryBody.List))
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
func MatchPattern(pattern *ast.BlockStmt, node ast.Node, fset *token.FileSet) bool {
	if pattern == nil || node == nil {
		return false
	}

	switch n := node.(type) {
	case *ast.BlockStmt:
		// Check if any subsequence of statements in the block matches our pattern
		for i := 0; i <= len(n.List)-len(pattern.List); i++ {
			match := true
			for j := 0; j < len(pattern.List); j++ {
				if !matchStatement(pattern.List[j], n.List[i+j], fset) {
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
			if MatchPattern(pattern, stmt, fset) {
				return true
			}
		}
	case *ast.FuncDecl:
		if n.Body != nil {
			return MatchPattern(pattern, n.Body, fset)
		}
	case *ast.File:
		// For source files, check all function bodies
		for _, decl := range n.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if funcDecl.Body != nil && MatchPattern(pattern, funcDecl.Body, fset) {
					return true
				}
			}
		}
	case *ast.IfStmt:
		// Check the body of if statements
		if n.Body != nil && MatchPattern(pattern, n.Body, fset) {
			return true
		}
		if n.Else != nil && MatchPattern(pattern, n.Else, fset) {
			return true
		}
	case *ast.ForStmt:
		// Check the body of for loops
		if n.Body != nil && MatchPattern(pattern, n.Body, fset) {
			return true
		}
	case *ast.RangeStmt:
		// Check the body of range statements
		if n.Body != nil && MatchPattern(pattern, n.Body, fset) {
			return true
		}
	}
	return false
}

// matchStatement checks if two AST statements match
func matchStatement(pattern, node ast.Stmt, fset *token.FileSet) bool {
	if pattern == nil || node == nil {
		return false
	}

	fmt.Printf("Debug: Matching statement types %T with %T\n", pattern, node)

	// Extract all method calls from pattern and node
	var patternCalls, nodeCalls []*ast.CallExpr

	// Extract method calls from pattern
	if exprStmt, ok := pattern.(*ast.ExprStmt); ok {
		if call, ok := exprStmt.X.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name != "asq_end" {
						fmt.Printf("Debug: Found pattern call at line %d\n", fset.Position(call.Pos()).Line)
						patternCalls = append(patternCalls, call)
					}
				} else {
					fmt.Printf("Debug: Found pattern call at line %d\n", fset.Position(call.Pos()).Line)
					patternCalls = append(patternCalls, call)
				}
			}
		}
	}

	// Extract method calls from node
	if exprStmt, ok := node.(*ast.ExprStmt); ok {
		if call, ok := exprStmt.X.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name != "asq_end" {
						fmt.Printf("Debug: Found node call at line %d\n", fset.Position(call.Pos()).Line)
						nodeCalls = append(nodeCalls, call)
					}
				} else {
					fmt.Printf("Debug: Found node call at line %d\n", fset.Position(call.Pos()).Line)
					nodeCalls = append(nodeCalls, call)
				}
			}
		}
	}

	// Try to match any pattern call with any node call
	for _, patternCall := range patternCalls {
		for _, nodeCall := range nodeCalls {
			fmt.Printf("Debug: Found method calls, checking method chain\n")
			fmt.Printf("Debug: Pattern call: %#v\n", patternCall)
			fmt.Printf("Debug: Node call: %#v\n", nodeCall)
			matched := matchMethodChain(patternCall, nodeCall)
			if matched {
				fmt.Printf("Debug: Found matching method chain\n")
				return true
			}
		}
	}

	fmt.Printf("Debug: No method calls found to match\n")
	return false
}

// matchMethodChain checks if two expressions form the same method chain
func matchMethodChain(pattern, node ast.Expr) bool {
	if pattern == nil || node == nil {
		fmt.Printf("Debug: matchMethodChain - nil pattern or node\n")
		return false
	}

	fmt.Printf("Debug: Matching pattern type %T with node type %T\n", pattern, node)

	// Extract method chains
	var patternChain, nodeChain []string

	// Helper function to extract method chain
	var extractChain func(expr ast.Expr) []string
	extractChain = func(expr ast.Expr) []string {
		var chain []string
		curr := expr
		for curr != nil {
			switch e := curr.(type) {
			case *ast.CallExpr:
				if selExpr, ok := e.Fun.(*ast.SelectorExpr); ok {
					fmt.Printf("Debug: Found method in chain: %s\n", selExpr.Sel.Name)
					chain = append([]string{selExpr.Sel.Name}, chain...)
					curr = selExpr.X
				} else {
					curr = nil
				}
			case *ast.SelectorExpr:
				fmt.Printf("Debug: Found selector in chain: %s\n", e.Sel.Name)
				chain = append([]string{e.Sel.Name}, chain...)
				curr = e.X
			case *ast.Ident:
				// Don't include the variable name in the chain
				fmt.Printf("Debug: Found identifier in chain: %s\n", e.Name)
				curr = nil
			default:
				fmt.Printf("Debug: Found unknown type in chain: %T\n", e)
				curr = nil
			}
		}
		return chain
	}

	patternChain = extractChain(pattern)
	nodeChain = extractChain(node)

	fmt.Printf("Debug: Pattern chain: %v\n", patternChain)
	fmt.Printf("Debug: Node chain: %v\n", nodeChain)

	// Compare chains
	if len(patternChain) != len(nodeChain) {
		fmt.Printf("Debug: Chain length mismatch: pattern=%d, node=%d\n", len(patternChain), len(nodeChain))
		return false
	}

	for i := range patternChain {
		if patternChain[i] != nodeChain[i] {
			fmt.Printf("Debug: Chain mismatch at position %d: pattern=%s, node=%s\n", i, patternChain[i], nodeChain[i])
			return false
		}
	}

	fmt.Printf("Debug: Found matching method chain\n")
	return true
}

// matchMethodPattern checks if a statement matches a method declaration
func matchMethodPattern(pattern ast.Stmt, method *ast.FuncDecl) bool {
	if pattern == nil || method == nil {
		return false
	}

	// Extract the pattern's method call
	var patternCall *ast.CallExpr
	if exprStmt, ok := pattern.(*ast.ExprStmt); ok {
		if call, ok := exprStmt.X.(*ast.CallExpr); ok {
			patternCall = call
		}
	}
	if patternCall == nil {
		return false
	}

	// Get the method name from the pattern
	var patternMethodName string
	if sel, ok := patternCall.Fun.(*ast.SelectorExpr); ok {
		patternMethodName = sel.Sel.Name
	} else if ident, ok := patternCall.Fun.(*ast.Ident); ok {
		patternMethodName = ident.Name
	}

	// Compare method names
	return method.Name.Name == patternMethodName
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
		// Only match the method name, not the receiver type
		return p.Sel.Name == n.Sel.Name && matchExpr(p.X, n.X)
	case *ast.Ident:
		_, ok := node.(*ast.Ident)
		if !ok {
			return false
		}
		// Allow any identifier when matching method chains
		return true
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

// Helper function to check if an expression contains another expression
func containsExpr(parent, child ast.Expr) bool {
	found := false
	ast.Inspect(parent, func(node ast.Node) bool {
		if node == child {
			found = true
			return false
		}
		return true
	})
	return found
}

// FindMatches finds all matches of the pattern in the given file
func FindMatches(pattern *ast.BlockStmt, filePath string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing file: %v", err)
	}

	// Skip files with asq_ prefix except asq_query.go
	if strings.HasPrefix(filepath.Base(filePath), "asq_") && filepath.Base(filePath) != "asq_query.go" {
		return nil, nil
	}

	var matches []string

	// Extract pattern method calls
	var patternCalls []*ast.CallExpr
	if pattern == nil {
		return nil, fmt.Errorf("pattern is nil")
	}
	fmt.Printf("Debug: Extracting pattern calls from %d statements\n", len(pattern.List))
	for _, stmt := range pattern.List {
		fmt.Printf("Debug: Processing statement type: %T\n", stmt)
		switch s := stmt.(type) {
		case *ast.ExprStmt:
			if call, ok := s.X.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					fmt.Printf("Debug: Found pattern method call: %s\n", sel.Sel.Name)
					if x, ok := sel.X.(*ast.SelectorExpr); ok {
						fmt.Printf("Debug: Pattern chain: %s.%s\n", x.Sel.Name, sel.Sel.Name)
					}
					patternCalls = append(patternCalls, call)
				}
			}
		case *ast.IfStmt:
			if call, ok := s.Cond.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					fmt.Printf("Debug: Found pattern method call in if condition: %s\n", sel.Sel.Name)
					if x, ok := sel.X.(*ast.SelectorExpr); ok {
						fmt.Printf("Debug: Pattern chain in if: %s.%s\n", x.Sel.Name, sel.Sel.Name)
					}
					patternCalls = append(patternCalls, call)
				}
			}
		}
	}

	fmt.Printf("Debug: Looking for pattern with %d method calls\n", len(patternCalls))

	// Find all method calls in regular functions
	var currentFunc *ast.FuncDecl
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		switch n := n.(type) {
		case *ast.FuncDecl:
			// Skip method declarations and asq_ prefixed functions
			if n.Recv != nil || strings.HasPrefix(n.Name.Name, "asq_") {
				currentFunc = nil
				return false
			}
			currentFunc = n
			fmt.Printf("Debug: Checking function %s\n", n.Name.Name)
			return true

		case *ast.CallExpr:
			// Only process calls within function bodies
			if currentFunc == nil || currentFunc.Body == nil {
				return true
			}

			// Check if this is a method call
			sel, ok := n.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			// Skip asq_end calls
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "asq_end" {
				return true
			}

			// Get line number
			line := fset.Position(n.Pos()).Line
			fmt.Printf("Debug: Checking call at line %d in function %s\n", line, currentFunc.Name.Name)

			// Log method chain details
			if sel2, ok := sel.X.(*ast.SelectorExpr); ok {
				fmt.Printf("Debug: Found method chain at line %d: %s.%s\n", line, sel2.Sel.Name, sel.Sel.Name)
			} else if ident, ok := sel.X.(*ast.Ident); ok {
				fmt.Printf("Debug: Found method call at line %d: %s.%s\n", line, ident.Name, sel.Sel.Name)
			}

			// Check against all pattern calls
			for _, patternCall := range patternCalls {
				patternSel, ok := patternCall.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				fmt.Printf("Debug: Comparing with pattern method: %s\n", patternSel.Sel.Name)
				if x, ok := patternSel.X.(*ast.SelectorExpr); ok {
					fmt.Printf("Debug: Pattern chain: %s.%s\n", x.Sel.Name, patternSel.Sel.Name)
				} else if ident, ok := patternSel.X.(*ast.Ident); ok {
					fmt.Printf("Debug: Pattern method: %s.%s\n", ident.Name, patternSel.Sel.Name)
				}

				if matchMethodChain(patternCall, n) {
					match := fmt.Sprintf("%s:%d", filepath.Base(filePath), line)
					matches = append(matches, match)
					fmt.Printf("Debug: Found match at line %d in function %s\n", line, currentFunc.Name.Name)
				}
			}
			return true

		case nil:
			if currentFunc != nil && currentFunc.End() == n.Pos() {
				currentFunc = nil
			}
			return true
		}
		return true
	})

	// Sort matches by line number
	sort.Slice(matches, func(i, j int) bool {
		iLine := strings.Split(matches[i], ":")[1]
		jLine := strings.Split(matches[j], ":")[1]
		iNum, _ := strconv.Atoi(iLine)
		jNum, _ := strconv.Atoi(jLine)
		return iNum < jNum
	})

	// Remove duplicates while preserving order
	if len(matches) > 0 {
		unique := make([]string, 0, len(matches))
		seen := make(map[string]bool)
		for _, match := range matches {
			if !seen[match] {
				seen[match] = true
				unique = append(unique, match)
			}
		}
		matches = unique
	}

	return matches, nil
}
