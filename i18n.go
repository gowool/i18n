package i18n

import (
	"fmt"
	"html/template"
	"sync"
	"sync/atomic"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var FuncMap = template.FuncMap{
	"L":    language.MustParse,
	"T":    trans,
	"t":    trans,
	"i18n": trans,
}

var (
	fallback atomic.Value
	printers sync.Map
)

func Fallback() language.Tag {
	if tag := fallback.Load(); tag != nil {
		return tag.(language.Tag)
	}
	return language.English
}

func SetFallback(tag language.Tag) {
	fallback.Store(tag)
}

func Printer(tag language.Tag) *message.Printer {
	if v, ok := printers.Load(tag); ok {
		return v.(*message.Printer)
	}

	printer := message.NewPrinter(tag)
	printers.Store(tag, printer)

	return printer
}

func SetPrinter(tag language.Tag, printer *message.Printer) {
	printers.Store(tag, printer)
}

func T(tag language.Tag, key message.Reference, a ...any) string {
	return Printer(tag).Sprintf(key, a...)
}

func trans(lang any, key message.Reference, a ...any) string {
	var tag language.Tag
	switch lang := lang.(type) {
	case language.Tag:
		tag = lang
	case *language.Tag:
		tag = *lang
	case string:
		tag = language.MustParse(lang)
	case *string:
		tag = language.MustParse(*lang)
	case fmt.Stringer:
		tag = language.MustParse(lang.String())
	default:
		panic("invalid language tag")
	}

	return T(tag, key, a...)
}
