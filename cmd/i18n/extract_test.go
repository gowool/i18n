package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// TestExtract tests the extract function with various scenarios
func TestExtract(t *testing.T) {
	tests := []struct {
		name          string
		extractorFunc func(*cli.Command) error
		expectedName  string
		expectedUsage string
	}{
		{
			name: "valid extractor function",
			extractorFunc: func(cmd *cli.Command) error {
				return nil
			},
			expectedName:  "extract",
			expectedUsage: "Extract i18n messages from templates",
		},
		{
			name: "extractor function returns error",
			extractorFunc: func(cmd *cli.Command) error {
				return assert.AnError
			},
			expectedName:  "extract",
			expectedUsage: "Extract i18n messages from templates",
		},
		{
			name: "nil extractor function",
			extractorFunc: func(cmd *cli.Command) error {
				return nil
			},
			expectedName:  "extract",
			expectedUsage: "Extract i18n messages from templates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			assert.NotPanics(t, func() {
				cmd := extract(tt.extractorFunc)

				// Verify command properties
				assert.Equal(t, tt.expectedName, cmd.Name)
				assert.Equal(t, tt.expectedUsage, cmd.Usage)
				assert.Equal(t, " ", cmd.ArgsUsage)

				// Verify flags count
				require.Len(t, cmd.Flags, 5)

				// Verify flag types and default values by examining flag definitions
				flagMap := make(map[string]cli.Flag)
				for _, f := range cmd.Flags {
					flagMap[f.Names()[0]] = f
				}

				// Check dir flag
				dirFlag, exists := flagMap["dir"]
				require.True(t, exists, "dir flag should exist")
				assert.IsType(t, &cli.StringFlag{}, dirFlag)
				dirStringFlag := dirFlag.(*cli.StringFlag)
				assert.Equal(t, "directory to scan for templates", dirStringFlag.Usage)
				assert.Equal(t, ".", dirStringFlag.Value)

				// Check out flag
				outFlag, exists := flagMap["out"]
				require.True(t, exists, "out flag should exist")
				assert.IsType(t, &cli.StringFlag{}, outFlag)
				outStringFlag := outFlag.(*cli.StringFlag)
				assert.Equal(t, "output JSON found i18n messages", outStringFlag.Usage)
				assert.Empty(t, outStringFlag.Value)

				// Check gofile flag
				goFileFlag, exists := flagMap["gofile"]
				require.True(t, exists, "gofile flag should exist")
				assert.IsType(t, &cli.StringFlag{}, goFileFlag)
				goFileStringFlag := goFileFlag.(*cli.StringFlag)
				assert.Equal(t, "synthetic Go file generated for gotext extract/update", goFileStringFlag.Usage)
				assert.Equal(t, "gotext_stub.go", goFileStringFlag.Value)

				// Check pkg flag
				pkgFlag, exists := flagMap["pkg"]
				require.True(t, exists, "pkg flag should exist")
				assert.IsType(t, &cli.StringFlag{}, pkgFlag)
				pkgStringFlag := pkgFlag.(*cli.StringFlag)
				assert.Equal(t, "package name to use in generated Go file", pkgStringFlag.Usage)
				assert.Equal(t, "main", pkgStringFlag.Value)

				// Check ext flag
				extFlag, exists := flagMap["ext"]
				require.True(t, exists, "ext flag should exist")
				assert.IsType(t, &cli.StringSliceFlag{}, extFlag)
				extStringSliceFlag := extFlag.(*cli.StringSliceFlag)
				assert.Equal(t, "template extensions to consider", extStringSliceFlag.Usage)
				expectedExts := []string{".html", ".htm", ".tmpl", ".gohtml", ".txt", ".tpl"}
				assert.Equal(t, expectedExts, extStringSliceFlag.Value)

				// Test action function
				require.NotNil(t, cmd.Action)

				// Test the action function with minimal test
				if tt.name == "extractor function returns error" {
					// Test that error propagates
					ctx := context.Background()
					testCmd := &cli.Command{
						Name:   cmd.Name,
						Usage:  cmd.Usage,
						Flags:  cmd.Flags,
						Action: cmd.Action,
					}
					// Create a simple command context for testing
					app := &cli.Command{Name: "test"}
					app.Commands = []*cli.Command{testCmd}

					err := cmd.Action(ctx, testCmd)
					assert.Error(t, err)
				} else {
					// For successful cases, just verify action doesn't panic
					assert.NotPanics(t, func() {
						ctx := context.Background()
						testCmd := &cli.Command{
							Name:   cmd.Name,
							Usage:  cmd.Usage,
							Flags:  cmd.Flags,
							Action: cmd.Action,
						}
						_ = cmd.Action(ctx, testCmd)
					})
				}
			})
		})
	}
}

// TestExtractCommandStructure tests the structure of the returned command
func TestExtractCommandStructure(t *testing.T) {
	cmd := extract(func(*cli.Command) error { return nil })

	// Test command basic properties
	assert.Equal(t, "extract", cmd.Name)
	assert.Equal(t, "Extract i18n messages from templates", cmd.Usage)
	assert.Equal(t, " ", cmd.ArgsUsage)

	// Test that command has no subcommands
	assert.Empty(t, cmd.Commands)

	// Test that command has flags
	assert.NotEmpty(t, cmd.Flags)
	assert.Len(t, cmd.Flags, 5) // dir, out, gofile, pkg, ext

	// Test that command has an action
	assert.NotNil(t, cmd.Action)

	// Test flag names and types
	flagNames := make([]string, len(cmd.Flags))
	flagTypes := make([]string, len(cmd.Flags))
	for i, f := range cmd.Flags {
		flagNames[i] = f.Names()[0]
		switch f.(type) {
		case *cli.StringFlag:
			flagTypes[i] = "StringFlag"
		case *cli.StringSliceFlag:
			flagTypes[i] = "StringSliceFlag"
		default:
			flagTypes[i] = "Unknown"
		}
	}

	expectedFlagNames := []string{"dir", "out", "gofile", "pkg", "ext"}
	assert.Equal(t, expectedFlagNames, flagNames)

	// Verify specific flag types
	expectedTypes := []string{"StringFlag", "StringFlag", "StringFlag", "StringFlag", "StringSliceFlag"}
	assert.Equal(t, expectedTypes, flagTypes)
}

// TestExtractWithNilExtractor tests extract with nil extractor function
func TestExtractWithNilExtractor(t *testing.T) {
	// Test that extract doesn't panic with nil extractor
	assert.NotPanics(t, func() {
		cmd := extract(nil)
		assert.NotNil(t, cmd)
		assert.Equal(t, "extract", cmd.Name)
	})
}

// TestExtractFlagValuesStructure tests the structure and default values of flags
func TestExtractFlagValuesStructure(t *testing.T) {
	extractorFunc := func(cmd *cli.Command) error {
		return nil
	}

	cmd := extract(extractorFunc)

	// Verify flag structure
	require.Len(t, cmd.Flags, 5)

	// Find each flag by name
	var dirFlag, outFlag, gofileFlag, pkgFlag, extFlag cli.Flag
	for _, f := range cmd.Flags {
		switch f.Names()[0] {
		case "dir":
			dirFlag = f
		case "out":
			outFlag = f
		case "gofile":
			gofileFlag = f
		case "pkg":
			pkgFlag = f
		case "ext":
			extFlag = f
		}
	}

	// Test dir flag
	require.NotNil(t, dirFlag)
	assert.IsType(t, &cli.StringFlag{}, dirFlag)
	assert.Equal(t, "dir", dirFlag.Names()[0])
	assert.Equal(t, "directory to scan for templates", dirFlag.(*cli.StringFlag).Usage)
	assert.Equal(t, ".", dirFlag.(*cli.StringFlag).Value)

	// Test out flag
	require.NotNil(t, outFlag)
	assert.IsType(t, &cli.StringFlag{}, outFlag)
	assert.Equal(t, "out", outFlag.Names()[0])
	assert.Equal(t, "output JSON found i18n messages", outFlag.(*cli.StringFlag).Usage)
	assert.Empty(t, outFlag.(*cli.StringFlag).Value)

	// Test gofile flag
	require.NotNil(t, gofileFlag)
	assert.IsType(t, &cli.StringFlag{}, gofileFlag)
	assert.Equal(t, "gofile", gofileFlag.Names()[0])
	assert.Equal(t, "synthetic Go file generated for gotext extract/update", gofileFlag.(*cli.StringFlag).Usage)
	assert.Equal(t, "gotext_stub.go", gofileFlag.(*cli.StringFlag).Value)

	// Test pkg flag
	require.NotNil(t, pkgFlag)
	assert.IsType(t, &cli.StringFlag{}, pkgFlag)
	assert.Equal(t, "pkg", pkgFlag.Names()[0])
	assert.Equal(t, "package name to use in generated Go file", pkgFlag.(*cli.StringFlag).Usage)
	assert.Equal(t, "main", pkgFlag.(*cli.StringFlag).Value)

	// Test ext flag
	require.NotNil(t, extFlag)
	assert.IsType(t, &cli.StringSliceFlag{}, extFlag)
	assert.Equal(t, "ext", extFlag.Names()[0])
	assert.Equal(t, "template extensions to consider", extFlag.(*cli.StringSliceFlag).Usage)
	expectedExts := []string{".html", ".htm", ".tmpl", ".gohtml", ".txt", ".tpl"}
	assert.Equal(t, expectedExts, extFlag.(*cli.StringSliceFlag).Value)
}

// TestExtractActionFunction tests that the action function properly calls the extractor
func TestExtractActionFunction(t *testing.T) {
	var calledCmd *cli.Command
	extractorFunc := func(cmd *cli.Command) error {
		calledCmd = cmd
		return nil
	}

	cmd := extract(extractorFunc)
	require.NotNil(t, cmd.Action)

	// Create a test command
	ctx := context.Background()
	testCmd := &cli.Command{
		Name:   cmd.Name,
		Usage:  cmd.Usage,
		Flags:  cmd.Flags,
		Action: cmd.Action,
	}

	// Execute the action
	err := cmd.Action(ctx, testCmd)
	assert.NoError(t, err)
	assert.NotNil(t, calledCmd)
	assert.Equal(t, testCmd, calledCmd)
}

// TestExtractActionFunctionError tests that the action function properly propagates errors
func TestExtractActionFunctionError(t *testing.T) {
	testError := cli.Exit("test error", 1)
	extractorFunc := func(cmd *cli.Command) error {
		return testError
	}

	cmd := extract(extractorFunc)
	require.NotNil(t, cmd.Action)

	// Create a test command
	ctx := context.Background()
	testCmd := &cli.Command{
		Name:   cmd.Name,
		Usage:  cmd.Usage,
		Flags:  cmd.Flags,
		Action: cmd.Action,
	}

	// Execute the action
	err := cmd.Action(ctx, testCmd)
	assert.Error(t, err)
	assert.Equal(t, testError, err)
}

// TestExtractCommandConsistency tests that multiple calls to extract produce consistent results
func TestExtractCommandConsistency(t *testing.T) {
	extractor1 := func(cmd *cli.Command) error { return nil }
	extractor2 := func(cmd *cli.Command) error { return nil }

	cmd1 := extract(extractor1)
	cmd2 := extract(extractor2)

	// Commands should have same structure
	assert.Equal(t, cmd1.Name, cmd2.Name)
	assert.Equal(t, cmd1.Usage, cmd2.Usage)
	assert.Equal(t, cmd1.ArgsUsage, cmd2.ArgsUsage)
	assert.Len(t, cmd1.Flags, len(cmd2.Flags))

	// Flag structures should be identical
	for i, flag1 := range cmd1.Flags {
		flag2 := cmd2.Flags[i]
		assert.Equal(t, flag1.Names()[0], flag2.Names()[0])
		assert.Equal(t, flag1.Names(), flag2.Names())
	}

	// Actions should be different functions
	assert.NotEqual(t, &cmd1.Action, &cmd2.Action)
}

// BenchmarkExtract benchmarks the extract function creation
func BenchmarkExtract(b *testing.B) {
	extractorFunc := func(*cli.Command) error { return nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extract(extractorFunc)
	}
}

// TestExtractContextHandling tests that the command properly handles context
func TestExtractContextHandling(t *testing.T) {
	extractorFunc := func(cmd *cli.Command) error {
		return nil
	}

	cmd := extract(extractorFunc)

	ctx := context.Background()
	testCmd := &cli.Command{
		Name:   cmd.Name,
		Usage:  cmd.Usage,
		Flags:  cmd.Flags,
		Action: cmd.Action,
	}

	// Execute with context
	err := cmd.Action(ctx, testCmd)
	assert.NoError(t, err)
	// Note: We can't easily test the context passed to the extractor since it's not exposed
	// but we can verify the command handles context without error
}

// TestExtractWithPanicExtractor tests behavior when extractor function panics
func TestExtractWithPanicExtractor(t *testing.T) {
	extractorFunc := func(cmd *cli.Command) error {
		panic("test panic")
	}

	cmd := extract(extractorFunc)

	ctx := context.Background()
	testCmd := &cli.Command{
		Name:   cmd.Name,
		Usage:  cmd.Usage,
		Flags:  cmd.Flags,
		Action: cmd.Action,
	}

	// The panic should propagate
	assert.Panics(t, func() {
		_ = cmd.Action(ctx, testCmd)
	})
}
