package jsn

import (
	"fmt"
	"reflect"
	"strings"
)

// StrMarshaler is implemented by types that can marshal themselves into a JSON string value.
// This is useful for types that need custom string representation in JSON.
type StrMarshaler interface {
	MarshalJSN() (string, error)
}

// ArrMarshaler is implemented by types that can marshal themselves into a JSON array.
// This interface provides full control over the array's JSON representation.
type ArrMarshaler interface {
	MarshalJSN(w ArrayWriter) error
}

// ObjMarshaler is implemented by types that can marshal themselves into a JSON object.
// This interface provides full control over the object's JSON representation.
type ObjMarshaler interface {
	MarshalJSN(w ObjectWriter) error
}

// ArrayWriter defines the interface for writing JSON arrays
type ArrayWriter interface {
	// Element writes supported value as an array element.
	Element(v any)
}

// ObjectWriter defines the interface for writing JSON objects
type ObjectWriter interface {
	// Member writes a key-value pair as an object member.
	Member(key string, v any)
}

// arrayWriter is the implementation of ArrayWriter interface
type arrayWriter struct {
	d              *decorator
	elementCounter int
}

// Value writes supported value as an array element.
func (w *arrayWriter) Element(v any) {
	w.d.arrayElement(w.elementCounter == 0)
	w.elementCounter++
	w.d.marshalValue(v)
}

// objectWriter is used to marshal objects into JSON.
type objectWriter struct {
	d            *decorator
	fieldCounter int
}

// Value writes any Go value as an array element.
func (w *objectWriter) Member(key string, v any) {
	w.d.objectField(key, w.fieldCounter == 0)
	w.fieldCounter++
	w.d.marshalValue(v)
}

// FloatPrecision specifies the number of decimal places to use when formatting floating-point numbers
type FloatPrecision struct {
	Precision int
}

func parseMarshalOptions(opts []any) (precision int, err error) {
	precision = 6

	for _, opt := range opts {
		switch v := opt.(type) {
		case FloatPrecision:
			if v.Precision < 0 {
				return 0, fmt.Errorf("invalid float precision: %d", v.Precision)
			}
			precision = v.Precision
		}
	}
	return precision, nil
}

// Marshal marshals any supported value into a JSON string.
func Marshal(v any, opts ...any) (string, error) {
	precision, err := parseMarshalOptions(opts)
	if err != nil {
		return "", err
	}

	striungBuilder := strings.Builder{}
	d := decorator{out: &striungBuilder, floatPrecision: precision}
	d.marshalValue(v)
	if d.err != nil {
		return "", d.err
	}
	return striungBuilder.String(), nil
}

// UnsupportedTypeError is returned when marshaling encounters a type
// that cannot be converted into JSON.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "jsn: unsupported type: " + e.Type.String()
}
