package jsn

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestReadValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    any
		wantErr error
	}{
		// String tests
		{name: "simple string", input: `"hello"`, want: "hello"},
		{name: "string with escapes", input: `"hello\nworld"`, want: "hello\nworld"},

		// Number tests
		{name: "integer", input: "42", want: float64(42)},
		{name: "negative", input: "-42", want: float64(-42)},
		{name: "float", input: "42.5", want: float64(42.5)},
		{name: "scientific", input: "1.2e3", want: float64(1200)},

		// Boolean tests
		{name: "true", input: "true", want: true},
		{name: "false", input: "false", want: false},

		// Null test
		{name: "null", input: "null", want: nil},

		// Object tests
		{name: "empty object", input: "{}", want: map[string]any{}},
		{name: "simple object", input: `{"key": "value"}`, want: map[string]any{"key": "value"}},

		// Array tests - being explicit about types
		{name: "empty array", input: "[]", want: []any{}},
		{name: "simple array", input: `[1,2,3]`, want: []any{float64(1), float64(2), float64(3)}},
		{name: "mixed array", input: `[1,"two",true]`, want: []any{float64(1), "two", true}},

		// Error cases
		{name: "invalid token", input: "invalid", wantErr: ErrUnexpectedToken},
		{name: "incomplete true", input: "tru", wantErr: ErrUnexpectedToken},
		{name: "incomplete false", input: "fals", wantErr: ErrUnexpectedToken},
		{name: "incomplete null", input: "nul", wantErr: ErrUnexpectedToken},

		// Additional number tests
		{name: "huge exponent", input: "1e999", wantErr: ErrNumericValueOutOfRange},
		{name: "tiny exponent", input: "1e-999", want: float64(0)},
		{name: "max float64", input: "1.8e308", wantErr: ErrNumericValueOutOfRange},
		{name: "min float64", input: "-1.8e308", wantErr: ErrNumericValueOutOfRange},

		// Invalid array syntax
		{name: "array comma after close", input: "[1],", wantErr: ErrUnexpectedToken},
		{name: "array trailing comma", input: "[1,]", wantErr: ErrUnexpectedToken},
		{name: "array missing comma", input: "[1 2]", wantErr: ErrUnexpectedToken},

		// Invalid object syntax
		{name: "object comma after close", input: `{"a":1},`, wantErr: ErrUnexpectedToken},
		{name: "object trailing comma", input: `{"a":1,}`, wantErr: ErrUnexpectedToken},
		{name: "object missing comma", input: `{"a":1 "b":2}`, wantErr: ErrUnexpectedToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner([]byte(tt.input))
			got, err := ReadValue(s)
			if err == nil {
				err = s.Finalize()
			}

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ReadValue() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ReadValue() unexpected error = %v", err)
				return
			}

			// Special handling for slices
			gotSlice, gotIsSlice := got.([]any)
			wantSlice, wantIsSlice := tt.want.([]any)
			if gotIsSlice && wantIsSlice {
				if len(gotSlice) == 0 && len(gotSlice) != len(wantSlice) {
					t.Errorf("ReadValue() = %v, want %v", got, tt.want)
				}
				return
			}

			// For non-slice types, use regular comparison
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadValueNested(t *testing.T) {
	input := `{
		"string": "hello",
		"number": 42,
		"array": [1, "two", true],
		"object": {"nested": "value"}
	}`

	want := map[string]any{
		"string": "hello",
		"number": float64(42),
		"array":  []any{float64(1), "two", true},
		"object": map[string]any{"nested": "value"},
	}

	s := NewScanner([]byte(input))
	got, err := ReadValue(s)
	if err == nil {
		err = s.Finalize()
	}

	if err != nil {
		t.Errorf("ReadValue() unexpected error = %v", err)
		return
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadValue() = %v, want %v", got, want)
	}
}

func TestReadObject(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]any
		wantErr error
	}{
		{
			name:  "empty object",
			input: "{}",
			want:  map[string]any{},
		},
		{
			name:  "simple object",
			input: `{"key": "value"}`,
			want:  map[string]any{"key": "value"},
		},
		{
			name:  "nested object",
			input: `{"outer": {"inner": "value"}}`,
			want:  map[string]any{"outer": map[string]any{"inner": "value"}},
		},
		{
			name:    "invalid object",
			input:   `{"key": }`,
			wantErr: ErrUnexpectedToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner([]byte(tt.input))
			got, err := ReadObject(s)
			if err == nil {
				err = s.Finalize()
			}

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ReadObject() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ReadObject() unexpected error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadArray(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []any
		wantErr error
	}{
		{
			name:  "empty array",
			input: "[]",
			want:  []any{},
		},
		{
			name:  "simple array",
			input: "[1,2,3]",
			want:  []any{float64(1), float64(2), float64(3)},
		},
		{
			name:  "mixed array",
			input: `[1,"two",true]`,
			want:  []any{float64(1), "two", true},
		},
		{
			name:    "invalid array",
			input:   "[1,]",
			wantErr: ErrUnexpectedToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner([]byte(tt.input))
			got, err := ReadArray(s)
			if err == nil {
				err = s.Finalize()
			}

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ReadArray() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ReadArray() unexpected error = %v", err)
				return
			}

			// If both slices are empty, consider them equal regardless of nil vs empty slice
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSuite(t *testing.T) {
	for _, tt := range NSTTestSuiteData {
		t.Run(tt.Name, func(t *testing.T) {
			s := NewScanner([]byte(tt.Content))
			_, err := ReadValue(s)
			if err == nil {
				err = s.Finalize()
			}

			switch tt.Kind {
			case "y":
				if err != nil {
					t.Errorf("ReadValue() unexpected error = %v", err)
					return
				}

			case "n":
				if err == nil {
					t.Errorf("ReadValue() expected error, got nil")
					return
				}

			case "i":
				if err == nil {
					t.Log("Accept")
				} else {
					t.Logf("Reject: %v", err)
				}
			}
		})
	}
}

func TestScanner_LargeArrayNesting(t *testing.T) {
	// Generate a string with 100,000 opening brackets and matching closing brackets
	const numBrackets = 1000

	t.Run("deeply nested array openers", func(t *testing.T) {
		input := strings.Repeat("[", numBrackets)

		s := NewScanner([]byte(input))
		_, err := ReadValue(s)
		if err == nil {
			err = s.Finalize()
		}

		if err == nil {
			t.Error("Expected error for deeply nested array openers")
		}
	})

	t.Run("deeply nested array", func(t *testing.T) {
		input := strings.Repeat("[", numBrackets) + strings.Repeat("]", numBrackets)

		s := NewScanner([]byte(input))
		_, err := ReadValue(s)
		if err == nil {
			err = s.Finalize()
		}

		if err != nil {
			t.Errorf("Unexpected error for deeply nested array closers: %v", err)
		}
	})
}

func TestReader_DeepMixedNesting(t *testing.T) {
	t.Run("deeply nested array-object structure", func(t *testing.T) {
		// Generate a pattern of [{"": repeated many times
		const numRepetitions = 1000
		pattern := `[{"":` // Base pattern to repeat

		// Build input string
		var builder strings.Builder
		builder.Grow(numRepetitions * len(pattern))
		for i := 0; i < numRepetitions; i++ {
			builder.WriteString(pattern)
		}
		input := builder.String()

		s := NewScanner([]byte(input))
		_, err := ReadValue(s)
		if err == nil {
			err = s.Finalize()
		}

		if err == nil {
			t.Error("Expected error for deeply nested array-object structure, got nil")
		}
	})
}

// Data extractedfrom Nicolas Seriot's JSONTestSuite
// https://github.com/nst/JSONTestSuite
// Copyright (c) 2016 Nicolas Seriot
var NSTTestSuiteData = []struct {
	Kind    string
	Name    string
	Content string
}{
	{Kind: "i", Name: "number_double_huge_neg_exp", Content: "[123.456e-789]"},
	{Kind: "i", Name: "number_huge_exp", Content: "[0.4e00669999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999969999999006]"},
	{Kind: "i", Name: "number_neg_int_huge_exp", Content: "[-1e+9999]"},
	{Kind: "i", Name: "number_pos_double_huge_exp", Content: "[1.5e+9999]"},
	{Kind: "i", Name: "number_real_neg_overflow", Content: "[-123123e100000]"},
	{Kind: "i", Name: "number_real_pos_overflow", Content: "[123123e100000]"},
	{Kind: "i", Name: "number_real_underflow", Content: "[123e-10000000]"},
	{Kind: "i", Name: "number_too_big_neg_int", Content: "[-123123123123123123123123123123]"},
	{Kind: "i", Name: "number_too_big_pos_int", Content: "[100000000000000000000]"},
	{Kind: "i", Name: "number_very_big_negative_int", Content: "[-237462374673276894279832749832423479823246327846]"},
	{Kind: "i", Name: "object_key_lone_2nd_surrogate", Content: "{\"\\uDFAA\":0}"},
	{Kind: "i", Name: "string_1st_surrogate_but_2nd_missing", Content: "[\"\\uDADA\"]"},
	{Kind: "i", Name: "string_1st_valid_surrogate_2nd_invalid", Content: "[\"\\uD888\\u1234\"]"},
	{Kind: "i", Name: "string_UTF-16LE_with_BOM", Content: "\xff\xfe[\x00\"\x00\xe9\x00\"\x00]\x00"},
	{Kind: "i", Name: "string_UTF-8_invalid_sequence", Content: "[\"日ш\xfa\"]"},
	{Kind: "i", Name: "string_UTF8_surrogate_U+D800", Content: "[\"\xed\xa0\x80\"]"},
	{Kind: "i", Name: "string_incomplete_surrogate_and_escape_valid", Content: "[\"\\uD800\\n\"]"},
	{Kind: "i", Name: "string_incomplete_surrogate_pair", Content: "[\"\\uDd1ea\"]"},
	{Kind: "i", Name: "string_incomplete_surrogates_escape_valid", Content: "[\"\\uD800\\uD800\\n\"]"},
	{Kind: "i", Name: "string_invalid_lonely_surrogate", Content: "[\"\\ud800\"]"},
	{Kind: "i", Name: "string_invalid_surrogate", Content: "[\"\\ud800abc\"]"},
	{Kind: "i", Name: "string_invalid_utf-8", Content: "[\"\xff\"]"},
	{Kind: "i", Name: "string_inverted_surrogates_U+1D11E", Content: "[\"\\uDd1e\\uD834\"]"},
	{Kind: "i", Name: "string_iso_latin_1", Content: "[\"\xe9\"]"},
	{Kind: "i", Name: "string_lone_second_surrogate", Content: "[\"\\uDFAA\"]"},
	{Kind: "i", Name: "string_lone_utf8_continuation_byte", Content: "[\"\x81\"]"},
	{Kind: "i", Name: "string_not_in_unicode_range", Content: "[\"\xf4\xbf\xbf\xbf\"]"},
	{Kind: "i", Name: "string_overlong_sequence_2_bytes", Content: "[\"\xc0\xaf\"]"},
	{Kind: "i", Name: "string_overlong_sequence_6_bytes", Content: "[\"\xfc\x83\xbf\xbf\xbf\xbf\"]"},
	{Kind: "i", Name: "string_overlong_sequence_6_bytes_null", Content: "[\"\xfc\x80\x80\x80\x80\x80\"]"},
	{Kind: "i", Name: "string_truncated-utf-8", Content: "[\"\xe0\xff\"]"},
	{Kind: "i", Name: "string_utf16BE_no_BOM", Content: "\x00[\x00\"\x00\xe9\x00\"\x00]"},
	{Kind: "i", Name: "string_utf16LE_no_BOM", Content: "[\x00\"\x00\xe9\x00\"\x00]\x00"},
	{Kind: "i", Name: "structure_UTF-8_BOM_empty_object", Content: "\ufeff{}"},
	{Kind: "n", Name: "array_1_true_without_comma", Content: "[1 true]"},
	{Kind: "n", Name: "array_a_invalid_utf8", Content: "[a\xe5]"},
	{Kind: "n", Name: "array_colon_instead_of_comma", Content: "[\"\": 1]"},
	{Kind: "n", Name: "array_comma_after_close", Content: "[\"\"],"},
	{Kind: "n", Name: "array_comma_and_number", Content: "[,1]"},
	{Kind: "n", Name: "array_double_comma", Content: "[1,,2]"},
	{Kind: "n", Name: "array_double_extra_comma", Content: "[\"x\",,]"},
	{Kind: "n", Name: "array_extra_close", Content: "[\"x\"]]"},
	{Kind: "n", Name: "array_extra_comma", Content: "[\"\",]"},
	{Kind: "n", Name: "array_incomplete", Content: "[\"x\""},
	{Kind: "n", Name: "array_incomplete_invalid_value", Content: "[x"},
	{Kind: "n", Name: "array_inner_array_no_comma", Content: "[3[4]]"},
	{Kind: "n", Name: "array_invalid_utf8", Content: "[\xff]"},
	{Kind: "n", Name: "array_items_separated_by_semicolon", Content: "[1:2]"},
	{Kind: "n", Name: "array_just_comma", Content: "[,]"},
	{Kind: "n", Name: "array_just_minus", Content: "[-]"},
	{Kind: "n", Name: "array_missing_value", Content: "[   , \"\"]"},
	{Kind: "n", Name: "array_newlines_unclosed", Content: "[\"a\",\r\n4\r\n,1,"},
	{Kind: "n", Name: "array_number_and_comma", Content: "[1,]"},
	{Kind: "n", Name: "array_number_and_several_commas", Content: "[1,,]"},
	{Kind: "n", Name: "array_spaces_vertical_tab_formfeed", Content: "[\"\va\"\\f]"},
	{Kind: "n", Name: "array_star_inside", Content: "[*]"},
	{Kind: "n", Name: "array_unclosed", Content: "[\"\""},
	{Kind: "n", Name: "array_unclosed_trailing_comma", Content: "[1,"},
	{Kind: "n", Name: "array_unclosed_with_new_lines", Content: "[1,\r\n1\r\n,1"},
	{Kind: "n", Name: "array_unclosed_with_object_inside", Content: "[{}"},
	{Kind: "n", Name: "incomplete_false", Content: "[fals]"},
	{Kind: "n", Name: "incomplete_null", Content: "[nul]"},
	{Kind: "n", Name: "incomplete_true", Content: "[tru]"},
	{Kind: "n", Name: "multidigit_number_then_00", Content: "123\x00"},
	{Kind: "n", Name: "number_++", Content: "[++1234]"},
	{Kind: "n", Name: "number_+1", Content: "[+1]"},
	{Kind: "n", Name: "number_+Inf", Content: "[+Inf]"},
	{Kind: "n", Name: "number_-01", Content: "[-01]"},
	{Kind: "n", Name: "number_-1.0.", Content: "[-1.0.]"},
	{Kind: "n", Name: "number_-2.", Content: "[-2.]"},
	{Kind: "n", Name: "number_-NaN", Content: "[-NaN]"},
	{Kind: "n", Name: "number_.-1", Content: "[.-1]"},
	{Kind: "n", Name: "number_.2e-3", Content: "[.2e-3]"},
	{Kind: "n", Name: "number_0.1.2", Content: "[0.1.2]"},
	{Kind: "n", Name: "number_0.3e+", Content: "[0.3e+]"},
	{Kind: "n", Name: "number_0.3e", Content: "[0.3e]"},
	{Kind: "n", Name: "number_0.e1", Content: "[0.e1]"},
	{Kind: "n", Name: "number_0_capital_E+", Content: "[0E+]"},
	{Kind: "n", Name: "number_0_capital_E", Content: "[0E]"},
	{Kind: "n", Name: "number_0e+", Content: "[0e+]"},
	{Kind: "n", Name: "number_0e", Content: "[0e]"},
	{Kind: "n", Name: "number_1.0e+", Content: "[1.0e+]"},
	{Kind: "n", Name: "number_1.0e-", Content: "[1.0e-]"},
	{Kind: "n", Name: "number_1.0e", Content: "[1.0e]"},
	{Kind: "n", Name: "number_1_000", Content: "[1 000.0]"},
	{Kind: "n", Name: "number_1eE2", Content: "[1eE2]"},
	{Kind: "n", Name: "number_2.e+3", Content: "[2.e+3]"},
	{Kind: "n", Name: "number_2.e-3", Content: "[2.e-3]"},
	{Kind: "n", Name: "number_2.e3", Content: "[2.e3]"},
	{Kind: "n", Name: "number_9.e+", Content: "[9.e+]"},
	{Kind: "n", Name: "number_Inf", Content: "[Inf]"},
	{Kind: "n", Name: "number_NaN", Content: "[NaN]"},
	{Kind: "n", Name: "number_U+FF11_fullwidth_digit_one", Content: "[１]"},
	{Kind: "n", Name: "number_expression", Content: "[1+2]"},
	{Kind: "n", Name: "number_hex_1_digit", Content: "[0x1]"},
	{Kind: "n", Name: "number_hex_2_digits", Content: "[0x42]"},
	{Kind: "n", Name: "number_infinity", Content: "[Infinity]"},
	{Kind: "n", Name: "number_invalid+-", Content: "[0e+-1]"},
	{Kind: "n", Name: "number_invalid-negative-real", Content: "[-123.123foo]"},
	{Kind: "n", Name: "number_invalid-utf-8-in-bigger-int", Content: "[123\xe5]"},
	{Kind: "n", Name: "number_invalid-utf-8-in-exponent", Content: "[1e1\xe5]"},
	{Kind: "n", Name: "number_invalid-utf-8-in-int", Content: "[0\xe5]\r\n"},
	{Kind: "n", Name: "number_minus_infinity", Content: "[-Infinity]"},
	{Kind: "n", Name: "number_minus_sign_with_trailing_garbage", Content: "[-foo]"},
	{Kind: "n", Name: "number_minus_space_1", Content: "[- 1]"},
	{Kind: "n", Name: "number_neg_int_starting_with_zero", Content: "[-012]"},
	{Kind: "n", Name: "number_neg_real_without_int_part", Content: "[-.123]"},
	{Kind: "n", Name: "number_neg_with_garbage_at_end", Content: "[-1x]"},
	{Kind: "n", Name: "number_real_garbage_after_e", Content: "[1ea]"},
	{Kind: "n", Name: "number_real_with_invalid_utf8_after_e", Content: "[1e\xe5]"},
	{Kind: "n", Name: "number_real_without_fractional_part", Content: "[1.]"},
	{Kind: "n", Name: "number_starting_with_dot", Content: "[.123]"},
	{Kind: "n", Name: "number_with_alpha", Content: "[1.2a-3]"},
	{Kind: "n", Name: "number_with_alpha_char", Content: "[1.8011670033376514H-308]"},
	{Kind: "n", Name: "number_with_leading_zero", Content: "[012]"},
	{Kind: "n", Name: "object_bad_value", Content: "[\"x\", truth]"},
	{Kind: "n", Name: "object_bracket_key", Content: "{[: \"x\"}\r\n"},
	{Kind: "n", Name: "object_comma_instead_of_colon", Content: "{\"x\", null}"},
	{Kind: "n", Name: "object_double_colon", Content: "{\"x\"::\"b\"}"},
	{Kind: "n", Name: "object_emoji", Content: "{🇨🇭}"},
	{Kind: "n", Name: "object_garbage_at_end", Content: "{\"a\":\"a\" 123}"},
	{Kind: "n", Name: "object_key_with_single_quotes", Content: "{key: 'value'}"},
	{Kind: "n", Name: "object_lone_continuation_byte_in_key_and_trailing_comma", Content: "{\"\xb9\":\"0\",}"},
	{Kind: "n", Name: "object_missing_colon", Content: "{\"a\" b}"},
	{Kind: "n", Name: "object_missing_key", Content: "{:\"b\"}"},
	{Kind: "n", Name: "object_missing_semicolon", Content: "{\"a\" \"b\"}"},
	{Kind: "n", Name: "object_missing_value", Content: "{\"a\":"},
	{Kind: "n", Name: "object_no-colon", Content: "{\"a\""},
	{Kind: "n", Name: "object_non_string_key", Content: "{1:1}"},
	{Kind: "n", Name: "object_non_string_key_but_huge_number_instead", Content: "{9999E9999:1}"},
	{Kind: "n", Name: "object_repeated_null_null", Content: "{null:null,null:null}"},
	{Kind: "n", Name: "object_several_trailing_commas", Content: "{\"id\":0,,,,,}"},
	{Kind: "n", Name: "object_single_quote", Content: "{'a':0}"},
	{Kind: "n", Name: "object_trailing_comma", Content: "{\"id\":0,}"},
	{Kind: "n", Name: "object_trailing_comment", Content: "{\"a\":\"b\"}/**/"},
	{Kind: "n", Name: "object_trailing_comment_open", Content: "{\"a\":\"b\"}/**//"},
	{Kind: "n", Name: "object_trailing_comment_slash_open", Content: "{\"a\":\"b\"}//"},
	{Kind: "n", Name: "object_trailing_comment_slash_open_incomplete", Content: "{\"a\":\"b\"}/"},
	{Kind: "n", Name: "object_two_commas_in_a_row", Content: "{\"a\":\"b\",,\"c\":\"d\"}"},
	{Kind: "n", Name: "object_unquoted_key", Content: "{a: \"b\"}"},
	{Kind: "n", Name: "object_unterminated-value", Content: "{\"a\":\"a"},
	{Kind: "n", Name: "object_with_single_string", Content: "{ \"foo\" : \"bar\", \"a\" }"},
	{Kind: "n", Name: "object_with_trailing_garbage", Content: "{\"a\":\"b\"}#"},
	{Kind: "n", Name: "single_space", Content: " "},
	{Kind: "n", Name: "string_1_surrogate_then_escape", Content: "[\"\\uD800\\\"]"},
	{Kind: "n", Name: "string_1_surrogate_then_escape_u", Content: "[\"\\uD800\\u\"]"},
	{Kind: "n", Name: "string_1_surrogate_then_escape_u1", Content: "[\"\\uD800\\u1\"]"},
	{Kind: "n", Name: "string_1_surrogate_then_escape_u1x", Content: "[\"\\uD800\\u1x\"]"},
	{Kind: "n", Name: "string_accentuated_char_no_quotes", Content: "[é]"},
	{Kind: "n", Name: "string_backslash_00", Content: "[\"\\\x00\"]"},
	{Kind: "n", Name: "string_escape_x", Content: "[\"\\x00\"]"},
	{Kind: "n", Name: "string_escaped_backslash_bad", Content: "[\"\\\\\\\"]"},
	{Kind: "n", Name: "string_escaped_ctrl_char_tab", Content: "[\"\\\t\"]"},
	{Kind: "n", Name: "string_escaped_emoji", Content: "[\"\\🌀\"]"},
	{Kind: "n", Name: "string_incomplete_escape", Content: "[\"\\\"]"},
	{Kind: "n", Name: "string_incomplete_escaped_character", Content: "[\"\\u00A\"]"},
	{Kind: "n", Name: "string_incomplete_surrogate", Content: "[\"\\uD834\\uDd\"]"},
	{Kind: "n", Name: "string_incomplete_surrogate_escape_invalid", Content: "[\"\\uD800\\uD800\\x\"]"},
	{Kind: "n", Name: "string_invalid-utf-8-in-escape", Content: "[\"\\u\xe5\"]"},
	{Kind: "n", Name: "string_invalid_backslash_esc", Content: "[\"\\a\"]"},
	{Kind: "n", Name: "string_invalid_unicode_escape", Content: "[\"\\uqqqq\"]"},
	{Kind: "n", Name: "string_invalid_utf8_after_escape", Content: "[\"\\\xe5\"]"},
	{Kind: "n", Name: "string_leading_uescaped_thinspace", Content: "[\\u0020\"asd\"]"},
	{Kind: "n", Name: "string_no_quotes_with_bad_escape", Content: "[\\n]"},
	{Kind: "n", Name: "string_single_doublequote", Content: "\""},
	{Kind: "n", Name: "string_single_quote", Content: "['single quote']"},
	{Kind: "n", Name: "string_single_string_no_double_quotes", Content: "abc"},
	{Kind: "n", Name: "string_start_escape_unclosed", Content: "[\"\\"},
	{Kind: "n", Name: "string_unescaped_ctrl_char", Content: "[\"a\x00a\"]"},
	{Kind: "n", Name: "string_unescaped_newline", Content: "[\"new\r\nline\"]"},
	{Kind: "n", Name: "string_unescaped_tab", Content: "[\"\t\"]"},
	{Kind: "n", Name: "string_unicode_CapitalU", Content: "\"\\UA66D\""},
	{Kind: "n", Name: "string_with_trailing_garbage", Content: "\"\"x"},
	{Kind: "n", Name: "structure_U+2060_word_joined", Content: "[\u2060]"},
	{Kind: "n", Name: "structure_UTF8_BOM_no_data", Content: "\ufeff"},
	{Kind: "n", Name: "structure_angle_bracket_.", Content: "<.>"},
	{Kind: "n", Name: "structure_angle_bracket_null", Content: "[<null>]"},
	{Kind: "n", Name: "structure_array_trailing_garbage", Content: "[1]x"},
	{Kind: "n", Name: "structure_array_with_extra_array_close", Content: "[1]]"},
	{Kind: "n", Name: "structure_array_with_unclosed_string", Content: "[\"asd]"},
	{Kind: "n", Name: "structure_ascii-unicode-identifier", Content: "aå"},
	{Kind: "n", Name: "structure_capitalized_True", Content: "[True]"},
	{Kind: "n", Name: "structure_close_unopened_array", Content: "1]"},
	{Kind: "n", Name: "structure_comma_instead_of_closing_brace", Content: "{\"x\": true,"},
	{Kind: "n", Name: "structure_double_array", Content: "[][]"},
	{Kind: "n", Name: "structure_end_array", Content: "]"},
	{Kind: "n", Name: "structure_incomplete_UTF8_BOM", Content: "\xef\xbb{}"},
	{Kind: "n", Name: "structure_lone-invalid-utf-8", Content: "\xe5"},
	{Kind: "n", Name: "structure_lone-open-bracket", Content: "["},
	{Kind: "n", Name: "structure_no_data", Content: ""},
	{Kind: "n", Name: "structure_null-byte-outside-string", Content: "[\x00]"},
	{Kind: "n", Name: "structure_number_with_trailing_garbage", Content: "2@"},
	{Kind: "n", Name: "structure_object_followed_by_closing_object", Content: "{}}"},
	{Kind: "n", Name: "structure_object_unclosed_no_value", Content: "{\"\":"},
	{Kind: "n", Name: "structure_object_with_comment", Content: "{\"a\":/*comment*/\"b\"}"},
	{Kind: "n", Name: "structure_object_with_trailing_garbage", Content: "{\"a\": true} \"x\""},
	{Kind: "n", Name: "structure_open_array_apostrophe", Content: "['"},
	{Kind: "n", Name: "structure_open_array_comma", Content: "[,"},
	{Kind: "n", Name: "structure_open_array_open_object", Content: "[{"},
	{Kind: "n", Name: "structure_open_array_open_string", Content: "[\"a"},
	{Kind: "n", Name: "structure_open_array_string", Content: "[\"a\""},
	{Kind: "n", Name: "structure_open_object", Content: "{"},
	{Kind: "n", Name: "structure_open_object_close_array", Content: "{]"},
	{Kind: "n", Name: "structure_open_object_comma", Content: "{,"},
	{Kind: "n", Name: "structure_open_object_open_array", Content: "{["},
	{Kind: "n", Name: "structure_open_object_open_string", Content: "{\"a"},
	{Kind: "n", Name: "structure_open_object_string_with_apostrophes", Content: "{'a'"},
	{Kind: "n", Name: "structure_open_open", Content: "[\"\\{[\"\\{[\"\\{[\"\\{"},
	{Kind: "n", Name: "structure_single_eacute", Content: "\xe9"},
	{Kind: "n", Name: "structure_single_star", Content: "*"},
	{Kind: "n", Name: "structure_trailing_#", Content: "{\"a\":\"b\"}#{}"},
	{Kind: "n", Name: "structure_uescaped_LF_before_string", Content: "[\\u000A\"\"]"},
	{Kind: "n", Name: "structure_unclosed_array", Content: "[1"},
	{Kind: "n", Name: "structure_unclosed_array_partial_null", Content: "[ false, nul"},
	{Kind: "n", Name: "structure_unclosed_array_unfinished_false", Content: "[ true, fals"},
	{Kind: "n", Name: "structure_unclosed_array_unfinished_true", Content: "[ false, tru"},
	{Kind: "n", Name: "structure_unclosed_object", Content: "{\"asd\":\"asd\""},
	{Kind: "n", Name: "structure_unicode-identifier", Content: "å"},
	{Kind: "n", Name: "structure_whitespace_U+2060_word_joiner", Content: "[\u2060]"},
	{Kind: "n", Name: "structure_whitespace_formfeed", Content: "[\f]"},
	{Kind: "y", Name: "array_arraysWithSpaces", Content: "[[]   ]"},
	{Kind: "y", Name: "array_empty-string", Content: "[\"\"]"},
	{Kind: "y", Name: "array_empty", Content: "[]"},
	{Kind: "y", Name: "array_ending_with_newline", Content: "[\"a\"]"},
	{Kind: "y", Name: "array_false", Content: "[false]"},
	{Kind: "y", Name: "array_heterogeneous", Content: "[null, 1, \"1\", {}]"},
	{Kind: "y", Name: "array_null", Content: "[null]"},
	{Kind: "y", Name: "array_with_1_and_newline", Content: "[1\r\n]"},
	{Kind: "y", Name: "array_with_leading_space", Content: " [1]"},
	{Kind: "y", Name: "array_with_several_null", Content: "[1,null,null,null,2]"},
	{Kind: "y", Name: "array_with_trailing_space", Content: "[2] "},
	{Kind: "y", Name: "number", Content: "[123e65]"},
	{Kind: "y", Name: "number_0e+1", Content: "[0e+1]"},
	{Kind: "y", Name: "number_0e1", Content: "[0e1]"},
	{Kind: "y", Name: "number_after_space", Content: "[ 4]"},
	{Kind: "y", Name: "number_double_close_to_zero", Content: "[-0.000000000000000000000000000000000000000000000000000000000000000000000000000001]\r\n"},
	{Kind: "y", Name: "number_int_with_exp", Content: "[20e1]"},
	{Kind: "y", Name: "number_minus_zero", Content: "[-0]"},
	{Kind: "y", Name: "number_negative_int", Content: "[-123]"},
	{Kind: "y", Name: "number_negative_one", Content: "[-1]"},
	{Kind: "y", Name: "number_negative_zero", Content: "[-0]"},
	{Kind: "y", Name: "number_real_capital_e", Content: "[1E22]"},
	{Kind: "y", Name: "number_real_capital_e_neg_exp", Content: "[1E-2]"},
	{Kind: "y", Name: "number_real_capital_e_pos_exp", Content: "[1E+2]"},
	{Kind: "y", Name: "number_real_exponent", Content: "[123e45]"},
	{Kind: "y", Name: "number_real_fraction_exponent", Content: "[123.456e78]"},
	{Kind: "y", Name: "number_real_neg_exp", Content: "[1e-2]"},
	{Kind: "y", Name: "number_real_pos_exponent", Content: "[1e+2]"},
	{Kind: "y", Name: "number_simple_int", Content: "[123]"},
	{Kind: "y", Name: "number_simple_real", Content: "[123.456789]"},
	{Kind: "y", Name: "object", Content: "{\"asd\":\"sdf\", \"dfg\":\"fgh\"}"},
	{Kind: "y", Name: "object_basic", Content: "{\"asd\":\"sdf\"}"},
	{Kind: "y", Name: "object_duplicated_key", Content: "{\"a\":\"b\",\"a\":\"c\"}"},
	{Kind: "y", Name: "object_duplicated_key_and_value", Content: "{\"a\":\"b\",\"a\":\"b\"}"},
	{Kind: "y", Name: "object_empty", Content: "{}"},
	{Kind: "y", Name: "object_empty_key", Content: "{\"\":0}"},
	{Kind: "y", Name: "object_escaped_null_in_key", Content: "{\"foo\\u0000bar\": 42}"},
	{Kind: "y", Name: "object_extreme_numbers", Content: "{ \"min\": -1.0e+28, \"max\": 1.0e+28 }"},
	{Kind: "y", Name: "object_long_strings", Content: "{\"x\":[{\"id\": \"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\"}], \"id\": \"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\"}"},
	{Kind: "y", Name: "object_simple", Content: "{\"a\":[]}"},
	{Kind: "y", Name: "object_string_unicode", Content: "{\"title\":\"\\u041f\\u043e\\u043b\\u0442\\u043e\\u0440\\u0430 \\u0417\\u0435\\u043c\\u043b\\u0435\\u043a\\u043e\\u043f\\u0430\" }"},
	{Kind: "y", Name: "object_with_newlines", Content: "{\r\n\"a\": \"b\"\r\n}"},
	{Kind: "y", Name: "string_1_2_3_bytes_UTF-8_sequences", Content: "[\"\\u0060\\u012a\\u12AB\"]"},
	{Kind: "y", Name: "string_accepted_surrogate_pair", Content: "[\"\\uD801\\udc37\"]"},
	{Kind: "y", Name: "string_accepted_surrogate_pairs", Content: "[\"\\ud83d\\ude39\\ud83d\\udc8d\"]"},
	{Kind: "y", Name: "string_allowed_escapes", Content: "[\"\\\"\\\\\\/\\b\\f\\n\\r\\t\"]"},
	{Kind: "y", Name: "string_backslash_and_u_escaped_zero", Content: "[\"\\\\u0000\"]"},
	{Kind: "y", Name: "string_backslash_doublequotes", Content: "[\"\\\"\"]"},
	{Kind: "y", Name: "string_comments", Content: "[\"a/*b*/c/*d//e\"]"},
	{Kind: "y", Name: "string_double_escape_a", Content: "[\"\\\\a\"]"},
	{Kind: "y", Name: "string_double_escape_n", Content: "[\"\\\\n\"]"},
	{Kind: "y", Name: "string_escaped_control_character", Content: "[\"\\u0012\"]"},
	{Kind: "y", Name: "string_escaped_noncharacter", Content: "[\"\\uFFFF\"]"},
	{Kind: "y", Name: "string_in_array", Content: "[\"asd\"]"},
	{Kind: "y", Name: "string_in_array_with_leading_space", Content: "[ \"asd\"]"},
	{Kind: "y", Name: "string_last_surrogates_1_and_2", Content: "[\"\\uDBFF\\uDFFF\"]"},
	{Kind: "y", Name: "string_nbsp_uescaped", Content: "[\"new\\u00A0line\"]"},
	{Kind: "y", Name: "string_nonCharacterInUTF-8_U+10FFFF", Content: "[\"\U0010ffff\"]"},
	{Kind: "y", Name: "string_nonCharacterInUTF-8_U+FFFF", Content: "[\"\uffff\"]"},
	{Kind: "y", Name: "string_null_escape", Content: "[\"\\u0000\"]"},
	{Kind: "y", Name: "string_one-byte-utf-8", Content: "[\"\\u002c\"]"},
	{Kind: "y", Name: "string_pi", Content: "[\"π\"]"},
	{Kind: "y", Name: "string_reservedCharacterInUTF-8_U+1BFFF", Content: "[\"\U0001bfff\"]"},
	{Kind: "y", Name: "string_simple_ascii", Content: "[\"asd \"]"},
	{Kind: "y", Name: "string_space", Content: "\" \""},
	{Kind: "y", Name: "string_surrogates_U+1D11E_MUSICAL_SYMBOL_G_CLEF", Content: "[\"\\uD834\\uDd1e\"]"},
	{Kind: "y", Name: "string_three-byte-utf-8", Content: "[\"\\u0821\"]"},
	{Kind: "y", Name: "string_two-byte-utf-8", Content: "[\"\\u0123\"]"},
	{Kind: "y", Name: "string_u+2028_line_sep", Content: "[\"\u2028\"]"},
	{Kind: "y", Name: "string_u+2029_par_sep", Content: "[\"\u2029\"]"},
	{Kind: "y", Name: "string_uEscape", Content: "[\"\\u0061\\u30af\\u30EA\\u30b9\"]"},
	{Kind: "y", Name: "string_uescaped_newline", Content: "[\"new\\u000Aline\"]"},
	{Kind: "y", Name: "string_unescaped_char_delete", Content: "[\"\x7f\"]"},
	{Kind: "y", Name: "string_unicode", Content: "[\"\\uA66D\"]"},
	{Kind: "y", Name: "string_unicodeEscapedBackslash", Content: "[\"\\u005C\"]"},
	{Kind: "y", Name: "string_unicode_2", Content: "[\"⍂㈴⍂\"]"},
	{Kind: "y", Name: "string_unicode_U+10FFFE_nonchar", Content: "[\"\\uDBFF\\uDFFE\"]"},
	{Kind: "y", Name: "string_unicode_U+1FFFE_nonchar", Content: "[\"\\uD83F\\uDFFE\"]"},
	{Kind: "y", Name: "string_unicode_U+200B_ZERO_WIDTH_SPACE", Content: "[\"\\u200B\"]"},
	{Kind: "y", Name: "string_unicode_U+2064_invisible_plus", Content: "[\"\\u2064\"]"},
	{Kind: "y", Name: "string_unicode_U+FDD0_nonchar", Content: "[\"\\uFDD0\"]"},
	{Kind: "y", Name: "string_unicode_U+FFFE_nonchar", Content: "[\"\\uFFFE\"]"},
	{Kind: "y", Name: "string_unicode_escaped_double_quote", Content: "[\"\\u0022\"]"},
	{Kind: "y", Name: "string_utf8", Content: "[\"€𝄞\"]"},
	{Kind: "y", Name: "string_with_del_character", Content: "[\"a\x7fa\"]"},
	{Kind: "y", Name: "structure_lonely_false", Content: "false"},
	{Kind: "y", Name: "structure_lonely_int", Content: "42"},
	{Kind: "y", Name: "structure_lonely_negative_real", Content: "-0.1"},
	{Kind: "y", Name: "structure_lonely_null", Content: "null"},
	{Kind: "y", Name: "structure_lonely_string", Content: "\"asd\""},
	{Kind: "y", Name: "structure_lonely_true", Content: "true"},
	{Kind: "y", Name: "structure_string_empty", Content: "\"\""},
	{Kind: "y", Name: "structure_trailing_newline", Content: "[\"a\"]\r\n"},
	{Kind: "y", Name: "structure_true_in_array", Content: "[true]"},
	{Kind: "y", Name: "structure_whitespace_array", Content: " [] "},
}

func ExampleReadValue() {
	// ReadValue can parse any JSON value directly
	inputs := []string{
		`42`,
		`3.14159`,
		`"hello"`,
		`true`,
		`[1,2,3]`,
		`{"name":"John"}`,
	}

	for _, input := range inputs {
		scanner := NewScanner([]byte(input))
		value, err := ReadValue(scanner)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}
		fmt.Printf("%T: %v\n", value, value)
	}
	// Output:
	// float64: 42
	// float64: 3.14159
	// string: hello
	// bool: true
	// []interface {}: [1 2 3]
	// map[string]interface {}: map[name:John]
}

func ExampleReadObject_direct() {
	input := `{
		"name": "John",
		"age": 30,
		"hobbies": ["reading", "music"],
		"address": {
			"city": "New York",
			"zip": "10001"
		}
	}`

	scanner := NewScanner([]byte(input))
	obj, err := ReadObject(scanner)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("%#v\n", obj)

	// Output:
	// map[string]interface {}{"address":map[string]interface {}{"city":"New York", "zip":"10001"}, "age":30, "hobbies":[]interface {}{"reading", "music"}, "name":"John"}
}

func ExampleReadArray_direct() {
	input := `[
		42,
		"hello",
		true,
		{"key": "value"},
		[1, 2, 3]
	]`

	scanner := NewScanner([]byte(input))
	arr, err := ReadArray(scanner)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("%#v\n", arr)

	// Output:
	// []interface {}{42, "hello", true, map[string]interface {}{"key":"value"}, []interface {}{1, 2, 3}}
}

func ExampleReadObjectCallback() {
	input := `{
		"name": "John",
		"age": 30,
		"address": {
			"city": "New York",
			"zip": "10001",
			"location": {
				"lat": 40.7128,
				"lon": -74.0060
			}
		},
		"orders": [
			{"id": "A123", "total": 50.00},
			{"id": "B456", "total": 30.00}
		]
	}`

	scanner := NewScanner([]byte(input))
	err := ReadObjectCallback(scanner, func(key string, value any) error {
		switch key {
		case "name", "age":
			fmt.Printf("%s: %v\n", key, value)

		case "address":
			// Handle nested object
			if addr, ok := value.(map[string]any); ok {
				fmt.Printf("city: %v\n", addr["city"])
				// Handle deeply nested object
				if loc, ok := addr["location"].(map[string]any); ok {
					fmt.Printf("coordinates: %v,%v\n", loc["lat"], loc["lon"])
				}
			}

		case "orders":
			// Handle array of objects
			if orders, ok := value.([]any); ok {
				for _, order := range orders {
					if o, ok := order.(map[string]any); ok {
						fmt.Printf("order: %v = $%v\n", o["id"], o["total"])
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// Output:
	// name: John
	// age: 30
	// city: New York
	// coordinates: 40.7128,-74.006
	// order: A123 = $50
	// order: B456 = $30
}

func ExampleReadArrayCallback() {
	input := `[
		{"type": "user", "name": "John"},
		{"type": "order", "id": "A123"},
		{"type": "user", "name": "Jane"},
		{"type": "order", "id": "B456"}
	]`

	scanner := NewScanner([]byte(input))
	err := ReadArrayCallback(scanner, func(value any) error {
		// Type-check and process each array element
		if obj, ok := value.(map[string]any); ok {
			switch obj["type"] {
			case "user":
				fmt.Printf("Found user: %v\n", obj["name"])
			case "order":
				fmt.Printf("Found order: %v\n", obj["id"])
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// Output:
	// Found user: John
	// Found order: A123
	// Found user: Jane
	// Found order: B456
}
