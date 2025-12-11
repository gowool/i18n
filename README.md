# I18N

[![Go Reference](https://pkg.go.dev/badge/github.com/gowool/i18n.svg)](https://pkg.go.dev/github.com/gowool/i18n)
[![Go Report Card](https://goreportcard.com/badge/github.com/gowool/i18n)](https://goreportcard.com/report/github.com/gowool/i18n)
[![codecov](https://codecov.io/github/gowool/i18n/graph/badge.svg?token=JJX69WTRMH)](https://codecov.io/github/gowool/i18n)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://github.com/gowool/i18n/blob/main/LICENSE)

A Go internationalization library that provides template integration and message extraction capabilities. This library offers a simple way to add i18n support to Go applications using Go templates, with automatic message extraction and integration with Go's standard `golang.org/x/text` package.

## Installation

```bash
go get github.com/gowool/i18n

# install tool
go get -tool github.com/gowool/i18n
go get -tool golang.org/x/text/cmd/gotext
```

## Extractor

```go
package locales

//go:generate go tool i18n extract --dir ./themes --pkg locales --out locales/messages.json --gofile locales/gotext_stub.go --ext .html --ext .gohtml
//go:generate go tool gotext -srclang=en-US update -out=catalog.go -lang=en-US,it-IT,fr-FR,de-DE github.com/dummy/example/locales
```

```bash
go generate ./...
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
