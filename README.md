# I18N

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
