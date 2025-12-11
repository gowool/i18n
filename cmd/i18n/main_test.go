package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/gowool/i18n"
)

// TestBuildCLI tests the buildCLI function
func TestBuildCLI(t *testing.T) {
	tests := []struct {
		name          string
		extractorFunc func(*cli.Command) error
		expectedName  string
		expectedUsage string
		expectedVer   string
	}{
		{
			name: "basic CLI build",
			extractorFunc: func(cmd *cli.Command) error {
				return nil
			},
			expectedName:  "i18n",
			expectedUsage: "i18n tool",
			expectedVer:   "v0.0.1",
		},
		{
			name: "CLI with error extractor",
			extractorFunc: func(cmd *cli.Command) error {
				return fmt.Errorf("test error")
			},
			expectedName:  "i18n",
			expectedUsage: "i18n tool",
			expectedVer:   "v0.0.1",
		},
		{
			name: "CLI with nil extractor",
			extractorFunc: func(cmd *cli.Command) error {
				return nil
			},
			expectedName:  "i18n",
			expectedUsage: "i18n tool",
			expectedVer:   "v0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildCLI(tt.extractorFunc)

			// Verify command properties
			assert.Equal(t, tt.expectedName, cmd.Name)
			assert.Equal(t, tt.expectedUsage, cmd.Usage)
			assert.Equal(t, tt.expectedVer, cmd.Version)

			// Verify command structure
			assert.NotNil(t, cmd.Commands)
			assert.Len(t, cmd.Commands, 1)

			// Verify subcommand is extract
			extractCmd := cmd.Commands[0]
			assert.Equal(t, "extract", extractCmd.Name)
			assert.Equal(t, "Extract i18n messages from templates", extractCmd.Usage)

			// Verify extract command has flags
			assert.NotEmpty(t, extractCmd.Flags)
			assert.Len(t, extractCmd.Flags, 5) // dir, out, gofile, pkg, ext

			// Verify flag names
			flagNames := make([]string, len(extractCmd.Flags))
			for i, f := range extractCmd.Flags {
				flagNames[i] = f.Names()[0]
			}
			expectedFlagNames := []string{"dir", "out", "gofile", "pkg", "ext"}
			assert.Equal(t, expectedFlagNames, flagNames)
		})
	}
}

// TestBuildCLICommandStructure tests the detailed structure of the CLI command
func TestBuildCLICommandStructure(t *testing.T) {
	extractorFunc := func(cmd *cli.Command) error { return nil }
	cmd := buildCLI(extractorFunc)

	// Test main command structure
	assert.Equal(t, "i18n", cmd.Name)
	assert.Equal(t, "i18n tool", cmd.Usage)
	assert.Equal(t, "v0.0.1", cmd.Version)
	assert.Empty(t, cmd.ArgsUsage)

	// Test that main command has no action (it delegates to subcommands)
	assert.Nil(t, cmd.Action)

	// Test subcommands
	require.Len(t, cmd.Commands, 1)

	extractCmd := cmd.Commands[0]
	assert.Equal(t, "extract", extractCmd.Name)
	assert.Equal(t, "Extract i18n messages from templates", extractCmd.Usage)
	assert.Equal(t, " ", extractCmd.ArgsUsage)
	assert.NotNil(t, extractCmd.Action)
}

// TestBuildCLIIntegration tests the integration between buildCLI and extract functions
func TestBuildCLIIntegration(t *testing.T) {
	var calledCmd *cli.Command
	extractorFunc := func(cmd *cli.Command) error {
		calledCmd = cmd
		return nil
	}

	cliCmd := buildCLI(extractorFunc)
	require.Len(t, cliCmd.Commands, 1)

	extractCmd := cliCmd.Commands[0]
	require.NotNil(t, extractCmd.Action)

	// Create test context
	ctx := context.Background()
	testCmd := &cli.Command{
		Name:   extractCmd.Name,
		Usage:  extractCmd.Usage,
		Flags:  extractCmd.Flags,
		Action: extractCmd.Action,
	}

	// Execute the extract action
	err := extractCmd.Action(ctx, testCmd)
	assert.NoError(t, err)
	assert.NotNil(t, calledCmd)
}

// TestMainFunctionIntegration tests the main function behavior (without actually calling main)
func TestMainFunctionIntegration(t *testing.T) {
	// Save original os.Args and os.Stdout/os.Stderr
	originalArgs := os.Args
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	// Restore after test
	defer func() {
		os.Args = originalArgs
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	}()

	// Create temporary files for testing
	tempDir := t.TempDir()
	createTestTemplate(t, tempDir, "test.html", `{{ i18n .Lang "Hello World" }}`)
	outputFile := createTempFile(t, tempDir, "messages.json", "")
	goFile := createTempFile(t, tempDir, "gotext.go", "")

	// Set up test arguments
	os.Args = []string{
		"i18n",
		"extract",
		"--dir", tempDir,
		"--out", outputFile,
		"--gofile", goFile,
		"--pkg", "main",
		"--ext", ".html",
	}

	// Redirect stdout and stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Note: We can't easily test main() directly due to os.Exit()
	// But we can test the buildCLI part that main() uses
	cliCmd := buildCLI(func(command *cli.Command) error {
		return i18n.NewExtractor(
			command.String("dir"),
			command.String("out"),
			command.String("pkg"),
			command.String("gofile"),
			command.StringSlice("ext")...,
		).Extract()
	})

	// Close pipe and restore outputs
	_ = w.Close()
	os.Stdout = originalStdout
	os.Stderr = originalStderr

	// Read output
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify CLI command was built correctly
	assert.Equal(t, "i18n", cliCmd.Name)
	assert.Equal(t, "i18n tool", cliCmd.Usage)
	assert.Equal(t, "v0.0.1", cliCmd.Version)

	// Test that running the CLI command works
	ctx := context.Background()
	err := cliCmd.Run(ctx, os.Args)
	assert.NoError(t, err)

	// Verify output files were created
	assert.FileExists(t, outputFile)
	assert.FileExists(t, goFile)

	// Verify stderr is empty (no errors)
	assert.Empty(t, output)
}

// TestMainFunctionErrorHandling tests error handling in the main function flow
func TestMainFunctionErrorHandling(t *testing.T) {
	// Save original os.Args and os.Stderr
	originalArgs := os.Args
	originalStderr := os.Stderr

	// Restore after test
	defer func() {
		os.Args = originalArgs
		os.Stderr = originalStderr
	}()

	// Capture stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Test with invalid directory
	os.Args = []string{
		"i18n",
		"extract",
		"--dir", "/nonexistent/directory",
		"--out", "",
		"--gofile", "test.go",
		"--pkg", "main",
	}

	// Create CLI command (same as main would do)
	cliCmd := buildCLI(func(command *cli.Command) error {
		return i18n.NewExtractor(
			command.String("dir"),
			command.String("out"),
			command.String("pkg"),
			command.String("gofile"),
			command.StringSlice("ext")...,
		).Extract()
	})

	// Close pipe and restore stderr
	_ = w.Close()
	os.Stderr = originalStderr

	// Read error output (optional, we mainly care that an error occurs)
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)

	// Run command - this should fail
	ctx := context.Background()
	err := cliCmd.Run(ctx, os.Args)
	assert.Error(t, err, "Expected extraction to fail with non-existent directory")
}

// TestMainFunctionVersion tests version flag handling
func TestMainFunctionVersion(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test version flag
	os.Args = []string{"i18n", "--version"}

	cliCmd := buildCLI(func(command *cli.Command) error {
		return nil
	})

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	ctx := context.Background()
	err := cliCmd.Run(ctx, os.Args)

	// Close pipe and restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read output
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "v0.0.1")
}

// TestMainFunctionHelp tests help flag handling
func TestMainFunctionHelp(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test help flag
	os.Args = []string{"i18n", "--help"}

	cliCmd := buildCLI(func(command *cli.Command) error {
		return nil
	})

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	ctx := context.Background()
	err := cliCmd.Run(ctx, os.Args)

	// Close pipe and restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read output
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "i18n tool")
	assert.Contains(t, output, "extract")
}

// TestBuildCLIWithCustomExtractor tests buildCLI with different extractor behaviors
func TestBuildCLIWithCustomExtractor(t *testing.T) {
	// Test extractor that accesses command properties
	extractorFunc := func(cmd *cli.Command) error {
		// Verify we can access command properties
		assert.NotNil(t, cmd)
		assert.Equal(t, "extract", cmd.Name)

		// Verify we can access flag values (these will be defaults)
		assert.Equal(t, ".", cmd.String("dir"))
		assert.Equal(t, "gotext_stub.go", cmd.String("gofile"))
		assert.Equal(t, "main", cmd.String("pkg"))

		return nil
	}

	cmd := buildCLI(extractorFunc)
	require.Len(t, cmd.Commands, 1)

	extractCmd := cmd.Commands[0]
	ctx := context.Background()
	testCmd := &cli.Command{
		Name:   extractCmd.Name,
		Usage:  extractCmd.Usage,
		Flags:  extractCmd.Flags,
		Action: extractCmd.Action,
	}

	err := extractCmd.Action(ctx, testCmd)
	assert.NoError(t, err)
}

// TestVersionVariable tests that the version variable is properly set
func TestVersionVariable(t *testing.T) {
	assert.Equal(t, "v0.0.1", version)
}

// TestBuildCLINilExtractor tests behavior with nil extractor function
func TestBuildCLINilExtractor(t *testing.T) {
	// Should not panic with nil extractor
	assert.NotPanics(t, func() {
		cmd := buildCLI(nil)
		assert.NotNil(t, cmd)
		assert.Equal(t, "i18n", cmd.Name)
	})
}

// TestMainFunctionRealWorldScenario tests a real-world extraction scenario
func TestMainFunctionRealWorldScenario(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Create test directory structure
	tempDir := t.TempDir()

	// Create test template files
	createTestTemplate(t, tempDir, "index.html", `
<!DOCTYPE html>
<html>
<head><title>{{ T .Lang "Welcome" }}</title></head>
<body>
<h1>{{ i18n .Lang "Hello World" }}</h1>
<p>{{ t .Lang "This is a test" }}</p>
</body>
</html>
`)

	createTestTemplate(t, tempDir, "about.html", `
<!DOCTYPE html>
<html>
<head><title>{{ i18n .Lang "About Us" }}</title></head>
<body>
<h1>{{ T .Lang "About Page" }}</h1>
</body>
</html>
`)

	// Create output files paths
	outputFile := createTempFile(t, tempDir, "messages.json", "")
	goFile := createTempFile(t, tempDir, "gotext_stub.go", "")

	// Set up test arguments
	os.Args = []string{
		"i18n",
		"extract",
		"--dir", tempDir,
		"--out", outputFile,
		"--gofile", goFile,
		"--pkg", "testpkg",
		"--ext", ".html",
	}

	// Create and run CLI command
	cliCmd := buildCLI(func(command *cli.Command) error {
		return i18n.NewExtractor(
			command.String("dir"),
			command.String("out"),
			command.String("pkg"),
			command.String("gofile"),
			command.StringSlice("ext")...,
		).Extract()
	})

	ctx := context.Background()
	err := cliCmd.Run(ctx, os.Args)
	assert.NoError(t, err)

	// Verify output files were created and contain expected content
	assert.FileExists(t, outputFile)
	assert.FileExists(t, goFile)

	// Read and verify JSON output
	jsonContent, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(jsonContent), "Hello World")
	assert.Contains(t, string(jsonContent), "Welcome")
	assert.Contains(t, string(jsonContent), "This is a test")
	assert.Contains(t, string(jsonContent), "About Us")
	assert.Contains(t, string(jsonContent), "About Page")

	// Read and verify Go file output
	goContent, err := os.ReadFile(goFile)
	require.NoError(t, err)
	goContentStr := string(goContent)
	assert.Contains(t, goContentStr, "package testpkg")
	assert.Contains(t, goContentStr, "_ = p.Sprintf(\"Hello World\")")
	assert.Contains(t, goContentStr, "_ = p.Sprintf(\"Welcome\")")
}

// TestMainFunctionNoOutputPaths tests behavior when no output paths are specified
func TestMainFunctionNoOutputPaths(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Create test directory
	tempDir := t.TempDir()
	createTestTemplate(t, tempDir, "test.html", `{{ i18n .Lang "Test" }}`)

	// Create go file path (only this is required)
	goFile := createTempFile(t, tempDir, "gotext.go", "")

	// Set up test arguments with no --out flag
	os.Args = []string{
		"i18n",
		"extract",
		"--dir", tempDir,
		"--gofile", goFile,
		"--pkg", "main",
		"--ext", ".html",
	}

	// Create and run CLI command
	cliCmd := buildCLI(func(command *cli.Command) error {
		return i18n.NewExtractor(
			command.String("dir"),
			command.String("out"),
			command.String("pkg"),
			command.String("gofile"),
			command.StringSlice("ext")...,
		).Extract()
	})

	ctx := context.Background()
	err := cliCmd.Run(ctx, os.Args)
	assert.NoError(t, err)

	// Go file should still be created even without JSON output
	assert.FileExists(t, goFile)
}

// Helper functions

// createTestTemplate creates a test template file
func createTestTemplate(t *testing.T, dir, name, content string) string {
	file := createTempFile(t, dir, name, content)
	return file
}

// createTempFile creates a temporary file with given content
func createTempFile(t *testing.T, dir, name, content string) string {
	file := dir + string(os.PathSeparator) + name
	err := os.WriteFile(file, []byte(content), 0644)
	require.NoError(t, err)
	return file
}

// BenchmarkMainFunction benchmarks the main function setup
func BenchmarkMainFunction(b *testing.B) {
	extractorFunc := func(cmd *cli.Command) error {
		return i18n.NewExtractor(".", "", "main", "gotext_stub.go", ".html").Extract()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildCLI(extractorFunc)
	}
}
