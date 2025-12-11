# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Building and Running
- `go run ./cmd/i18n` - Run the i18n CLI tool
- `go build ./cmd/i18n` - Build the i18n binary

### Testing
- `go test ./...` - Run all tests
- `go test -v ./...` - Run tests with verbose output
- `go test -race ./...` - Run tests with race detector

### Code Quality
- `go fmt ./...` - Format code
- `go vet ./...` - Static analysis
- `golangci-lint run -v --timeout=5m --build-tags=race --output.code-climate.path gl-code-quality-report.json` - Lint code

### Message Extraction
- `i18n extract` - Extract i18n messages from templates in current directory
- `i18n extract --dir ./templates --out messages.json` - Extract from specific directory to JSON file
- `i18n extract --ext .html --ext .tmpl` - Extract from specific file extensions
- `gotext extract/update` - Process the generated `gotext_stub.go` file for i18n

## Architecture

This is a Go internationalization (i18n) library with two main components:

### Core Library (`i18n.go`)
- **Translation Functions**: Provides `T()` function and template helpers (`L`, `T`, `t`, `i18n`)
- **Language Management**: Handles language tags and fallback logic using `golang.org/x/text/language`
- **Printer Caching**: Thread-safe caching of message printers using `sync.Map`
- **Template Integration**: Template function map for Go templates integration

### Message Extractor (`extractor.go`)
- **Template Parsing**: Scans template files for i18n function calls using regex patterns
- **Message Collection**: Extracts message strings and their source positions
- **Dual Output**: Generates both JSON message files and synthetic Go files for `gotext` tool
- **File Extension Support**: Configurable template file extensions (.html, .htm, .tmpl, .gohtml, .txt, .tpl)

### CLI Tool (`cmd/i18n/main.go`)
- **urfave/cli/v3**: Modern CLI framework with command structure
- **Extract Command**: Primary command for message extraction with configurable flags
- **Integration**: Wraps the core extractor functionality for command-line use

## Key Dependencies
- `golang.org/x/text/language` - Language tag parsing and matching
- `golang.org/x/text/message` - Message formatting and printer functionality
- `github.com/urfave/cli/v3` - CLI framework for the command-line tool
- `github.com/stretchr/testify` - Testing framework

## Template Integration
The library integrates with Go templates through the `FuncMap` which provides:
- `L`: Parse language tags
- `T/t/i18n`: Translation functions (all aliases to the same `trans` function)

Templates use the syntax: `{{ i18n .Lang "message_key" "arg1" "arg2" }}` or `{{ T .Lang "message_key" }}`
