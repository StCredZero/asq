package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestMetaqQuery(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "metaq-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file
	testCode := `package example1

type Thingy1 struct{}

func (t Thingy1) Inst() Thingy1 {
	return t
}
func (t Thingy1) Foo() bool {
	return true
}

var e = new(Thingy1)

func asq_query2() {
	//asq_start
	e.Inst().Foo()
	//asq_end
}`

	testFile := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Run metaq
	os.Args = []string{"metaq", testFile}
	main()

	// Restore stdout and get output
	w.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	expected := "(call_expression function: (selector_expression operand: (call_expression function: (selector_expression operand: (identifier) field: (field_identifier)) arguments: (argument_list)) field: (field_identifier)) arguments: (argument_list)) @x\n"
	if got := string(output); got != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, got)
	}
}
