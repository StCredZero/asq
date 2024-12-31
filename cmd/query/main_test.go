package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"fmt"
	"testing"
)

func TestMainIntegration(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "asq-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := map[string]string{
		"asq_query.go": `package test

type E struct{}

func (e *E) Inst() *E { return e }

func asq_query() {
	e := &E{}
	e.Inst().Foo()
	asq_end()
}`,
		"test001.go": `package test

type E struct{}

func (e *E) Inst() *E { return e }

func Test1() {
    e := &E{}
    e.Inst().Foo()
}

func Test2() {
    e := &E{}
    x := 42
    e.Inst().Foo()
    y := 43
}`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", filepath.Dir(fullPath), err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Save original args and working directory
	oldArgs := os.Args
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		os.Args = oldArgs
		os.Chdir(oldWd)
	}()

	// Change to test directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:    "No subcommand",
			args:    []string{"asq"},
			wantErr: true,
		},
		{
			name: "Find command with pattern",
			args: []string{"asq", "find", "asq_query.go"},
			contains: []string{
				"test001.go:6",
				"test001.go:12",
			},
		},
		{
			name:    "Find command with non-existent file",
			args:    []string{"asq", "find", "nonexistent.go"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Set test args
			os.Args = tt.args

			// Create a channel to capture exit code
			exitChan := make(chan int, 1)
			oldOsExit := osExit
			osExit = func(code int) {
				exitChan <- code
				panic(fmt.Sprintf("os.Exit(%d)", code))
			}
			defer func() { osExit = oldOsExit }()

			// Run main and capture exit
			go func() {
				defer func() {
					if r := recover(); r != nil {
						if exitStr, ok := r.(string); ok && strings.HasPrefix(exitStr, "os.Exit(") {
							// Already handled
							return
						}
						// Unexpected panic
						t.Errorf("Unexpected panic: %v", r)
						exitChan <- 2
					}
				}()
				main()
				exitChan <- 0
			}()

			// Wait for exit code
			exitCode := <-exitChan

			// Restore stdout
			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			os.Stdout = oldStdout
			if (exitCode != 0) != tt.wantErr {
				t.Errorf("main() exit = %v, wantErr %v", exitCode, tt.wantErr)
			}

			// Check output
			output := buf.String()
			t.Logf("\nTest case: %s\nArgs: %v\nOutput:\n%s\n", tt.name, tt.args, output)
			
			if tt.wantErr {
				if !strings.Contains(output, "Usage:") || !strings.Contains(output, "error: No subcommand specified") {
					t.Errorf("Expected usage and error message not found in output.\nGot output:\n%s", output)
				}
			} else {
				for _, want := range tt.contains {
					if !strings.Contains(output, want) {
						t.Errorf("Expected output to contain %q but it didn't.\nFull output:\n%s", want, output)
					}
				}
			}
		})
	}
}
