package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCloneAsqTest(t *testing.T) {
	// 1. Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "asq-external-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 2. Clone the asq-test repository
	cloneCmd := exec.Command("git", "clone", "https://github.com/StCredZero/asq-test.git")
	cloneCmd.Dir = tmpDir
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to clone asq-test repository: %v\nOutput: %s", err, string(out))
	}

	// 3. Look for testcases/ subdirectories in asq-test
	testcasesDir := filepath.Join(tmpDir, "asq-test", "testcases")
	entries, err := os.ReadDir(testcasesDir)
	if err != nil {
		t.Skipf("Skipping test: testcases directory not found: %v\nThis is expected if the asq-test repository hasn't been populated with test cases yet.", err)
		return
	}
	if len(entries) == 0 {
		t.Skip("Skipping test: no test cases found in testcases directory")
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		testDir := filepath.Join(testcasesDir, entry.Name())
		// e.g. testDir = .../testcases/test0001

		// 4. Build the 'asq query _asq_query.go' command
		cmd := exec.Command("asq", "query", "_asq_query.go")
		cmd.Dir = testDir
		var stdoutBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf

		// 5. Run the command
		if err := cmd.Run(); err != nil {
			t.Fatalf("asq query command failed in %s: %v", testDir, err)
		}

		// 6. Compare output to 'expected' file
		expectedFilePath := filepath.Join(testDir, "expected")
		expectedContent, err := os.ReadFile(expectedFilePath)
		if err != nil {
			t.Fatalf("Failed to read expected file '%s': %v", expectedFilePath, err)
		}

		actualOutput := stdoutBuf.String()
		// Normalize whitespace like in ts_query_test.go
		actualLines := strings.Split(actualOutput, "\n")
		expectedLines := strings.Split(string(expectedContent), "\n")
		
		// Normalize both sets of lines
		for i := range actualLines {
			actualLines[i] = strings.TrimSpace(actualLines[i])
		}
		for i := range expectedLines {
			expectedLines[i] = strings.TrimSpace(expectedLines[i])
		}
		
		actualNormalized := strings.Join(actualLines, "\n")
		expectedNormalized := strings.Join(expectedLines, "\n")

		if actualNormalized != expectedNormalized {
			t.Errorf("Mismatch in directory %s\nExpected:\n%s\nGot:\n%s",
				entry.Name(), expectedNormalized, actualNormalized)
		}
	}
}
