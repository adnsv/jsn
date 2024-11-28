package jsn

import (
	"encoding"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strconv"
)

// decorator handles the low-level writing of JSON values with proper formatting.
type decorator struct {
	out            io.Writer // The underlying writer where JSON output is written
	floatPrecision int       // Precision used when formatting floating-point numbers
	err            error     // Whether an error has occurred
}

// handleError sets the error if it hasn't been set yet.
func (d *decorator) handleError(e error) {
	if d.err == nil {
		d.err = e
	}
}

// hadError returns whether an error has occurred.
func (d *decorator) hadError() bool {
	return d.err != nil
}

// put writes a string to the underlying writer.
func (d *decorator) put(s string) {
	if d.err != nil {
		return // block output if an error has occurred
	}
	_, err := d.out.Write([]byte(s))
	if err != nil {
		d.handleError(err)
	}
}

func (d *decorator) marshalNull() {
	d.put("null")
}

func (d *decorator) marshalBool(v bool) {
	if v {
		d.put("true")
	} else {
		d.put("false")
	}
}

func (d *decorator) marshalFloat64(v float64) {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		d.handleError(fmt.Errorf("unsupported float value: %v", v))
	}
	d.put(strconv.FormatFloat(v, 'g', d.floatPrecision, 64))
}

func (d *decorator) marshalString(v string) {
	d.put("\"")
	d.scrambleStr(v)
	d.put("\"")
}

// Object handling methods
func (d *decorator) objectBegin() {}

func (d *decorator) objectField(name string, first bool) {
	if first {
		d.put("{\"")
	} else {
		d.put(",\"")
	}
	d.scrambleStr(name)
	d.put("\":")
}

func (d *decorator) objectEnd(wasEmpty bool) {
	if wasEmpty {
		d.put("{}")
	} else {
		d.put("}")
	}
}

func (d *decorator) marshalObj(m ObjMarshaler) {
	if d.hadError() {
		return // early exit if an error has occurred
	}

	d.objectBegin()
	ow := objectWriter{d: d}
	err := m.MarshalJSN(&ow)
	if err != nil {
		d.handleError(err)
	}
	d.objectEnd(ow.fieldCounter == 0)
}

// Array handling methods
func (d *decorator) arrayBegin() {}

func (d *decorator) arrayElement(first bool) {
	if first {
		d.put("[")
	} else {
		d.put(",")
	}
}

func (d *decorator) arrayEnd(wasEmpty bool) {
	if wasEmpty {
		d.put("[]")
	} else {
		d.put("]")
	}
}

func (d *decorator) marshalArr(m ArrMarshaler) {
	if d.hadError() {
		return // early exit if an error has occurred
	}
	d.arrayBegin()
	aw := arrayWriter{d: d}
	err := m.MarshalJSN(&aw)
	if err != nil {
		d.handleError(err)
	}
	d.arrayEnd(aw.elementCounter == 0)
}

// Complex value handling
func (d *decorator) marshalValue(v any) {
	if d.hadError() {
		return // early exit if an error has occurred
	}

	val := reflect.ValueOf(v)

	// JSON null
	if !val.IsValid() {
		d.marshalNull()
		return
	}
	if val.Kind() == reflect.Ptr && val.IsNil() {
		d.marshalNull()
	}

	// Handle functional inputs
	switch typ := v.(type) {
	case func(ArrayWriter):
		d.arrayBegin()
		aw := arrayWriter{d: d}
		typ(&aw)
		d.arrayEnd(aw.elementCounter == 0)
		return

	case func(ArrayWriter) error:
		d.arrayBegin()
		aw := arrayWriter{d: d}
		err := typ(&aw)
		if err != nil {
			d.handleError(err)
			return
		}
		d.arrayEnd(aw.elementCounter == 0)
		return

	case func(ObjectWriter):
		d.objectBegin()
		ow := objectWriter{d: d}
		typ(&ow)
		d.objectEnd(ow.fieldCounter == 0)
		return

	case func(ObjectWriter) error:
		d.objectBegin()
		ow := objectWriter{d: d}
		err := typ(&ow)
		if err != nil {
			d.handleError(err)
			return
		}
		d.objectEnd(ow.fieldCounter == 0)
		return
	}

	for val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		if val.IsNil() {
			// Null interface or pointer
			d.marshalNull()
			return
		}
		val = val.Elem()
	}

	typ := val.Type()

	if val.CanInterface() {
		if typ.Implements(objMarshalerType) {
			d.marshalObj(val.Interface().(ObjMarshaler))
			return
		} else if typ.Implements(arrMarshalerType) {
			d.marshalArr(val.Interface().(ArrMarshaler))
			return
		} else if typ.Implements(strMarshalerType) {
			s, err := val.Interface().(StrMarshaler).MarshalJSN()
			if err != nil {
				d.handleError(err)
				return
			}
			d.marshalString(s)
			return
		} else if typ.Implements(textMarshalerType) {
			s, err := val.Interface().(encoding.TextMarshaler).MarshalText()
			if err != nil {
				d.handleError(err)
			}
			d.marshalString(string(s))
			return
		}
	}

	if val.CanAddr() {
		pv := val.Addr()
		if pv.CanInterface() {
			if pv.Type().Implements(objMarshalerType) {
				d.marshalObj(pv.Interface().(ObjMarshaler))
				return
			} else if pv.Type().Implements(arrMarshalerType) {
				d.marshalArr(pv.Interface().(ArrMarshaler))
				return
			} else if pv.Type().Implements(strMarshalerType) {
				s, err := pv.Interface().(StrMarshaler).MarshalJSN()
				if err != nil {
					d.handleError(err)
					return
				}
				d.marshalString(s)
				return
			} else if pv.Type().Implements(textMarshalerType) {
				s, err := pv.Interface().(encoding.TextMarshaler).MarshalText()
				if err != nil {
					d.handleError(err)
					return
				}
				d.marshalString(string(s))
				return
			}
		}
	}

	k := val.Kind()
	if (k == reflect.Slice || val.Kind() == reflect.Array) && typ.Elem().Kind() != reflect.Uint8 {
		d.arrayBegin()
		for i, n := 0, val.Len(); i < n; i++ {
			d.arrayElement(i == 0)
			d.marshalValue(val.Index(i).Interface())
			if d.hadError() {
				return // early exit if an error has occurred
			}
		}
		d.arrayEnd(val.Len() == 0)
		return
	}

	// TODO: keys convertible to string
	if k == reflect.Map && typ.Key().Kind() == reflect.String {
		type pair struct {
			k string
			v reflect.Value
		}
		pairs := make([]pair, val.Len())
		mi := val.MapRange()
		for i := 0; mi.Next(); i++ {
			pairs[i].k = mi.Key().String()
			pairs[i].v = mi.Value()
		}

		// coming from a map, the only way to produce a stable repeatable output
		// is to sort the keys
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })

		d.objectBegin()
		for i, kv := range pairs {
			d.objectField(kv.k, i == 0)
			d.marshalValue(kv.v.Interface())
			if d.hadError() {
				return // early exit if an error has occurred
			}
		}
		d.objectEnd(val.Len() == 0)
		return
	}

	// simple types
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d.put(strconv.FormatInt(val.Int(), 10))
		return
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		d.put(strconv.FormatUint(val.Uint(), 10))
		return
	case reflect.Float32, reflect.Float64:
		d.marshalFloat64(val.Float())
		return
	case reflect.String:
		d.marshalString(val.String())
		return
	case reflect.Bool:
		d.marshalBool(val.Bool())
		return
	case reflect.Array:
		if typ.Elem().Kind() != reflect.Uint8 {
			break
		}
		// [...]byte
		var bytes []byte
		if val.CanAddr() {
			bytes = val.Slice(0, val.Len()).Bytes()
		} else {
			bytes = make([]byte, val.Len())
			reflect.Copy(reflect.ValueOf(bytes), val)
		}
		d.marshalString(string(bytes))
		return

	case reflect.Slice:
		if typ.Elem().Kind() != reflect.Uint8 {
			break
		}
		d.marshalString(string(val.Bytes()))
		return
	}
	d.handleError(&UnsupportedTypeError{typ})
}

// String handling utilities
func (d *decorator) scrambleStr(s string) {
	if s == "" || d.hadError() {
		return
	}
	c, b := 0, 0
	e := len(s)

	replace := func(with string) {
		d.put(s[b:c])
		c++
		b = c
		d.put(with)
	}

	for c != e {
		cp := s[c]
		switch cp {
		case '\b':
			replace("\\b")
		case '\f':
			replace("\\f")
		case '\n':
			replace("\\n")
		case '\r':
			replace("\\r")
		case '\t':
			replace("\\t")
		case '\\':
			replace("\\\\")
		case '"':
			replace("\\\"")
		default:
			if cp <= 0x0f {
				with := []byte("\\u0000")
				with[5] += cp
				if cp >= 0x0a {
					with[5] += 'a' - ':'
				}
				replace(string(with))
			} else if cp <= 0x1f {
				with := []byte("\\u0010")
				with[5] += cp - 16
				if cp >= 0x1a {
					with[5] += 'a' - ':'
				}
				replace(string(with))
			} else {
				c++
			}
		}
	}

	if c != b {
		d.put(s[b:c])
	}
}

var (
	strMarshalerType  = reflect.TypeOf((*StrMarshaler)(nil)).Elem()
	objMarshalerType  = reflect.TypeOf((*ObjMarshaler)(nil)).Elem()
	arrMarshalerType  = reflect.TypeOf((*ArrMarshaler)(nil)).Elem()
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)
