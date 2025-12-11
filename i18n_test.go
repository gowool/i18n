package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestFallback(t *testing.T) {
	t.Run("Default fallback is English", func(t *testing.T) {
		fallback := Fallback()
		assert.Equal(t, language.English, fallback)
	})

	t.Run("Set and get fallback", func(t *testing.T) {
		original := Fallback()
		defer SetFallback(original)

		newFallback := language.French
		SetFallback(newFallback)

		retrieved := Fallback()
		assert.Equal(t, newFallback, retrieved)
	})

	t.Run("Concurrent fallback access", func(t *testing.T) {
		original := Fallback()
		defer SetFallback(original)

		newFallback := language.German
		SetFallback(newFallback)

		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				retrieved := Fallback()
				assert.Equal(t, newFallback, retrieved)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestPrinter(t *testing.T) {
	t.Run("Create and retrieve printer", func(t *testing.T) {
		tag := language.Spanish
		printer1 := Printer(tag)
		assert.NotNil(t, printer1)

		printer2 := Printer(tag)
		assert.Same(t, printer1, printer2, "Should return the same cached printer")
	})

	t.Run("Custom printer setting", func(t *testing.T) {
		tag := language.Italian
		customPrinter := message.NewPrinter(tag)

		SetPrinter(tag, customPrinter)
		retrieved := Printer(tag)

		assert.Same(t, customPrinter, retrieved)
	})

	t.Run("Different tags return different printers", func(t *testing.T) {
		tag1 := language.Japanese
		tag2 := language.Korean

		printer1 := Printer(tag1)
		printer2 := Printer(tag2)

		assert.NotSame(t, printer1, printer2)
		assert.NotNil(t, printer1)
		assert.NotNil(t, printer2)
	})

	t.Run("Concurrent printer access", func(t *testing.T) {
		tag := language.Portuguese
		done := make(chan bool, 20)

		for i := 0; i < 10; i++ {
			go func() {
				printer := Printer(tag)
				assert.NotNil(t, printer)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			go func() {
				customPrinter := message.NewPrinter(tag)
				SetPrinter(tag, customPrinter)
				done <- true
			}()
		}

		for i := 0; i < 20; i++ {
			<-done
		}

		finalPrinter := Printer(tag)
		assert.NotNil(t, finalPrinter)
	})
}

func TestT(t *testing.T) {
	t.Run("Basic translation", func(t *testing.T) {
		tag := language.English
		key := "Hello, %s!"
		arg := "World"

		result := T(tag, key, arg)
		assert.Equal(t, "Hello, World!", result)
	})

	t.Run("Translation without arguments", func(t *testing.T) {
		tag := language.English
		key := "Simple message"

		result := T(tag, key)
		assert.Equal(t, "Simple message", result)
	})

	t.Run("Translation with multiple arguments", func(t *testing.T) {
		tag := language.English
		key := "%d %s cost $%.2f"
		args := []any{3, "apples", 5.99}

		result := T(tag, key, args...)
		assert.Equal(t, "3 apples cost $5.99", result)
	})

	t.Run("Different language tags", func(t *testing.T) {
		key := "Number: %d"
		arg := 42

		enResult := T(language.English, key, arg)
		frResult := T(language.French, key, arg)
		esResult := T(language.Spanish, key, arg)

		assert.Equal(t, "Number: 42", enResult)
		assert.Equal(t, "Number: 42", frResult)
		assert.Equal(t, "Number: 42", esResult)
	})
}

func TestTrans(t *testing.T) {
	t.Run("Language.Tag input", func(t *testing.T) {
		tag := language.German
		key := "Test message"

		result := trans(tag, key)
		assert.Equal(t, "Test message", result)
	})

	t.Run("*language.Tag input", func(t *testing.T) {
		tag := language.Russian
		key := "Test message"

		result := trans(&tag, key)
		assert.Equal(t, "Test message", result)
	})

	t.Run("String input", func(t *testing.T) {
		langStr := "zh-CN"
		key := "Test message"

		result := trans(langStr, key)
		assert.Equal(t, "Test message", result)
	})

	t.Run("*String input", func(t *testing.T) {
		langStr := "ja"
		key := "Test message"

		result := trans(&langStr, key)
		assert.Equal(t, "Test message", result)
	})

	t.Run("fmt.Stringer input", func(t *testing.T) {
		stringer := &testStringer{"ar"}
		key := "Test message"

		result := trans(stringer, key)
		assert.Equal(t, "Test message", result)
	})

	t.Run("With arguments", func(t *testing.T) {
		tag := language.Polish
		key := "Hello %s, you have %d messages"
		args := []any{"Alice", 5}

		result := trans(tag, key, args...)
		assert.Equal(t, "Hello Alice, you have 5 messages", result)
	})

	t.Run("Invalid language type should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			trans(123, "test message")
		}, "Should panic with invalid language tag")

		assert.Panics(t, func() {
			trans(nil, "test message")
		}, "Should panic with nil language tag")

		assert.Panics(t, func() {
			trans([]string{"en"}, "test message")
		}, "Should panic with slice language tag")
	})
}

func TestFuncMap(t *testing.T) {
	t.Run("FuncMap contains expected functions", func(t *testing.T) {
		assert.Contains(t, FuncMap, "L")
		assert.Contains(t, FuncMap, "T")
		assert.Contains(t, FuncMap, "t")
		assert.Contains(t, FuncMap, "i18n")
	})

	t.Run("L function parses language tags", func(t *testing.T) {
		lFunc, ok := FuncMap["L"].(func(string) language.Tag)
		require.True(t, ok)

		tag := lFunc("fr")
		assert.Equal(t, language.French, tag)
	})

	t.Run("T function is trans function", func(t *testing.T) {
		tFunc, ok := FuncMap["T"].(func(any, message.Reference, ...any) string)
		require.True(t, ok)

		result := tFunc("en", "Hello %s", "World")
		assert.Equal(t, "Hello World", result)
	})

	t.Run("t function is trans function", func(t *testing.T) {
		tFunc, ok := FuncMap["t"].(func(any, message.Reference, ...any) string)
		require.True(t, ok)

		result := tFunc("en", "Hello %s", "World")
		assert.Equal(t, "Hello World", result)
	})

	t.Run("i18n function is trans function", func(t *testing.T) {
		i18nFunc, ok := FuncMap["i18n"].(func(any, message.Reference, ...any) string)
		require.True(t, ok)

		result := i18nFunc("en", "Hello %s", "World")
		assert.Equal(t, "Hello World", result)
	})
}

func TestIntegration(t *testing.T) {
	t.Run("Full translation workflow", func(t *testing.T) {
		originalFallback := Fallback()
		defer SetFallback(originalFallback)

		SetFallback(language.French)

		tag := language.German
		printer := message.NewPrinter(tag)
		SetPrinter(tag, printer)

		key := "Welcome %s!"
		name := "User"

		result := T(tag, key, name)
		assert.Equal(t, "Welcome User!", result)

		templateResult := trans(tag, key, name)
		assert.Equal(t, result, templateResult)
	})

	t.Run("Template function integration", func(t *testing.T) {
		lFunc := FuncMap["L"]
		tFunc := FuncMap["T"]

		lang := lFunc.(func(string) language.Tag)("es")
		assert.Equal(t, language.Spanish, lang)

		result := tFunc.(func(any, message.Reference, ...any) string)(lang, "Count: %d", 42)
		assert.Equal(t, "Count: 42", result)
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Run("Concurrent fallback and printer operations", func(t *testing.T) {
		originalFallback := Fallback()
		defer SetFallback(originalFallback)

		done := make(chan bool, 30)
		tags := []language.Tag{
			language.English,
			language.French,
			language.German,
			language.Spanish,
			language.Italian,
		}

		for i := 0; i < 15; i++ {
			go func(idx int) {
				tag := tags[idx%len(tags)]
				printer := Printer(tag)
				assert.NotNil(t, printer)

				result := T(tag, "Test %d", idx)
				assert.NotEmpty(t, result)
				done <- true
			}(i)
		}

		for i := 0; i < 15; i++ {
			go func(idx int) {
				newFallback := tags[idx%len(tags)]
				SetFallback(newFallback)
				retrieved := Fallback()
				// Due to race conditions, we just verify that we get a valid fallback
				assert.Contains(t, tags, retrieved)
				done <- true
			}(i)
		}

		for i := 0; i < 30; i++ {
			<-done
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("Empty message key", func(t *testing.T) {
		result := T(language.English, "")
		assert.Equal(t, "", result)
	})

	t.Run("Special characters in message", func(t *testing.T) {
		key := "Special chars: !@#$%%^&*()_+-=[]{}|;':\",./<>?"
		result := T(language.English, key)
		assert.Equal(t, "Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?", result)
	})

	t.Run("Unicode characters", func(t *testing.T) {
		key := "Unicode: 你好 🌍 Café"
		result := T(language.English, key)
		assert.Equal(t, key, result)
	})

	t.Run("Complex format string", func(t *testing.T) {
		key := "Value: %08.2f, Count: %04d, String: %-10s"
		result := T(language.English, key, 3.14159, 42, "test")
		assert.Equal(t, "Value: 00,003.14, Count: 0,042, String: test      ", result)
	})
}

type testStringer struct {
	value string
}

func (t *testStringer) String() string {
	return t.value
}
