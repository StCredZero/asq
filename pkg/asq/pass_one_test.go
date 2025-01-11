package asq

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestPassOne_MultipleWildcardTags(t *testing.T) {
	fset := token.NewFileSet()
	p := newPassOne(fset)

	// Create test file with multiple intervals on one line
	src := `package test
func main() {
	//asq_start
	/***/if /***/x >= 10 /***/&& y < 20 /***/{ foo.Bar()./***/Baz() }
	//asq_end
}`

	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Add test intervals (simulating what parser would do)
	// Intervals for: if, x, &&, {, and Baz
	p.addInterval(3, 20, 30)  // /***/if
	p.addInterval(3, 35, 45)  // /***/x
	p.addInterval(3, 50, 60)  // /***/&&
	p.addInterval(3, 70, 80)  // /***/{ 
	p.addInterval(3, 90, 100) // /***/Baz

	// Create test identifiers at various positions
	ident1 := &ast.Ident{NamePos: file.Pos(25), Name: "if"}    // in first interval
	ident2 := &ast.Ident{NamePos: file.Pos(40), Name: "x"}     // in second interval
	ident3 := &ast.Ident{NamePos: file.Pos(42), Name: "y"}     // not in any interval
	ident4 := &ast.Ident{NamePos: file.Pos(75), Name: "foo"}   // in fourth interval
	ident5 := &ast.Ident{NamePos: file.Pos(95), Name: "Baz"}   // in fifth interval

	// Test first interval
	if !p.isWildcard(ident1) {
		t.Error("Expected 'if' to be wildcarded")
	}

	// Test second interval
	if !p.isWildcard(ident2) {
		t.Error("Expected 'x' to be wildcarded")
	}

	// Test identifier between intervals
	if p.isWildcard(ident3) {
		t.Error("Expected 'y' to not be wildcarded")
	}

	// Test fourth interval
	if !p.isWildcard(ident4) {
		t.Error("Expected 'foo' to be wildcarded")
	}

	// Test fifth interval
	if !p.isWildcard(ident5) {
		t.Error("Expected 'Baz' to be wildcarded")
	}

	// Test that intervals are marked as used
	for i, interval := range p.wildcardRanges {
		if !interval.Used {
			t.Errorf("Expected interval %d to be marked as used", i)
		}
	}

	// Test non-Ident node
	badNode := &ast.BadExpr{}
	if p.isWildcard(badNode) {
		t.Error("Expected non-Ident node to not be wildcarded")
	}
}

func TestPassOne_WhitespaceAndReformat(t *testing.T) {
	fset := token.NewFileSet()
	p := newPassOne(fset)

	// Test with various whitespace patterns and line continuations
	src := `package test
func main() {
	//asq_start
	/***/   if    /***/   x   >=   10 \
        /***/   &&    y   <    20
	//asq_end
}`

	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Add test intervals with extra whitespace
	p.addInterval(3, 20, 35)  // /***/   if
	p.addInterval(3, 40, 55)  // /***/   x
	p.addInterval(4, 10, 25)  // /***/   &&

	// Create test identifiers
	ident1 := &ast.Ident{NamePos: file.Pos(25), Name: "if"}
	ident2 := &ast.Ident{NamePos: file.Pos(45), Name: "x"}
	ident3 := &ast.Ident{NamePos: file.Pos(15), Name: "&&"}
	ident4 := &ast.Ident{NamePos: file.Pos(20), Name: "y"}

	// Test that whitespace and line continuations don't affect wildcarding
	if !p.isWildcard(ident1) {
		t.Error("Expected 'if' to be wildcarded despite whitespace")
	}
	if !p.isWildcard(ident2) {
		t.Error("Expected 'x' to be wildcarded despite whitespace")
	}
	if !p.isWildcard(ident3) {
		t.Error("Expected '&&' to be wildcarded despite line continuation")
	}
	if p.isWildcard(ident4) {
		t.Error("Expected 'y' to not be wildcarded")
	}
}
