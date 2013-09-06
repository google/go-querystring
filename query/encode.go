package query

import (
	"bytes"
	"fmt"
	"net/url"
	"reflect"

	"time"
)

var timeType = reflect.TypeOf(time.Time{})

// Values returns the url.Values encoding of v.
func Values(v interface{}) (url.Values, error) {
	values := url.Values{}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("query: Values() expects struct input")
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" { // unexported
			continue
		}

		tag := sf.Tag.Get("url")
		if tag == "-" {
			continue
		}
		name, opts := parseTag(tag)
		if name == "" {
			name = sf.Name
		}

		sv := val.Field(i)
		if opts.Contains("omitempty") && isEmptyValue(sv) {
			continue
		}

		if !isEmptyValue(sv) && sv.Kind() == reflect.Ptr {
			sv = sv.Elem()
		}

		switch sv.Kind() {
		case reflect.Slice, reflect.Array:
			var del string
			if opts.Contains("comma") {
				del = ","
			} else if opts.Contains("space") {
				del = " "
			}

			if del != "" {
				s := new(bytes.Buffer)
				first := true
				for i := 0; i < sv.Len(); i++ {
					if first {
						first = false
					} else {
						fmt.Fprint(s, del)
					}
					fmt.Fprint(s, sv.Index(i))
				}
				values.Add(name, s.String())
			} else {
				for i := 0; i < sv.Len(); i++ {
					values.Add(name, fmt.Sprint(sv.Index(i)))
				}
			}
		case reflect.Bool:
			var value string
			if opts.Contains("int") {
				if sv.Bool() {
					value = "1"
				} else {
					value = "0"
				}
			} else {
				value = fmt.Sprint(sv.Interface())
			}

			values.Add(name, value)
		default:
			switch sv.Type() {
			case timeType:
				t := sv.Interface().(time.Time)
				if opts.Contains("unix") {
					values.Add(name, fmt.Sprint(t.Unix()))
				} else {
					values.Add(name, t.Format(time.RFC3339))
				}
			default:
				values.Add(name, fmt.Sprint(sv.Interface()))
			}
		}
	}

	return values, nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	switch v.Type() {
	case timeType:
		return v.Interface().(time.Time).IsZero()
	}

	return false
}
