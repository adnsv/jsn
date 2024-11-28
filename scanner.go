package jsn

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrUnexpectedToken        = errors.New("unexpected token")
	ErrUnexpectedEOF          = errors.New("unexpected EOF")
	ErrInvalidNumber          = errors.New("invalid number")
	ErrInvalidString          = errors.New("invalid string")
	ErrInvalidUnicodeEscape   = errors.New("invalid unicode escape")
	ErrNumericValueOutOfRange = errors.New("numeric value out of range")
)

type ScannerFlag int

const (
	ScannerFlagDoNotSkipBOM ScannerFlag = 1 << iota
	ScannerFlagDoNotSkipInitialWhitespace
)

// Scanner is a simple parser for JSON data
type Scanner struct {
	data  []byte
	cur   int
	flags ScannerFlag
}

// NewScanner creates a new scanner and skips the BOM and optional whitespace at
// the start of the data
func NewScanner(data []byte, opts ...any) *Scanner {
	s := &Scanner{data: data}
	for _, opt := range opts {
		switch v := opt.(type) {
		case ScannerFlag:
			s.flags |= v
		default:
			panic(fmt.Sprintf("jsn: unsupported scanner option type: %T", v))
		}
	}
	if s.flags&ScannerFlagDoNotSkipBOM == 0 {
		s.SkipBOM()
	}
	if s.flags&ScannerFlagDoNotSkipInitialWhitespace == 0 {
		s.skipWhitespace()
	}
	return s
}

// IsEOF returns true if the scanner has reached the end of input
func (s *Scanner) IsEOF() bool {
	return s.cur >= len(s.data)
}

// SkipBOM skips the UTF-8 Byte Order Mark (BOM) if present at the start of the data
func (s *Scanner) SkipBOM() bool {
	// UTF-8 BOM is bytes: 0xEF, 0xBB, 0xBF
	if len(s.data) >= 3 &&
		s.data[0] == 0xEF &&
		s.data[1] == 0xBB &&
		s.data[2] == 0xBF {
		s.cur += 3
		return true
	}
	return false
}

// Finalize ensures that the scanner has consumed all input
func (s *Scanner) Finalize() error {
	s.skipWhitespace()
	if !s.IsEOF() {
		return ErrUnexpectedToken
	}
	return nil
}

func (s *Scanner) next() byte {
	if s.cur >= len(s.data) {
		return 0
	}
	c := s.data[s.cur]
	s.cur++
	return c
}

func (s *Scanner) peek() byte {
	if s.cur >= len(s.data) {
		return 0
	}
	return s.data[s.cur]
}

func (s *Scanner) skipByte(b byte) bool {
	if s.cur >= len(s.data) {
		return false
	}
	if s.data[s.cur] == b {
		s.cur++
		return true
	}
	return false
}

func (s *Scanner) skipWhitespace() {
	for s.cur < len(s.data) {
		c := s.data[s.cur]
		// In strict JSON, only space, tab, CR, and LF are allowed as whitespace
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			s.cur++
			continue
		}
		return
	}
}

func (s *Scanner) isDecimalDigit() bool {
	return s.cur < len(s.data) && s.data[s.cur] >= '0' && s.data[s.cur] <= '9'
}

func (s *Scanner) skipDecimalDigits() bool {
	startPos := s.cur
	for s.isDecimalDigit() {
		s.cur++
	}
	return s.cur > startPos
}

func (s *Scanner) skipSequence(seq []byte) bool {
	if s.cur+len(seq) > len(s.data) {
		return false
	}
	for i, b := range seq {
		if s.data[s.cur+i] != b {
			return false
		}
	}
	s.cur += len(seq)
	return true
}

func (s *Scanner) parseString() (string, error) {
	if s.peek() != '"' {
		return "", ErrUnexpectedToken
	}
	s.cur++

	start := s.cur
	escaped := false

	// Fast path for unescaped strings
	for s.cur < len(s.data) {
		c := s.data[s.cur]
		if c <= 0x1F {
			return "", ErrInvalidString
		}
		if c == '\\' {
			escaped = true
			break
		}
		if c == '"' {
			// notice that this always creates a new string and copies the data,
			// while this is not the fastest approach, it also has benefits  avoids holding references to the original data.
			result := string(s.data[start:s.cur])
			s.cur++
			return result, nil
		}
		s.cur++
	}

	// If we get here without finding a closing quote
	if !escaped {
		return "", ErrInvalidString
	}

	// Slow path for escaped strings
	s.cur = start
	var buf []byte
	for {
		c := s.next()
		if c == 0 {
			return "", ErrInvalidString
		}
		if c <= 0x1F {
			return "", ErrInvalidString
		}
		if c == '"' {
			break
		}
		if c == '\\' {
			if s.cur >= len(s.data) {
				return "", ErrInvalidString
			}
			c = s.peek()
			switch c {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't', 'u':
				s.cur++
				switch c {
				case '"', '\\', '/':
					buf = append(buf, c)
				case 'b':
					buf = append(buf, '\b')
				case 'f':
					buf = append(buf, '\f')
				case 'n':
					buf = append(buf, '\n')
				case 'r':
					buf = append(buf, '\r')
				case 't':
					buf = append(buf, '\t')
				case 'u':
					r, err := s.parseUnicode()
					if err != nil {
						return "", err
					}
					buf = append(buf, string(r)...)
				}
			default:
				return "", ErrInvalidString
			}
		} else {
			buf = append(buf, c)
		}
	}
	return string(buf), nil
}

func (s *Scanner) parseUnicode() (rune, error) {
	if len(s.data) < s.cur+4 {
		return 0, ErrInvalidUnicodeEscape
	}

	hex := string(s.data[s.cur : s.cur+4])
	s.cur += 4

	v, err := strconv.ParseUint(hex, 16, 16)
	if err != nil {
		return 0, ErrInvalidUnicodeEscape
	}

	return rune(v), nil
}

func (s *Scanner) parseNumber() (float64, error) {
	start := s.cur

	// Optional minus
	s.skipByte('-')

	// Integer part
	if s.skipByte('0') {
		if s.isDecimalDigit() {
			return 0, ErrInvalidNumber
		}
	} else {
		if s.cur >= len(s.data) || s.data[s.cur] < '1' || s.data[s.cur] > '9' {
			return 0, ErrInvalidNumber
		}
		s.cur++
		s.skipDecimalDigits()
	}

	// Fractional part
	if s.skipByte('.') {
		if !s.skipDecimalDigits() {
			return 0, ErrInvalidNumber
		}
		// After a valid decimal part, another dot is an error
		if s.skipByte('.') {
			return 0, ErrInvalidNumber
		}
	}

	// Exponent part
	if s.skipByte('e') || s.skipByte('E') {
		if !s.skipByte('+') {
			s.skipByte('-')
		}
		if !s.skipDecimalDigits() {
			return 0, ErrInvalidNumber
		}
		// After a valid exponent, another exponent is an error
		if s.skipByte('e') || s.skipByte('E') {
			return 0, ErrInvalidNumber
		}
	}

	num := string(s.data[start:s.cur])
	val, err := strconv.ParseFloat(num, 64)
	if err != nil {
		if numError := err.(*strconv.NumError); numError.Err == strconv.ErrRange {
			return 0, ErrNumericValueOutOfRange
		}
		return 0, ErrInvalidNumber
	}

	return val, nil
}
