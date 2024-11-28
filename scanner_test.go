package jsn

import (
	"testing"
)

func TestScanner_Basic(t *testing.T) {
	s := NewScanner([]byte("hello"))

	if s.peek() != 'h' {
		t.Errorf("Expected 'h', got %c", s.peek())
	}

	if s.next() != 'h' {
		t.Errorf("Expected 'h', got %c", s.next())
	}

	if s.cur != 1 {
		t.Errorf("Expected position 1, got %d", s.cur)
	}
}

func TestScanner_SkipWhitespace(t *testing.T) {
	s := NewScanner([]byte("  \t\n\r  hello"))
	s.skipWhitespace()

	if s.peek() != 'h' {
		t.Errorf("Expected 'h', got %c", s.peek())
	}
}

func TestScanner_ParseString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		// Basic cases
		{name: "simple string", input: `"hello"`, want: "hello"},
		{name: "empty string", input: `""`, want: ""},
		{name: "single space", input: `" "`, want: " "},

		// Escape sequences
		{name: "escaped backspace", input: `"\b"`, want: "\b"},
		{name: "escaped tab", input: `"\t"`, want: "\t"},
		{name: "escaped newline", input: `"\n"`, want: "\n"},
		{name: "escaped form feed", input: `"\f"`, want: "\f"},
		{name: "escaped carriage return", input: `"\r"`, want: "\r"},
		{name: "escaped quote", input: `"\""`, want: "\""},
		{name: "escaped backslash", input: `"\\"`, want: "\\"},
		{name: "escaped forward slash", input: `"\/"`, want: "/"},
		{name: "multiple escapes", input: `"\t\n\r\b\f"`, want: "\t\n\r\b\f"},
		{name: "mixed escapes", input: `"hello\tworld\n"`, want: "hello\tworld\n"},

		// Unicode escapes
		{name: "unicode space", input: `"\u0020"`, want: " "},
		{name: "unicode null", input: `"\u0000"`, want: "\u0000"},
		{name: "unicode max", input: `"\uFFFF"`, want: "\uFFFF"},
		{name: "multiple unicode", input: `"\u0020\u0020"`, want: "  "},
		{name: "text with escaped null", input: `"hello\u0000world"`, want: "hello\u0000world"},

		// Error cases
		{name: "unterminated string", input: `"hello`, wantErr: ErrInvalidString},
		{name: "invalid escape", input: `"\k"`, wantErr: ErrInvalidString},
		{name: "incomplete unicode", input: `"\u123"`, wantErr: ErrInvalidUnicodeEscape},
		{name: "invalid unicode", input: `"\uGGGG"`, wantErr: ErrInvalidUnicodeEscape},
		{name: "bare backslash", input: `"\"`, wantErr: ErrInvalidString},
		{name: "null after escape", input: "\"\\x00\"", wantErr: ErrInvalidString},

		// Control characters (should be invalid unless escaped)
		{name: "raw null", input: "\"\x00\"", wantErr: ErrInvalidString},
		{name: "raw control", input: "\"\x1F\"", wantErr: ErrInvalidString},
		{name: "raw bell", input: "\"\x07\"", wantErr: ErrInvalidString},
		{name: "raw tab", input: "\"\x09\"", wantErr: ErrInvalidString},
		{name: "raw newline", input: "\"\x0A\"", wantErr: ErrInvalidString},
		{name: "raw return", input: "\"\x0D\"", wantErr: ErrInvalidString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner([]byte(tt.input))
			got, err := s.parseString()

			if err != tt.wantErr {
				t.Errorf("parseString() error = %v, want %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("parseString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestScanner_ParseNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr error
	}{
		// Valid numbers
		{name: "integer", input: "123", want: 123},
		{name: "negative", input: "-123", want: -123},
		{name: "decimal", input: "123.456", want: 123.456},
		{name: "exponent", input: "1.23e2", want: 123},
		{name: "negative exponent", input: "1.23e-2", want: 0.0123},
		{name: "zero", input: "0", want: 0},
		{name: "negative zero", input: "-0", want: 0},

		// Invalid numbers
		{name: "leading zero", input: "01", wantErr: ErrInvalidNumber},
		{name: "multiple dots", input: "12.34.56", wantErr: ErrInvalidNumber},
		{name: "trailing dot", input: "123.", wantErr: ErrInvalidNumber},
		{name: "missing exponent", input: "1e", wantErr: ErrInvalidNumber},
		{name: "invalid exponent", input: "1e-", wantErr: ErrInvalidNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner([]byte(tt.input))
			got, err := s.parseNumber()
			if err == nil {
				err = s.Finalize()
			}

			if err != tt.wantErr {
				t.Errorf("parseNumber() error = %v, want %v", err, tt.wantErr)
				return
			}

			if err == nil && got != tt.want {
				t.Errorf("parseNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanner_Value(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		// Valid values
		{name: "empty object", input: "{}", wantErr: nil},
		{name: "simple object", input: `{"key": "value"}`, wantErr: nil},
		{name: "empty array", input: "[]", wantErr: nil},
		{name: "number array", input: "[1, 2, 3]", wantErr: nil},
		{name: "nested structure", input: `{"array": [1, 2, {"key": "value"}]}`, wantErr: nil},
		{name: "null", input: "null", wantErr: nil},
		{name: "true", input: "true", wantErr: nil},
		{name: "false", input: "false", wantErr: nil},

		// Invalid values
		{name: "empty input", input: "", wantErr: ErrUnexpectedEOF},
		{name: "only whitespace", input: "   ", wantErr: ErrUnexpectedEOF},
		{name: "incomplete true", input: "tr", wantErr: ErrUnexpectedToken},
		{name: "incomplete false", input: "fals", wantErr: ErrUnexpectedToken},
		{name: "incomplete null", input: "nul", wantErr: ErrUnexpectedToken},
		{name: "invalid number", input: "01", wantErr: ErrInvalidNumber},
		{name: "unterminated string", input: `"hello`, wantErr: ErrInvalidString},
		{name: "unterminated array", input: "[1, 2", wantErr: ErrUnexpectedEOF},
		{name: "unterminated object", input: `{"key": "value"`, wantErr: ErrUnexpectedEOF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner([]byte(tt.input))
			_, err := ReadValue(s)
			if err == nil {
				err = s.Finalize()
			}

			if err != tt.wantErr {
				t.Errorf("ReadValue() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanner_SkipBOM(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantSkip bool
	}{
		{
			name:     "with BOM",
			input:    []byte{0xEF, 0xBB, 0xBF, 'h', 'e', 'l', 'l', 'o'},
			wantSkip: true,
		},
		{
			name:     "without BOM",
			input:    []byte("hello"),
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(tt.input)
			if got := s.SkipBOM(); got != tt.wantSkip {
				t.Errorf("SkipBOM() = %v, want %v", got, tt.wantSkip)
			}
		})
	}
}

func TestScanner_Flags(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		flags    []any
		wantNext byte
	}{
		{
			name:     "default behavior skips BOM",
			input:    []byte{0xEF, 0xBB, 0xBF, 'h', 'e', 'l', 'l', 'o'},
			flags:    nil,
			wantNext: 'h',
		},
		{
			name:     "keep BOM",
			input:    []byte{0xEF, 0xBB, 0xBF, 'h', 'e', 'l', 'l', 'o'},
			flags:    []any{ScannerFlagDoNotSkipBOM},
			wantNext: 0xEF,
		},
		{
			name:     "keep whitespace",
			input:    []byte("  hello"),
			flags:    []any{ScannerFlagDoNotSkipInitialWhitespace},
			wantNext: ' ',
		},
		{
			name:     "multiple flags",
			input:    []byte{0xEF, 0xBB, 0xBF, ' ', 'h'},
			flags:    []any{ScannerFlagDoNotSkipBOM, ScannerFlagDoNotSkipInitialWhitespace},
			wantNext: 0xEF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(tt.input, tt.flags...)
			if got := s.peek(); got != tt.wantNext {
				t.Errorf("Scanner.peek() = %v, want %v", got, tt.wantNext)
			}
		})
	}
}

func TestScanner_InvalidOption(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid option type")
		}
	}()

	NewScanner([]byte("test"), "invalid option")
}

func TestScannerEdgeCases(t *testing.T) {
	stringTests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: ErrUnexpectedEOF,
		},
		{
			name:    "invalid unicode escape",
			input:   `"\u123"`, // incomplete unicode escape
			wantErr: ErrInvalidUnicodeEscape,
		},
		{
			name:    "invalid unicode escape sequence",
			input:   `"\uXYZW"`,
			wantErr: ErrInvalidUnicodeEscape,
		},
		{
			name:    "invalid unterminated string",
			input:   `"hello`,
			wantErr: ErrInvalidString,
		},
		{
			name:    "invalid escape sequence",
			input:   `"\x"`,
			wantErr: ErrInvalidString,
		},
	}

	numberTests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "invalid number format",
			input:   "123.456.789",
			wantErr: ErrInvalidNumber,
		},
		{
			name:    "number with multiple exponents",
			input:   "1e2e3",
			wantErr: ErrInvalidNumber,
		},
		{
			name:    "number with invalid exponent",
			input:   "1e",
			wantErr: ErrInvalidNumber,
		},
	}

	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner([]byte(tt.input))
			var err error
			if s.IsEOF() {
				err = ErrUnexpectedEOF
			} else {
				_, err = s.parseString()
				if err == nil {
					err = s.Finalize()
				}
			}
			if err != tt.wantErr {
				t.Errorf("Scanner string error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	for _, tt := range numberTests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner([]byte(tt.input))
			var err error
			if s.IsEOF() {
				err = ErrUnexpectedEOF
			} else {
				_, err = s.parseNumber()
				if err == nil {
					err = s.Finalize()
				}
			}
			if err != tt.wantErr {
				t.Errorf("Scanner number error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestScannerFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		flags    ScannerFlag
		wantByte byte
	}{
		{
			name:     "skip BOM by default",
			input:    []byte{0xEF, 0xBB, 0xBF, 'a'},
			flags:    0,
			wantByte: 'a',
		},
		{
			name:     "do not skip BOM",
			input:    []byte{0xEF, 0xBB, 0xBF, 'a'},
			flags:    ScannerFlagDoNotSkipBOM,
			wantByte: 0xEF,
		},
		{
			name:     "skip whitespace by default",
			input:    []byte(" \t\n\ra"),
			flags:    0,
			wantByte: 'a',
		},
		{
			name:     "do not skip whitespace",
			input:    []byte(" a"),
			flags:    ScannerFlagDoNotSkipInitialWhitespace,
			wantByte: ' ',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(tt.input, tt.flags)
			if got := s.peek(); got != tt.wantByte {
				t.Errorf("Scanner.peek() = %v, want %v", got, tt.wantByte)
			}
		})
	}
}

func TestScannerNumberParsing(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr error
	}{
		{
			name:  "integer zero",
			input: "0",
			want:  0,
		},
		{
			name:  "negative zero",
			input: "-0",
			want:  0,
		},
		{
			name:  "decimal zero",
			input: "0.0",
			want:  0,
		},
		{
			name:  "exponential zero",
			input: "0e0",
			want:  0,
		},
		{
			name:    "leading zeros",
			input:   "00123",
			wantErr: ErrInvalidNumber,
		},
		{
			name:    "negative leading zeros",
			input:   "-00123",
			wantErr: ErrInvalidNumber,
		},
		{
			name:    "multiple decimal points",
			input:   "123.456.789",
			wantErr: ErrInvalidNumber,
		},
		{
			name:    "multiple exponents",
			input:   "1e2e3",
			wantErr: ErrInvalidNumber,
		},
		{
			name:    "invalid exponent",
			input:   "1e",
			wantErr: ErrInvalidNumber,
		},
		{
			name:    "missing exponent value",
			input:   "1e+",
			wantErr: ErrInvalidNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner([]byte(tt.input))
			got, err := s.parseNumber()
			if err != tt.wantErr {
				t.Errorf("Scanner.parseNumber() error = %v, want %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.want {
				t.Errorf("Scanner.parseNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}
