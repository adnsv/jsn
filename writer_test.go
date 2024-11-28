package jsn

import (
	"errors"
	"fmt"
	"testing"
)

func ExampleMarshal_primitives() {
	// Marshal primitive types
	fmt.Println(Marshal("hello")) // string
	fmt.Println(Marshal(42))      // int
	fmt.Println(Marshal(3.14))    // float
	fmt.Println(Marshal(true))    // bool
	fmt.Println(Marshal(nil))     // null
	// Output:
	// "hello" <nil>
	// 42 <nil>
	// 3.14 <nil>
	// true <nil>
	// null <nil>
}

func ExampleMarshal_array() {
	// Marshal arrays and slices
	numbers := []int{1, 2, 3}
	result, _ := Marshal(numbers)
	fmt.Println(result)
	// Output: [1,2,3]
}

// Define a custom type that implements ObjMarshaler
type Person struct {
	name string
	age  int
}

func (p Person) MarshalJSN(w ObjectWriter) error {
	w.Member("name", p.name)
	w.Member("age", p.age)
	return nil
}

func ExampleMarshal_customObject() {
	// Marshal custom object
	person := Person{name: "John", age: 30}
	result, _ := Marshal(person)
	fmt.Println(result)
	// Output: {"name":"John","age":30}
}

func ExampleMarshal_floatPrecision() {
	// Control float precision
	pi := 3.14159265359
	result, _ := Marshal(pi, FloatPrecision{Precision: 3})
	fmt.Println(result)
	// Output: 3.14
}

// Marshal nested structures
type ArrayContainer struct {
	data []int
}

func (c ArrayContainer) MarshalJSN(w ArrayWriter) error {
	for _, v := range c.data {
		w.Element(v)
	}
	return nil
}

func ExampleMarshal_nested() {
	container := ArrayContainer{data: []int{1, 2}}
	nested := []any{container, "str", 42}
	result, _ := Marshal(nested)
	fmt.Println(result)
	// Output: [[1,2],"str",42]
}

func ExampleMarshal_functionalArray() {
	result, _ := Marshal(func(w ArrayWriter) error {
		w.Element(1)
		w.Element(2)
		w.Element(3)
		return nil
	})
	fmt.Println(result)
	// Output: [1,2,3]
}

func ExampleMarshal_functionalObject() {
	result, _ := Marshal(func(w ObjectWriter) error {
		w.Member("name", "John")
		w.Member("age", 30)
		w.Member("hobbies", []string{"reading", "coding"})
		return nil
	})
	fmt.Println(result)
	// Output: {"name":"John","age":30,"hobbies":["reading","coding"]}
}

func ExampleMarshal_functionalNested() {
	result, _ := Marshal(func(w ObjectWriter) error {
		w.Member("name", "John")
		w.Member("address", func(w ObjectWriter) error {
			w.Member("street", "123 Main St")
			w.Member("city", "Springfield")
			return nil
		})
		w.Member("hobbies", func(w ArrayWriter) error {
			w.Element("reading")
			w.Element("coding")
			return nil
		})
		w.Member("scores", map[string]int{
			"math":    95,
			"english": 87,
		})
		return nil
	})
	fmt.Println(result)
	// Output: {"name":"John","address":{"street":"123 Main St","city":"Springfield"},"hobbies":["reading","coding"],"scores":{"english":87,"math":95}}
}

func TestMarshalPrimitives(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:  "string",
			input: "hello",
			want:  `"hello"`,
		},
		{
			name:  "number int",
			input: 42,
			want:  "42",
		},
		{
			name:  "number float",
			input: 3.14,
			want:  "3.14",
		},
		{
			name:  "boolean true",
			input: true,
			want:  "true",
		},
		{
			name:  "boolean false",
			input: false,
			want:  "false",
		},
		{
			name:  "null",
			input: nil,
			want:  "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalFloatPrecision(t *testing.T) {
	tests := []struct {
		name      string
		input     float64
		precision FloatPrecision
		want      string
		wantErr   bool
	}{
		{
			name:      "default precision (6 significant digits)",
			input:     3.14159265359,
			precision: FloatPrecision{Precision: 6},
			want:      "3.14159",
		},
		{
			name:      "custom precision 3 significant digits",
			input:     3.14159,
			precision: FloatPrecision{Precision: 3},
			want:      "3.14",
		},
		{
			name:      "large number with scientific notation",
			input:     1.23457e+08,
			precision: FloatPrecision{Precision: 6},
			want:      "1.23457e+08",
		},
		{
			name:      "small number",
			input:     0.0001234567,
			precision: FloatPrecision{Precision: 4},
			want:      "0.0001235",
		},
		{
			name:      "very large number",
			input:     1.23e+20,
			precision: FloatPrecision{Precision: 3},
			want:      "1.23e+20",
		},
		{
			name:      "very small number",
			input:     1.23e-20,
			precision: FloatPrecision{Precision: 3},
			want:      "1.23e-20",
		},
		{
			name:      "invalid negative precision",
			input:     3.14159,
			precision: FloatPrecision{Precision: -1},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input, tt.precision)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

type customStrMarshaler struct {
	value string
}

func (c customStrMarshaler) MarshalJSN() (string, error) {
	return c.value, nil
}

type customArrMarshaler struct {
	values []int
}

func (c customArrMarshaler) MarshalJSN(w ArrayWriter) error {
	for _, v := range c.values {
		w.Element(v)
	}
	return nil
}

type customObjMarshaler struct {
	name  string
	value int
}

func (c customObjMarshaler) MarshalJSN(w ObjectWriter) error {
	w.Member("name", c.name)
	w.Member("value", c.value)
	return nil
}

func TestMarshalCustom(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:  "StrMarshaler",
			input: customStrMarshaler{value: "custom"},
			want:  `"custom"`,
		},
		{
			name:  "ArrMarshaler",
			input: customArrMarshaler{values: []int{1, 2, 3}},
			want:  "[1,2,3]",
		},
		{
			name:  "ObjMarshaler",
			input: customObjMarshaler{name: "test", value: 42},
			want:  `{"name":"test","value":42}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:  "empty slice",
			input: []int{},
			want:  "[]",
		},
		{
			name:  "int slice",
			input: []int{1, 2, 3},
			want:  "[1,2,3]",
		},
		{
			name:  "string slice",
			input: []string{"a", "b", "c"},
			want:  `["a","b","c"]`,
		},
		{
			name:  "mixed slice",
			input: []any{1, "two", true},
			want:  `[1,"two",true]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalUnsupportedTypes(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "struct without marshaler",
			input: struct {
				Name string
				Age  int
			}{
				Name: "John",
				Age:  30,
			},
			wantErr: true,
		},
		{
			name: "struct with json tags",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			wantErr: true,
		},
		{
			name: "struct with jsn tags but no marshaler",
			input: struct {
				Name string `jsn:"name"`
				Age  int    `jsn:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			wantErr: true,
		},
		{
			name:    "channel",
			input:   make(chan int),
			wantErr: true,
		},
		{
			name:    "function",
			input:   func() {},
			wantErr: true,
		},
		{
			name:    "complex number",
			input:   complex(1, 2),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMarshalErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr string
	}{
		{
			name:    "error from ArrayWriter",
			input:   func(w ArrayWriter) error { return errors.New("array error") },
			wantErr: "array error",
		},
		{
			name:    "error from ObjectWriter",
			input:   func(w ObjectWriter) error { return errors.New("object error") },
			wantErr: "object error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Marshal(tt.input)
			if err == nil {
				t.Error("Marshal() expected error, got nil")
				return
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Marshal() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestMarshalNested(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name: "nested array",
			input: customArrMarshaler{
				values: []int{1, 2, 3},
			},
			want: "[1,2,3]",
		},
		{
			name: "array with custom elements",
			input: []customStrMarshaler{
				{value: "one"},
				{value: "two"},
			},
			want: `["one","two"]`,
		},
		{
			name: "object with array",
			input: struct {
				Arr customArrMarshaler
			}{
				Arr: customArrMarshaler{values: []int{1, 2}},
			},
			wantErr: true, // struct without marshaler interface
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper types for error testing
type errorStrMarshaler struct{ err error }

func (e errorStrMarshaler) MarshalJSN() (string, error) {
	return "", e.err
}

type errorArrMarshaler struct{ err error }

func (e errorArrMarshaler) MarshalJSN(w ArrayWriter) error {
	return e.err
}

type errorObjMarshaler struct{ err error }

func (e errorObjMarshaler) MarshalJSN(w ObjectWriter) error {
	return e.err
}

func TestMarshalCustomErrors(t *testing.T) {
	customErr := fmt.Errorf("custom error")
	tests := []struct {
		name    string
		input   any
		wantErr error
	}{
		{
			name:    "StrMarshaler error",
			input:   errorStrMarshaler{err: customErr},
			wantErr: customErr,
		},
		{
			name:    "ArrMarshaler error",
			input:   errorArrMarshaler{err: customErr},
			wantErr: customErr,
		},
		{
			name:    "ObjMarshaler error",
			input:   errorObjMarshaler{err: customErr},
			wantErr: customErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Marshal(tt.input)
			if err != tt.wantErr {
				t.Errorf("Marshal() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

type nestedObj struct {
	data string
}

// Implement ObjMarshaler for nestedObj
func (n nestedObj) MarshalJSN(w ObjectWriter) error {
	w.Member("data", n.data)
	return nil
}

type complexObj struct {
	nested []nestedObj
}

// Implement ObjMarshaler for complexObj
func (c complexObj) MarshalJSN(w ObjectWriter) error {
	w.Member("nested", c.nested)
	return nil
}

func TestMarshalDeepStructures(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name: "deeply nested array",
			input: []any{
				[]int{1, 2},
				[]string{"a", "b"},
				[]customStrMarshaler{
					{value: "deep"},
					{value: "structure"},
				},
			},
			want: `[[1,2],["a","b"],["deep","structure"]]`,
		},
		{
			name: "mixed nested objects",
			input: complexObj{
				nested: []nestedObj{
					{data: "first"},
					{data: "second"},
				},
			},
			want: `{"nested":[{"data":"first"},{"data":"second"}]}`,
		},
		{
			name: "array of custom objects",
			input: []customObjMarshaler{
				{name: "first", value: 1},
				{name: "second", value: 2},
			},
			want: `[{"name":"first","value":1},{"name":"second","value":2}]`,
		},
		{
			name: "object with array of arrays",
			input: matrixMarshaler{
				name: "matrix",
				matrix: [][]int{
					{1, 2},
					{3, 4},
				},
			},
			want: `{"name":"matrix","value":[[1,2],[3,4]]}`,
		},
		{
			name: "mixed types in array",
			input: []any{
				customStrMarshaler{value: "str"},
				customArrMarshaler{values: []int{1, 2}},
				customObjMarshaler{name: "obj", value: 42},
			},
			want: `["str",[1,2],{"name":"obj","value":42}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

type recursiveArr struct {
	values []any
}

func (r recursiveArr) MarshalJSN(w ArrayWriter) error {
	for _, v := range r.values {
		w.Element(v)
	}
	return nil
}

type recursiveObj struct {
	name     string
	children any
}

func (r recursiveObj) MarshalJSN(w ObjectWriter) error {
	w.Member("name", r.name)
	w.Member("children", r.children)
	return nil
}

func TestMarshalRecursive(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name: "recursive array structure",
			input: recursiveArr{
				values: []any{
					1,
					recursiveArr{values: []any{2, 3}},
					4,
				},
			},
			want: `[1,[2,3],4]`,
		},
		{
			name: "recursive object structure",
			input: recursiveObj{
				name: "root",
				children: recursiveObj{
					name:     "child",
					children: []int{1, 2, 3},
				},
			},
			want: `{"name":"root","children":{"name":"child","children":[1,2,3]}}`,
		},
		{
			name: "mixed recursive structures",
			input: recursiveObj{
				name: "root",
				children: recursiveArr{
					values: []any{
						recursiveObj{name: "child1", children: nil},
						recursiveObj{name: "child2", children: []int{1, 2}},
					},
				},
			},
			want: `{"name":"root","children":[{"name":"child1","children":null},{"name":"child2","children":[1,2]}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

type matrixMarshaler struct {
	name   string
	matrix [][]int
}

func (m matrixMarshaler) MarshalJSN(w ObjectWriter) error {
	w.Member("name", m.name)
	w.Member("value", m.matrix)
	return nil
}

func TestMarshalEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:  "empty array marshaler",
			input: customArrMarshaler{values: []int{}},
			want:  "[]",
		},
		{
			name:  "empty string marshaler",
			input: customStrMarshaler{value: ""},
			want:  `""`,
		},
		{
			name: "empty object marshaler",
			input: recursiveObj{
				name:     "",
				children: nil,
			},
			want: `{"name":"","children":null}`,
		},
		{
			name: "deeply nested empty structures",
			input: recursiveObj{
				name: "root",
				children: recursiveArr{
					values: []any{
						recursiveObj{name: "", children: []any{}},
						recursiveArr{values: []any{}},
						nil,
					},
				},
			},
			want: `{"name":"root","children":[{"name":"","children":[]},[],null]}`,
		},
		{
			name:  "special characters in strings",
			input: customStrMarshaler{value: "\"\\\b\f\n\r\t\u0000\u001f"},
			want:  `"\"\\\b\f\n\r\t\u0000\u001f"`,
		},
		{
			name:  "unicode characters",
			input: customStrMarshaler{value: "Hello, 世界"},
			want:  `"Hello, 世界"`,
		},
		{
			name: "max float values",
			input: []float64{
				1.7976931348623157e+308,  // max float64
				-1.7976931348623157e+308, // min float64
				4.9406564584124654e-324,  // smallest positive float64
			},
			want: "[1.79769e+308,-1.79769e+308,4.94066e-324]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}
