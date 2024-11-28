# JSN

A high-performance JSON serialization package for Go. JSN offers both a
functional-style API and reflection-based marshaling for flexible and efficient
JSON generation and parsing.

## Features

* **High Performance:** Allows to reduce memory allocations for both reading and
  writing, making it suitable for performance-sensitive applications.

* **Flexible Reading API:** Process JSON using callbacks, maps, slices, or a mix
  of approaches for optimal control and efficiency. No struct tags are needed.

* **Flexible Writing API:** Fine-grained control over JSON structure through a
  callback-based approach or custom type marshaling.

* **Configurable Float Precision:** Control the precision of floating-point
  number formatting


## Reading JSON

JSN reads values through a Scanner object that takes input buffer and optional flags:

~~~go
scanner := jsn.NewScanner(buffer, /*<options>...*/)
~~~

By default, the scanner skips the BOM and initial whitespace, which can be disabled using the following options:

Flags:
- `jsn.ScannerFlagDoNotSkipBOM` - Do not skip the BOM at the start of the buffer
- `jsn.ScannerFlagDoNotSkipInitialWhitespace` - Do not skip initial whitespace at the start of the buffer

JSN provides two approaches to reading JSON:

1. Direct value reading - returns parsed values:
~~~go
// Read any JSON value:
value, err := jsn.ReadValue(scanner)    // returns any

// Read a JSON object:
input := `{
    "name": "John",
    "age": 30,
    "address": {
        "city": "New York",
        "location": {"lat": 40.7128, "lon": -74.0060}
    }
}`
obj, err := jsn.ReadObject(scanner)     // returns map[string]any

// Read a JSON array:
input := `[
    42,
    "hello",
    true,
    {"key": "value"},
    [1, 2, 3]
]`
arr, err := jsn.ReadArray(scanner)      // returns []any
~~~

2. Callback-based reading - for memory-efficient processing:
~~~go
// Process object fields selectively:
input := `{
    "name": "John",
    "age": 30,
    "address": {
        "city": "New York",
        "location": {"lat": 40.7128, "lon": -74.0060}
    }
}`
err := jsn.ReadObjectCallback(scanner, func(key string, value any) error {
    switch key {
    case "name":
        fmt.Printf("name: %v\n", value)
    case "address":
        if addr, ok := value.(map[string]any); ok {
            if loc, ok := addr["location"].(map[string]any); ok {
                fmt.Printf("coordinates: %v,%v\n", loc["lat"], loc["lon"])
            }
        }
    }
    return nil
})

// Filter array elements by type:
input := `[
    {"type": "user", "name": "John"},
    {"type": "order", "id": "A123"}
]`
err := jsn.ReadArrayCallback(scanner, func(value any) error {
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
~~~

Example of direct reading:
~~~go
func main() {
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
        scanner := jsn.NewScanner([]byte(input))
        value, err := jsn.ReadValue(scanner)
        if err != nil {
            // Handle error
            continue
        }
        // Values are automatically converted to appropriate Go types:
        // 42 -> float64: 42
        // 3.14159 -> float64: 3.14159
        // "hello" -> string: hello
        // true -> bool: true
        // [1,2,3] -> []interface {}: [1 2 3]
        // {"name":"John"} -> map[string]interface {}: map[name:John]
    }
}
~~~

## Writing JSON

The package provides flexible ways to write JSON through the `Marshal` function and custom marshalers.

### Supported Types

JSN supports direct marshaling of the following Go types without requiring custom marshalers:

Basic Types:
- `bool` - Marshaled as JSON boolean
- `string` - Marshaled as JSON string
- All numeric types (`int`, `int8`...`int64`, `uint`...`uint64`, `float32`, `float64`) - Marshaled as JSON numbers
- Custom types based on basic types (e.g., `type MyInt int`) - Automatically marshaled as their underlying type
- It is also possible to customize marshaling for basic types using the `StrMarshaler` interface.

Collection Types:
- `[]T` where T is any supported type - Marshaled as JSON arrays
- `map[string]T` where T is any supported type - Marshaled as JSON objects
- `[]byte` and `[N]byte` - Marshaled as JSON strings

Special Types:
- `nil` - Marshaled as JSON null
- `any` (interface{}) containing any supported type

Callback Types:
- `func(ArrayWriter)` - Marshaled as JSON arrays
- `func(ArrayWriter) error` - Marshaled as JSON arrays
- `func(ObjectWriter)` - Marshaled as JSON objects
- `func(ObjectWriter) error` - Marshaled as JSON objects

For other types (like structs), you need to implement one of the marshaler interfaces:
- `ObjMarshaler` for types that should be marshaled as JSON objects
- `ArrMarshaler` for types that should be marshaled as JSON arrays
- `StrMarshaler` for types that should be marshaled as JSON strings
- Types implementing `encoding.TextMarshaler` are supported and marshaled as strings.

### Basic Usage

JSN supports direct marshaling of primitive types and collections:

~~~go
// Marshal primitive types
str, _ := jsn.Marshal("hello")     // "hello"
num, _ := jsn.Marshal(42)          // 42
arr, _ := jsn.Marshal([]int{1,2})  // [1,2]

// Control float precision
pi, _ := jsn.Marshal(3.14159, jsn.FloatPrecision{Precision: 3})  // 3.14
~~~

### Custom Marshalers

Three interfaces are available for custom JSON serialization:

~~~go
// For custom string values
type StrMarshaler interface {
    MarshalJSN() (string, error)
}

// For custom array values
type ArrMarshaler interface {
    MarshalJSN(w ArrayWriter) error
}

// For custom object values
type ObjMarshaler interface {
    MarshalJSN(w ObjectWriter) error
}
~~~

Example usage:

~~~go
type Person struct {
    name string
    age  int
}

func (p Person) MarshalJSN(w jsn.ObjectWriter) error {
    w.Member("name", p.name)
    w.Member("age", p.age)
    return nil
}

person := Person{name: "John", age: 30}
result, _ := jsn.Marshal(person)  // {"name":"John","age":30}
~~~

### Functional Writing

The package supports a functional approach to writing JSON:

~~~go
// Write arrays using functions
writeNumbers := func(w jsn.ArrayWriter) error {
    w.Element(1)
    w.Element(2)
    return nil
}
result, _ := jsn.Marshal(writeNumbers)  // [1,2]

// Write objects using functions
writePerson := func(w jsn.ObjectWriter) error {
    w.Member("name", "John")
    w.Member("hobbies", []string{"reading"})
    return nil
}
result, _ := jsn.Marshal(writePerson)  // {"name":"John","hobbies":["reading"]}
~~~

### Nested Structures

Writers can be nested to create complex JSON structures. Here's an example of multi-level functional writing:

~~~go
result, _ := jsn.Marshal(func(w jsn.ObjectWriter) error {
    w.Member("name", "John")
    w.Member("address", func(w jsn.ObjectWriter) error {
        w.Member("street", "123 Main St")
        w.Member("city", "Springfield")
        return nil
    })
    w.Member("hobbies", func(w jsn.ArrayWriter) error {
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
// Output: {"name":"John","address":{"street":"123 Main St","city":"Springfield"},"hobbies":["reading","coding"],"scores":{"english":87,"math":95}}
~~~
