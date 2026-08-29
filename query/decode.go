package query

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

func Decode(values url.Values, out interface{}) error {
	val := reflect.ValueOf(out)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("query: Decode expects a non-nil pointer to a struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("query: Decode expects a pointer to a struct. Got %v", val.Kind())
	}

	return decodeStruct(values, val, "")
}

func decodeStruct(values url.Values, val reflect.Value, scope string) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		sf := typ.Field(i)

		if sf.PkgPath != "" {
			continue
		}

		tag := sf.Tag.Get("url")
		if tag == "-" {
			continue
		}

		name, _ := parseTag(tag)
		if name == "" {
			name = sf.Name
		}

		fullName := name
		if scope != "" {
			fullName = fmt.Sprintf("%s[%s]", scope, name)
		}

		if field.Kind() == reflect.Struct && sf.Type.Kind() != reflect.Struct {
			continue
		}
		if field.Kind() == reflect.Struct && sf.Type != reflect.TypeOf(timeType) {
			if err := decodeStruct(values, field, fullName); err != nil {
				return err
			}
			continue
		}

		vals, ok := values[fullName]
		if !ok || len(vals) == 0 {
			continue
		}
		raw := vals[0]

		if err := setFieldValue(field, raw); err != nil {
			return fmt.Errorf("query: cannot set field %q: %w", sf.Name, err)
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, raw string) error {
	if !field.CanSet() {
		return fmt.Errorf("field not settable")
	}

	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		field.SetBool(b)
	default:
		return fmt.Errorf("unsupported kind %s", field.Kind())
	}

	return nil
}
