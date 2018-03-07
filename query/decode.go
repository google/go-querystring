package query

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// A Decoder provides query value mapping to go struct.
type Decoder struct {
	Query url.Values
}

// NewDecoder creates new Decoder struct.
func NewDecoder(q url.Values) *Decoder {
	return &Decoder{Query: q}
}

type decodeContext struct {
	sv    reflect.Value
	scope string
}

// Decode Query and map to go struct.
func (d *Decoder) Decode(v interface{}) (err error) {
	sv := reflect.ValueOf(v)
	if sv.Kind() != reflect.Ptr || sv.IsNil() {
		return &InvalidDecodeError{reflect.TypeOf(v)}
	}

	_, err = d.decode(decodeContext{sv: sv.Elem()}, nil)
	return
}

func (d *Decoder) decode(c decodeContext, opts tagOptions) (bool, error) {
	switch c.sv.Kind() {
	case reflect.Chan, reflect.Complex64, reflect.Complex128, reflect.Invalid, reflect.Func, reflect.UnsafePointer, reflect.Uintptr:
		return false, &UnsupportedTypeError{Type: c.sv.Type()}
	case reflect.Bool:
		return d.decodeBool(c, opts)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return d.decodeInt(c)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return d.decodeUint(c)
	case reflect.Float32, reflect.Float64:
		return d.decodeFloat(c)
	case reflect.Array:
		return d.decodeSlice(decodeContext{sv: c.sv.Slice(0, c.sv.Len()), scope: c.scope}, opts)
	case reflect.Slice:
		return d.decodeSlice(c, opts)
	case reflect.Map:
		return d.decodeMap(c)
	case reflect.Interface:
		sv := reflect.New(c.sv.Elem().Type()).Elem()
		setOK, err := d.decode(decodeContext{sv: sv, scope: c.scope}, opts)
		if err != nil {
			return false, err
		}
		if setOK {
			c.sv.Set(sv)
			return true, nil
		}
		return false, nil
	case reflect.Ptr:
		sv := reflect.New(c.sv.Type().Elem())
		setOK, err := d.decode(decodeContext{sv: sv.Elem(), scope: c.scope}, opts)
		if err != nil {
			return false, err
		}
		if setOK {
			c.sv.Set(sv)
			return true, nil
		}
		return false, nil
	case reflect.String:
		v := d.decodeString(c)
		return v, nil
	case reflect.Struct:
		return d.decodeStruct(c, opts)
	}

	return false, fmt.Errorf("Unknown Type: %s", c.sv.Kind().String())
}

func (d *Decoder) decodeBool(c decodeContext, opts tagOptions) (bool, error) {
	if _, ok := d.Query[c.scope]; !ok {
		return false, nil
	}

	trueValue, falseValue := "true", "false"
	if opts.Contains("int") {
		trueValue, falseValue = "1", "0"
	}

	s := d.Query.Get(c.scope)
	if s == trueValue {
		c.sv.SetBool(true)
		return true, nil
	} else if s == falseValue {
		c.sv.SetBool(false)
		return true, nil
	}
	return false, &InvalidBooleanValueError{
		Query:      d.Query,
		Key:        c.scope,
		QueryValue: s,
		TrueValue:  trueValue,
		FalseValue: falseValue,
	}
}

func (d *Decoder) decodeInt(c decodeContext) (bool, error) {
	if _, ok := d.Query[c.scope]; !ok {
		return false, nil
	}

	s := d.Query.Get(c.scope)
	i, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return false, errors.Wrapf(err, "decodeInt(%#v)", c.scope)
	}

	c.sv.SetInt(i)
	return true, nil
}

func (d *Decoder) decodeUint(c decodeContext) (bool, error) {
	if _, ok := d.Query[c.scope]; !ok {
		return false, nil
	}

	s := d.Query.Get(c.scope)
	i, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return false, errors.Wrapf(err, "decodeUint(%#v)", c.scope)
	}

	c.sv.SetUint(i)
	return true, nil
}

func (d *Decoder) decodeFloat(c decodeContext) (bool, error) {
	s := d.Query.Get(c.scope)
	if s == "" {
		return false, nil
	}

	bitSize := 64
	if c.sv.Kind() == reflect.Float32 {
		bitSize = 32
	}

	f, err := strconv.ParseFloat(s, bitSize)
	if err != nil {
		return false, errors.Wrapf(err, "decodeFloat(%#v)", c.scope)
	}

	c.sv.SetFloat(f)
	return true, nil
}

func (d *Decoder) decodeSlice(c decodeContext, opts tagOptions) (bool, error) {
	typ := c.sv.Type()

	// delimiter slice
	var del string
	if opts.Contains("comma") {
		del = ","
	} else if opts.Contains("space") {
		del = " "
	} else if opts.Contains("semicolon") {
		del = ";"
	}
	if del != "" {
		if _, ok := d.Query[c.scope]; !ok {
			return false, nil
		}

		s := d.Query.Get(c.scope)
		parts := strings.Split(s, del)
		if len(parts) == 0 {
			return true, nil
		}

		av, err := decodeSliceParts(typ, parts, opts)
		if err != nil {
			return false, errors.Wrapf(err, "decodeSlice(%#v)", c.scope)
		}

		if c.sv.CanSet() {
			c.sv.Set(av)
		} else {
			reflect.Copy(c.sv, av)
		}
		return true, nil
	}

	// numbered
	if opts.Contains("numbered") {
		return d.decodeNumberedSlice(c, opts)
	}

	// branckets or default
	suffix := ""
	if opts.Contains("brackets") {
		suffix = "[]"
	}

	// parse only non-nested
	if parts, ok := d.Query[c.scope+suffix]; !ok {
		return false, nil
	} else if len(parts) == 0 {
		return true, nil
	} else {
		av, err := decodeSliceParts(typ, parts, opts)
		if err != nil {
			return false, errors.Wrapf(err, "decodeSlice(%#v)", c.scope)
		}

		if c.sv.CanSet() {
			c.sv.Set(av)
		} else {
			reflect.Copy(c.sv, av)
		}
		return true, nil
	}
}

func (d *Decoder) decodeNumberedSlice(c decodeContext, opts tagOptions) (bool, error) {
	av := c.sv

	var setCount int
	for key, values := range d.Query {
		if !strings.HasPrefix(key, c.scope) {
			continue
		}

		// get index number
		s := strings.TrimPrefix(key, c.scope)
		i, err := strconv.Atoi(s)
		if err != nil {
			// invalid index number means this is another field's key
			continue
		}

		// expand and get value
		if i >= av.Len() {
			av := reflect.MakeSlice(av.Type(), i+1, 2*(i+1))
			reflect.Copy(av, c.sv)
			c.sv.Set(av)
		}
		iv := av.Index(i)

		// decode in scope
		vd := &Decoder{Query: url.Values{c.scope: values}}
		setOK, err := vd.decode(decodeContext{sv: iv, scope: c.scope}, opts)
		if err != nil {
			return false, errors.Wrapf(err, "decodeNumberedSlice(%#v)", c.scope)
		}
		if setOK {
			setCount++
		}
	}

	setOK := setCount > 0
	return setOK, nil
}

func decodeSliceParts(typ reflect.Type, parts []string, opts tagOptions) (reflect.Value, error) {
	vd := &Decoder{Query: url.Values{}}
	for i, raw := range parts {
		vd.Query.Set(strconv.Itoa(i), raw)
	}

	av := reflect.MakeSlice(typ, len(parts), len(parts))
	for i := range parts {
		iv := av.Index(i)

		// decode it
		_, err := vd.decode(decodeContext{sv: iv, scope: strconv.Itoa(i)}, opts)
		if err != nil {
			return reflect.Value{}, err
		}
	}

	return av, nil
}

func (d *Decoder) decodeMap(c decodeContext) (bool, error) {
	keyTyp := c.sv.Type().Key()

	var mapKeyGen func(s string) (reflect.Value, error)
	switch keyTyp.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		mapKeyGen = intMapKey
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		mapKeyGen = uintMapKey
	case reflect.String:
		mapKeyGen = stringMapKey
	}
	if mapKeyGen == nil {
		return false, &UnsupportedTypeError{Type: keyTyp, ContextType: "mapKey"}
	}

	var setCount int
	prefix := c.scope + "["
	for key, values := range d.Query {
		if !(strings.HasPrefix(key, prefix) && key[len(key)-1] == ']') {
			continue
		}

		// get index key
		s := strings.TrimPrefix(key, prefix)
		s = s[:len(s)-1]

		k, err := mapKeyGen(s)
		if err != nil {
			return false, errors.Wrapf(err, "decodeMap(%#v)", c.scope)
		}

		iv := c.sv.MapIndex(k)

		// decode in scope
		vd := &Decoder{Query: url.Values{c.scope: values}}
		setOK, err := vd.decode(decodeContext{sv: iv, scope: c.scope}, nil)
		if err != nil {
			return false, errors.Wrapf(err, "decodeMap(%#v)", c.scope)
		}
		if setOK {
			setCount++
		}
	}

	setOK := setCount > 0
	return setOK, nil
}

func stringMapKey(s string) (reflect.Value, error) {
	return reflect.ValueOf(s), nil
}

func intMapKey(s string) (reflect.Value, error) {
	i, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(i), nil
}

func uintMapKey(s string) (reflect.Value, error) {
	i, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(i), nil
}

func (d *Decoder) decodeStruct(c decodeContext, opts tagOptions) (bool, error) {
	typ := c.sv.Type()

	// for time.Time struct
	if typ == timeType {
		if opts != nil && opts.Contains("unix") {
			s := d.Query.Get(c.scope)
			i, err := strconv.ParseInt(s, 10, 0)
			if err != nil {
				return false, errors.Wrapf(err, "decodeStruct(%#v):time.Time(unix)", c.scope)
			}

			t := time.Unix(i, 0)
			c.sv.Set(reflect.ValueOf(t))
			return true, nil
		}

		s := d.Query.Get(c.scope)
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return false, errors.Wrapf(err, "decodeStruct(%#v):time.Time", c.scope)
		}

		c.sv.Set(reflect.ValueOf(t))
		return true, nil
	}

	// for normal struct
	var setCount int
	for i := 0; i < c.sv.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous { // unexported
			continue
		}

		sv := c.sv.Field(i)

		// detect name
		var name string
		var opts tagOptions
		if tag, ok := sf.Tag.Lookup("url"); ok {
			name, opts = parseTag(tag)
		} else {
			name = sf.Name
		}

		// normalize name
		if name == "-" {
			continue
		} else if name == "" {
			name = sf.Name
		} else if sf.Anonymous && sv.Kind() == reflect.Struct {
			name = ""
		}

		// detect scope
		scope := c.scope
		if scope != "" && name != "" {
			scope += "[" + name + "]"
		} else if scope == "" && name != "" {
			scope = name
		}

		// decode it
		setOK, err := d.decode(decodeContext{sv: sv, scope: scope}, opts)
		if err != nil {
			return false, errors.Wrapf(err, "decodeStruct(%#v)", c.scope)
		}
		if setOK {
			setCount++
		}
	}

	setOK := setCount > 0
	return setOK, nil
}

func (d *Decoder) decodeString(c decodeContext) bool {
	if _, ok := d.Query[c.scope]; !ok {
		return false
	}

	s := d.Query.Get(c.scope)
	c.sv.SetString(s)
	return true
}

// An InvalidDecodeError describes an invalid argument passed to Decode.
// (The argument to Decode must be a non-nil pointer.)
type InvalidDecodeError struct {
	Type reflect.Type
}

func (e *InvalidDecodeError) Error() string {
	if e.Type == nil {
		return "query: Decode(data, nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "query: Decode(data, non-pointer " + e.Type.String() + ")"
	}
	return "query: Decode(data, nil " + e.Type.String() + ")"
}

// An UnsupportedTypeError describes an unsupported type value is passed to Decode.
type UnsupportedTypeError struct {
	Type        reflect.Type
	ContextType string
}

func (e *UnsupportedTypeError) Error() string {
	msg := e.Type.String() + " is unsupported"

	switch e.ContextType {
	case "mapKey":
		msg += " in map key"
	default:
	}

	return msg
}

// An InvalidBooleanValueError describes an invalid boolean value passed to Decode.
type InvalidBooleanValueError struct {
	Query      url.Values
	Key        string
	QueryValue string
	TrueValue  string
	FalseValue string
}

func (e *InvalidBooleanValueError) Error() string {
	return fmt.Sprintf("Invalid boolean value: %s (key: %#v, expected true: %#v, false: %#v)", e.Key, e.QueryValue, e.TrueValue, e.FalseValue)
}
