package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ExtractorTestSuite contains all tests for the extractor functionality
type ExtractorTestSuite struct {
	suite.Suite
	tempDir string
}

// SetupSuite creates a temporary directory for test files
func (suite *ExtractorTestSuite) SetupSuite() {
	tempDir, err := os.MkdirTemp("", "i18n_extractor_test")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir
}

// TearDownSuite cleans up the temporary directory
func (suite *ExtractorTestSuite) TearDownSuite() {
	_ = os.RemoveAll(suite.tempDir)
}

// SetupTest runs before each test
func (suite *ExtractorTestSuite) SetupTest() {
	// Clean temp directory before each test
	files, _ := filepath.Glob(filepath.Join(suite.tempDir, "*"))
	for _, file := range files {
		_ = os.RemoveAll(file)
	}
}

// TestNewExtractor tests the NewExtractor constructor function
func (suite *ExtractorTestSuite) TestNewExtractor() {
	tests := []struct {
		name     string
		dir      string
		out      string
		pkg      string
		goFile   string
		exts     []string
		expected *Extractor
	}{
		{
			name:   "basic extractor",
			dir:    "/tmp",
			out:    "out.json",
			pkg:    "main",
			goFile: "gotext_stub.go",
			exts:   []string{".html", ".tmpl"},
			expected: &Extractor{
				dir:    "/tmp",
				out:    "out.json",
				pkg:    "main",
				goFile: "gotext_stub.go",
				exts:   map[string]struct{}{".html": {}, ".tmpl": {}},
			},
		},
		{
			name:   "no extensions",
			dir:    "/tmp",
			out:    "",
			pkg:    "",
			goFile: "stub.go",
			exts:   []string{},
			expected: &Extractor{
				dir:    "/tmp",
				out:    "",
				pkg:    "",
				goFile: "stub.go",
				exts:   map[string]struct{}{},
			},
		},
		{
			name:   "single extension",
			dir:    "templates",
			out:    "messages.json",
			pkg:    "mypackage",
			goFile: "extract.go",
			exts:   []string{".gohtml"},
			expected: &Extractor{
				dir:    "templates",
				out:    "messages.json",
				pkg:    "mypackage",
				goFile: "extract.go",
				exts:   map[string]struct{}{".gohtml": {}},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			extractor := NewExtractor(tt.dir, tt.out, tt.pkg, tt.goFile, tt.exts...)
			assert.Equal(suite.T(), tt.expected, extractor)
		})
	}
}

// TestIsTemplate tests the isTemplate method
func (suite *ExtractorTestSuite) TestIsTemplate() {
	extractor := NewExtractor("", "", "", "", ".html", ".tmpl", ".gohtml")

	tests := []struct {
		path     string
		expected bool
	}{
		{"/path/to/template.html", true},
		{"/path/to/template.tmpl", true},
		{"/path/to/template.gohtml", true},
		{"/path/to/template.htm", false},
		{"/path/to/template.txt", false},
		{"/path/to/template.HTML", true}, // case insensitive
		{"/path/to/template.TMPL", true}, // case insensitive
		{"template.html", true},
		{"template", false},
		{"", false},
	}

	for _, tt := range tests {
		suite.Run(tt.path, func() {
			result := extractor.isTemplate(tt.path)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestExtractFromContent tests the extractFromContent function
func (suite *ExtractorTestSuite) TestExtractFromContent() {
	tests := []struct {
		name     string
		content  string
		relPath  string
		expected map[string]*Message
	}{
		{
			name:     "simple i18n call",
			content:  `{{ i18n .Lang "Hello world" }}`,
			relPath:  "template.html",
			expected: map[string]*Message{"Hello world": {ID: "Hello world", Positions: []string{"template.html:1:10"}}},
		},
		{
			name:     "T function call",
			content:  `<h1>{{ T .Lang "Page title" }}</h1>`,
			relPath:  "index.html",
			expected: map[string]*Message{"Page title": {ID: "Page title", Positions: []string{"index.html:1:11"}}},
		},
		{
			name:    "multiple i18n calls",
			content: `{{ i18n .Lang "Hello" }} {{ T .Lang "World" }}`,
			relPath: "multi.html",
			expected: map[string]*Message{
				"Hello": {ID: "Hello", Positions: []string{"multi.html:1:10"}},
				"World": {ID: "World", Positions: []string{"multi.html:1:32"}},
			},
		},
		{
			name:    "duplicate messages",
			content: `{{ i18n .Lang "Save" }} {{ T .Lang "Save" }}`,
			relPath: "dup.html",
			expected: map[string]*Message{
				"Save": {ID: "Save", Positions: []string{"dup.html:1:10", "dup.html:1:31"}},
			},
		},
		{
			name:     "escaped quotes",
			content:  `{{ i18n .Lang "Say \"Hello\"" }}`,
			relPath:  "escape.html",
			expected: map[string]*Message{`Say "Hello"`: {ID: `Say "Hello"`, Positions: []string{"escape.html:1:10"}}},
		},
		{
			name:     "no i18n calls",
			content:  `<div>Regular HTML content</div>`,
			relPath:  "plain.html",
			expected: map[string]*Message{},
		},
		{
			name:     "incomplete i18n syntax",
			content:  `{{ i18n .Lang "Incomplete message"`,
			relPath:  "incomplete.html",
			expected: map[string]*Message{},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := make(map[string]*Message)
			extractFromContent(result, []byte(tt.content), tt.relPath)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestPositionFor tests the positionFor function
func (suite *ExtractorTestSuite) TestPositionFor() {
	tests := []struct {
		name     string
		content  string
		pos      int
		relPath  string
		expected string
	}{
		{
			name:     "first line first character",
			content:  "Hello world",
			pos:      0,
			relPath:  "test.html",
			expected: "test.html:?:?",
		},
		{
			name:     "first line middle",
			content:  "Hello world",
			pos:      5,
			relPath:  "test.html",
			expected: "test.html:1:6",
		},
		{
			name:     "second line",
			content:  "Hello\nWorld",
			pos:      7,
			relPath:  "test.html",
			expected: "test.html:2:2",
		},
		{
			name:     "multiple lines",
			content:  "Line1\nLine2\nLine3",
			pos:      15,
			relPath:  "test.html",
			expected: "test.html:3:4",
		},
		{
			name:     "negative position",
			content:  "Hello",
			pos:      -1,
			relPath:  "test.html",
			expected: "test.html:?:?",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := positionFor([]byte(tt.content), tt.pos, tt.relPath)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestStrconvQuote tests the strconvQuote function
func (suite *ExtractorTestSuite) TestStrconvQuote() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "string with quotes",
			input:    `say "hello"`,
			expected: `"say \"hello\""`,
		},
		{
			name:     "string with backslash",
			input:    `path\to\file`,
			expected: `"path\\to\\file"`,
		},
		{
			name:     "string with newline",
			input:    "line1\nline2",
			expected: `"line1\nline2"`,
		},
		{
			name:     "string with tab",
			input:    "col1\tcol2",
			expected: `"col1\tcol2"`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: `""`,
		},
		{
			name:     "complex string with newline and tab",
			input:    "Hello\n\tWorld",
			expected: `"Hello\n\tWorld"`,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := strconvQuote(tt.input)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestSanitizePkgName tests the sanitizePkgName function
func (suite *ExtractorTestSuite) TestSanitizePkgName() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid package name",
			input:    "mypackage",
			expected: "mypackage",
		},
		{
			name:     "package with numbers",
			input:    "pkg123",
			expected: "pkg123",
		},
		{
			name:     "package with underscore",
			input:    "my_package",
			expected: "my_package",
		},
		{
			name:     "mixed case",
			input:    "MyPackage",
			expected: "MyPackage",
		},
		{
			name:     "starts with number",
			input:    "123package",
			expected: "23package",
		},
		{
			name:     "invalid characters only",
			input:    "123!@#",
			expected: "23",
		},
		{
			name:     "mixed valid and invalid",
			input:    "my-package.name",
			expected: "mypackagename",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "main",
		},
		{
			name:     "space at start",
			input:    " mypackage",
			expected: "mypackage",
		},
		{
			name:     "number in middle",
			input:    "my123package",
			expected: "my123package",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := sanitizePkgName(tt.input)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestExtract tests the full extraction process
func (suite *ExtractorTestSuite) TestExtract() {
	// Create test template files
	template1 := `<!DOCTYPE html>
<html>
<head><title>{{ T .Lang "Welcome" }}</title></head>
<body>
<h1>{{ i18n .Lang "Hello World" }}</h1>
<p>{{ t .Lang "This is a test" }}</p>
</body>
</html>`

	template2 := `{{ i18n .Lang "Save" }}
{{ i18n .Lang "Cancel" }}
{{ i18n .Lang "Welcome" }}` // duplicate message

	// Write test files
	file1 := filepath.Join(suite.tempDir, "template1.html")
	file2 := filepath.Join(suite.tempDir, "template2.tmpl")

	err := os.WriteFile(file1, []byte(template1), 0644)
	require.NoError(suite.T(), err)

	err = os.WriteFile(file2, []byte(template2), 0644)
	require.NoError(suite.T(), err)

	// Test extraction
	extractor := NewExtractor(suite.tempDir, "", "testpkg", filepath.Join(suite.tempDir, "gotext.go"), ".html", ".tmpl")
	messages, err := extractor.extract()
	require.NoError(suite.T(), err)

	// Verify extracted messages
	expectedMessages := []*Message{
		{ID: "Cancel", Positions: []string{"template2.tmpl:2:10"}},
		{ID: "Hello World", Positions: []string{"template1.html:5:14"}},
		{ID: "Save", Positions: []string{"template2.tmpl:1:10"}},
		{ID: "This is a test", Positions: []string{"template1.html:6:10"}},
		{ID: "Welcome", Positions: []string{"template1.html:3:20", "template2.tmpl:3:10"}},
	}

	assert.Equal(suite.T(), len(expectedMessages), len(messages))
	for i, msg := range messages {
		assert.Equal(suite.T(), expectedMessages[i].ID, msg.ID)
		assert.Equal(suite.T(), expectedMessages[i].Positions, msg.Positions)
	}
}

// TestSaveMessages tests the saveMessages functionality
func (suite *ExtractorTestSuite) TestSaveMessages() {
	messages := []*Message{
		{ID: "Hello", Positions: []string{"file1.html:1:1"}},
		{ID: "World", Positions: []string{"file2.html:2:2"}},
	}

	outputFile := filepath.Join(suite.tempDir, "messages.json")
	extractor := NewExtractor("", outputFile, "test", "test.go")

	err := extractor.saveMessages(messages)
	require.NoError(suite.T(), err)

	// Verify file was created
	_, err = os.Stat(outputFile)
	assert.NoError(suite.T(), err)

	// Verify content
	data, err := os.ReadFile(outputFile)
	require.NoError(suite.T(), err)

	var outputJSON OutputJSON
	err = json.Unmarshal(data, &outputJSON)
	require.NoError(suite.T(), err)

	assert.Len(suite.T(), outputJSON.Messages, 2)
	assert.Equal(suite.T(), "Hello", outputJSON.Messages[0].ID)
	assert.Equal(suite.T(), "World", outputJSON.Messages[1].ID)
}

// TestSaveMessagesEmptyOutput tests saveMessages with empty output path
func (suite *ExtractorTestSuite) TestSaveMessagesEmptyOutput() {
	messages := []*Message{{ID: "Hello", Positions: []string{"file1.html:1:1"}}}
	extractor := NewExtractor("", "", "test", "test.go")

	err := extractor.saveMessages(messages)
	assert.NoError(suite.T(), err)
}

// TestSaveGoFile tests the saveGoFile functionality
func (suite *ExtractorTestSuite) TestSaveGoFile() {
	messages := []*Message{
		{ID: "Hello", Positions: []string{"file1.html:1:1"}},
		{ID: `Say "Hello"`, Positions: []string{"file2.html:2:2"}}, // test with quotes
	}

	goFile := filepath.Join(suite.tempDir, "gotext_stub.go")
	extractor := NewExtractor("", "", "testpkg", goFile)

	err := extractor.saveGoFile(messages)
	require.NoError(suite.T(), err)

	// Verify file was created
	_, err = os.Stat(goFile)
	assert.NoError(suite.T(), err)

	// Verify content
	data, err := os.ReadFile(goFile)
	require.NoError(suite.T(), err)

	content := string(data)
	assert.Contains(suite.T(), content, "// Code generated by i18n-extract. DO NOT EDIT.")
	assert.Contains(suite.T(), content, "package testpkg")
	assert.Contains(suite.T(), content, "_ = p.Sprintf(\"Hello\")")
	assert.Contains(suite.T(), content, "_ = p.Sprintf(\"Say \\\"Hello\\\"\")")
	assert.Contains(suite.T(), content, "func _i18n_extract() {")
	assert.Contains(suite.T(), content, "}")
}

// TestBuildSyntheticGo tests the buildSyntheticGo method
func (suite *ExtractorTestSuite) TestBuildSyntheticGo() {
	tests := []struct {
		name     string
		pkg      string
		messages []*Message
		expected []string
	}{
		{
			name: "single message",
			pkg:  "main",
			messages: []*Message{
				{ID: "Hello", Positions: []string{"file.html:1:1"}},
			},
			expected: []string{"package main", "_ = p.Sprintf(\"Hello\")"},
		},
		{
			name: "multiple messages",
			pkg:  "mypkg",
			messages: []*Message{
				{ID: "Hello", Positions: []string{"file.html:1:1"}},
				{ID: "World", Positions: []string{"file2.html:2:2"}},
			},
			expected: []string{"package mypkg", "_ = p.Sprintf(\"Hello\")", "_ = p.Sprintf(\"World\")"},
		},
		{
			name:     "no messages",
			pkg:      "empty",
			messages: []*Message{},
			expected: []string{"package empty"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			extractor := NewExtractor("", "", tt.pkg, "test.go")
			result, err := extractor.buildSyntheticGo(tt.messages)
			require.NoError(suite.T(), err)

			content := string(result)
			for _, expected := range tt.expected {
				assert.Contains(suite.T(), content, expected)
			}
		})
	}
}

// TestExtractError tests error handling in extract method
func (suite *ExtractorTestSuite) TestExtractError() {
	// Create a directory that doesn't exist
	extractor := NewExtractor("/nonexistent/directory", "", "test", "test.go", ".html")

	messages, err := extractor.extract()
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), messages)
}

// TestSaveMessagesError tests error handling in saveMessages
func (suite *ExtractorTestSuite) TestSaveMessagesError() {
	// Try to write to a directory that doesn't exist
	messages := []*Message{{ID: "Hello", Positions: []string{"file.html:1:1"}}}
	extractor := NewExtractor("", "/nonexistent/path/messages.json", "test", "test.go")

	err := extractor.saveMessages(messages)
	assert.Error(suite.T(), err)
}

// TestSaveGoFileError tests error handling in saveGoFile
func (suite *ExtractorTestSuite) TestSaveGoFileError() {
	messages := []*Message{{ID: "Hello", Positions: []string{"file.html:1:1"}}}
	extractor := NewExtractor("", "/nonexistent/path/gotext.go", "test", "/nonexistent/path/gotext.go")

	err := extractor.saveGoFile(messages)
	assert.Error(suite.T(), err)
}

// TestExtractFullWorkflow tests the complete Extract workflow
func (suite *ExtractorTestSuite) TestExtractFullWorkflow() {
	// Create test template - using simpler messages to avoid regex issues with nested braces
	template := `<!DOCTYPE html>
<html>
<head><title>{{ i18n .Lang "Welcome Page" }}</title></head>
<body>
<h1>{{ T .Lang "Hello User" }}</h1>
<p>{{ t .Lang "This is test message" }}</p>
</body>
</html>`

	templateFile := filepath.Join(suite.tempDir, "test.html")
	err := os.WriteFile(templateFile, []byte(template), 0644)
	require.NoError(suite.T(), err)

	// Configure extractor
	jsonOutput := filepath.Join(suite.tempDir, "messages.json")
	goOutput := filepath.Join(suite.tempDir, "gotext_stub.go")
	extractor := NewExtractor(suite.tempDir, jsonOutput, "testpkg", goOutput, ".html")

	// Run full extraction
	err = extractor.Extract()
	require.NoError(suite.T(), err)

	// Verify JSON file exists and contains expected data
	_, err = os.Stat(jsonOutput)
	assert.NoError(suite.T(), err)

	jsonData, err := os.ReadFile(jsonOutput)
	require.NoError(suite.T(), err)

	var outputJSON OutputJSON
	err = json.Unmarshal(jsonData, &outputJSON)
	require.NoError(suite.T(), err)

	expectedMessages := []string{
		"Hello User",
		"This is test message",
		"Welcome Page",
	}

	assert.Len(suite.T(), outputJSON.Messages, len(expectedMessages))
	for i, msg := range outputJSON.Messages {
		assert.Equal(suite.T(), expectedMessages[i], msg.ID)
		assert.Len(suite.T(), msg.Positions, 1)
		assert.Contains(suite.T(), msg.Positions[0], "test.html:")
	}

	// Verify Go file exists and contains expected code
	_, err = os.Stat(goOutput)
	assert.NoError(suite.T(), err)

	goData, err := os.ReadFile(goOutput)
	require.NoError(suite.T(), err)

	goContent := string(goData)
	assert.Contains(suite.T(), goContent, "package testpkg")
	for _, expectedMsg := range expectedMessages {
		quoted := strconvQuote(expectedMsg)
		assert.Contains(suite.T(), goContent, "_ = p.Sprintf("+quoted+")")
	}
}

// TestExtractWithNoOutputPaths tests extraction with no output files
func (suite *ExtractorTestSuite) TestExtractWithNoOutputPaths() {
	// Create test template
	template := `{{ i18n .Lang "Test message" }}`
	templateFile := filepath.Join(suite.tempDir, "test.html")
	err := os.WriteFile(templateFile, []byte(template), 0644)
	require.NoError(suite.T(), err)

	// Configure extractor with empty output paths
	extractor := NewExtractor(suite.tempDir, "", "testpkg", filepath.Join(suite.tempDir, "gotext.go"), ".html")

	// This should not fail even with empty JSON output
	err = extractor.Extract()
	assert.NoError(suite.T(), err)

	// Go file should still be created
	_, err = os.Stat(filepath.Join(suite.tempDir, "gotext.go"))
	assert.NoError(suite.T(), err)
}

// Run the test suite
func TestExtractorTestSuite(t *testing.T) {
	suite.Run(t, new(ExtractorTestSuite))
}
