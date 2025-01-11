package asq

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestPassOne_IsWildcard(t *testing.T) {
	fset := token.NewFileSet()
	p := newPassOne(fset)

	// Test Ident node
	ident := &ast.Ident{Name: "testIdent"}
	if p.isWildcard(ident) {
		t.Error("Expected non-wildcarded Ident to return false")
	}

	p.markWildcard(ident)
	if !p.isWildcard(ident) {
		t.Error("Expected wildcarded Ident to return true")
	}

	// Test non-Ident node
	nonIdent := &ast.BasicLit{Kind: token.INT, Value: "42"}
	if p.isWildcard(nonIdent) {
		t.Error("Expected non-Ident node to return false")
	}
}

func TestPassOne_MarkWildcard(t *testing.T) {
	fset := token.NewFileSet()
	p := newPassOne(fset)

	ident1 := &ast.Ident{Name: "ident1"}
	ident2 := &ast.Ident{Name: "ident2"}

	// Mark first ident as wildcard
	p.markWildcard(ident1)
	if !p.isWildcard(ident1) {
		t.Error("Expected ident1 to be marked as wildcard")
	}
	if p.isWildcard(ident2) {
		t.Error("Expected ident2 to not be marked as wildcard")
	}

	// Mark second ident as wildcard
	p.markWildcard(ident2)
	if !p.isWildcard(ident2) {
		t.Error("Expected ident2 to be marked as wildcard")
	}
}
