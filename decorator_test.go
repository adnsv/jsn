package jsn

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

type errorWriter struct {
	err error
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}

func TestDecoratorErrorHandling(t *testing.T) {
	testErr := fmt.Errorf("test error")
	tests := []struct {
		name      string
		writer    io.Writer
		input     any
		wantError error
	}{
		{
			name:      "writer error on string",
			writer:    &errorWriter{err: testErr},
			input:     "test",
			wantError: testErr,
		},
		{
			name:      "writer error on number",
			writer:    &errorWriter{err: testErr},
			input:     42,
			wantError: testErr,
		},
		{
			name:      "writer error on bool",
			writer:    &errorWriter{err: testErr},
			input:     true,
			wantError: testErr,
		},
		{
			name:      "writer error on null",
			writer:    &errorWriter{err: testErr},
			input:     nil,
			wantError: testErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := decorator{out: tt.writer}
			d.marshalValue(tt.input)
			if d.err != tt.wantError {
				t.Errorf("decorator error = %v, want %v", d.err, tt.wantError)
			}
		})
	}
}

func TestMarshalerErrors(t *testing.T) {
	testErr := fmt.Errorf("test marshaler error")
	tests := []struct {
		name      string
		input     any
		wantError error
	}{
		{
			name:      "StrMarshaler error",
			input:     errorStrMarshaler{err: testErr},
			wantError: testErr,
		},
		{
			name:      "ArrMarshaler error",
			input:     errorArrMarshaler{err: testErr},
			wantError: testErr,
		},
		{
			name:      "ObjMarshaler error",
			input:     errorObjMarshaler{err: testErr},
			wantError: testErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			d := decorator{out: &sb}
			d.marshalValue(tt.input)
			if d.err != tt.wantError {
				t.Errorf("decorator error = %v, want %v", d.err, tt.wantError)
			}
		})
	}
}

func Test_decorator_scrambleStr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"simple string", "abc", "abc"},
		{"special characters", "\b\f\n\r\t\\\"", "\\b\\f\\n\\r\\t\\\\\\\""},
		{"unicode control characters", "\u0000\u001f", "\\u0000\\u001f"},
		{"normal string", "normal string", "normal string"},
		{"escaped backslash", "\\", "\\\\"},
		{"escaped quote", "\"", "\\\""},
		{"mixed special and normal", "abc\nxyz", "abc\\nxyz"},
		{"unicode characters", "こんにちは", "こんにちは"},
		{"emoji", "😀", "😀"},
		{"long string", strings.Repeat("a", 1000), strings.Repeat("a", 1000)},
		{"all special characters", "\b\f\n\r\t\\\"/", "\\b\\f\\n\\r\\t\\\\\\\"/"},
		{"mixed unicode and special", "こんにちは\n世界", "こんにちは\\n世界"},
		{"control characters", "\x01\x02\x03", "\\u0001\\u0002\\u0003"},
		{"string with null character", "null\x00char", "null\\u0000char"},
		{"string with backspace", "backspace\btest", "backspace\\btest"},
		{"string with form feed", "formfeed\ftest", "formfeed\\ftest"},
		{"string with carriage return", "carriage\rreturn", "carriage\\rreturn"},
		{"string with tab", "tab\ttest", "tab\\ttest"},
		{"string with newline", "newline\ntest", "newline\\ntest"},
		{"string with slash", "slash/test", "slash/test"},
		{"very long unicode", strings.Repeat("🚀", 100), strings.Repeat("🚀", 100)},
		{"mixed control chars", "\x00\n\t\r\x1F", "\\u0000\\n\\t\\r\\u001f"},
		{"all ascii control chars", "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x0C\x0D\x0E\x0F\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1A\x1B\x1C\x1D\x1E\x1F",
			"\\u0000\\u0001\\u0002\\u0003\\u0004\\u0005\\u0006\\u0007\\b\\t\\n\\u000b\\f\\r\\u000e\\u000f\\u0010\\u0011\\u0012\\u0013\\u0014\\u0015\\u0016\\u0017\\u0018\\u0019\\u001a\\u001b\\u001c\\u001d\\u001e\\u001f"},
		{"surrogate pairs", "𝄞", "𝄞"}, // musical G clef
		{"zero-width chars", "\u200B\u200C\u200D", "\u200B\u200C\u200D"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			d := decorator{out: &sb}
			d.scrambleStr(tt.input)
			got := sb.String()
			if got != tt.want {
				t.Errorf("scrambleStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
