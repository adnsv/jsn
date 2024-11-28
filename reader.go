package jsn

// ReadObjectCallback reads a JSON object and invokes the callback function for each key-value pair.
// The callback receives the key as a string and the value as an interface{}.
// This allows for memory-efficient processing of JSON objects without storing the entire structure.
//
// Example:
//
//	err := ReadObjectCallback(scanner, func(key string, value any) error {
//	    if key == "name" {
//	        fmt.Printf("name: %v\n", value)
//	    }
//	    return nil
//	})
func ReadObjectCallback(s *Scanner, callback func(k string, v any) error) error {
	if !s.skipByte('{') {
		return ErrUnexpectedToken
	}

	s.skipWhitespace()
	if s.skipByte('}') {
		return nil
	}

	var err error
	var key string
	var value any

	for {
		// Parse key
		s.skipWhitespace()
		key, err = s.parseString()
		if err != nil {
			return err
		}

		s.skipWhitespace()
		if !s.skipByte(':') {
			return ErrUnexpectedToken
		}

		// Parse value
		s.skipWhitespace()
		value, err = ReadValue(s)
		if err != nil {
			return err
		}
		err = callback(key, value)
		if err != nil {
			return err
		}

		s.skipWhitespace()
		if s.IsEOF() {
			return ErrUnexpectedEOF
		}
		if s.skipByte(',') {
			continue
		}
		if s.skipByte('}') {
			return nil
		}
		return ErrUnexpectedToken
	}
}

// ReadObject reads a JSON object and returns it as map[string]any
func ReadObject(s *Scanner) (map[string]any, error) {
	m := make(map[string]any)
	err := ReadObjectCallback(s, func(key string, value any) error {
		m[key] = value
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ReadValue reads any JSON value and returns it as a Go value.
// The mapping of JSON types to Go types is as follows:
//   - JSON null -> nil
//   - JSON boolean -> bool
//   - JSON number -> float64
//   - JSON string -> string
//   - JSON array -> []any
//   - JSON object -> map[string]any
//
// This function is recursive and will handle nested structures of any depth,
// limited only by available stack space.
func ReadValue(s *Scanner) (any, error) {
	s.skipWhitespace()

	if s.IsEOF() {
		return nil, ErrUnexpectedEOF
	}

	switch s.peek() {
	case '{':
		s.cur++
		m := make(map[string]any)
		s.skipWhitespace()
		if s.skipByte('}') {
			return m, nil
		}
		for {
			s.skipWhitespace()
			if s.IsEOF() {
				return nil, ErrUnexpectedEOF
			}
			// Key must be a string in strict JSON
			key, err := s.parseString()
			if err != nil {
				return nil, err
			}

			s.skipWhitespace()
			if s.IsEOF() {
				return nil, ErrUnexpectedEOF
			}
			if !s.skipByte(':') {
				return nil, ErrUnexpectedToken
			}

			val, err := ReadValue(s)
			if err != nil {
				return nil, err
			}
			m[key] = val

			s.skipWhitespace()
			if s.IsEOF() {
				return nil, ErrUnexpectedEOF
			}
			if s.skipByte('}') {
				return m, nil
			}
			if !s.skipByte(',') {
				return nil, ErrUnexpectedToken
			}
		}

	case '[':
		s.cur++
		var arr []any
		s.skipWhitespace()
		if s.skipByte(']') {
			return arr, nil
		}
		for {
			if s.IsEOF() {
				return nil, ErrUnexpectedEOF
			}
			val, err := ReadValue(s)
			if err != nil {
				return nil, err
			}
			arr = append(arr, val)

			s.skipWhitespace()
			if s.IsEOF() {
				return nil, ErrUnexpectedEOF
			}
			if s.skipByte(']') {
				return arr, nil
			}
			if !s.skipByte(',') {
				return nil, ErrUnexpectedToken
			}
			s.skipWhitespace()
		}

	case '"':
		return s.parseString()

	case 't':
		if !s.skipSequence([]byte("true")) {
			return nil, ErrUnexpectedToken
		}
		return true, nil

	case 'f':
		if !s.skipSequence([]byte("false")) {
			return nil, ErrUnexpectedToken
		}
		return false, nil

	case 'n':
		if !s.skipSequence([]byte("null")) {
			return nil, ErrUnexpectedToken
		}
		return nil, nil

	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return s.parseNumber()

	default:
		return nil, ErrUnexpectedToken
	}
}

// ReadArrayCallback reads a JSON array and invokes the callback function for each element.
// This allows for memory-efficient processing of arrays without storing the entire structure.
//
// Example:
//
//	err := ReadArrayCallback(scanner, func(value any) error {
//	    fmt.Printf("value: %v\n", value)
//	    return nil
//	})
func ReadArrayCallback(s *Scanner, callback func(any) error) error {
	if !s.skipByte('[') {
		return ErrUnexpectedToken
	}

	s.skipWhitespace()
	if s.skipByte(']') {
		return nil
	}

	for {
		s.skipWhitespace()
		if s.IsEOF() {
			return ErrUnexpectedEOF
		}
		value, err := ReadValue(s)
		if err != nil {
			return err
		}

		if err := callback(value); err != nil {
			return err
		}

		s.skipWhitespace()
		if s.IsEOF() {
			return ErrUnexpectedEOF
		}
		if s.skipByte(',') {
			continue
		}
		if s.skipByte(']') {
			return nil
		}
		return ErrUnexpectedToken
	}
}

// ReadArray reads a JSON array and returns it as []any
func ReadArray(s *Scanner) ([]any, error) {
	var arr []any
	err := ReadArrayCallback(s, func(value any) error {
		arr = append(arr, value)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return arr, nil
}
