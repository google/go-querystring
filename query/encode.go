package query

import (
	"fmt"
	"reflect"

	"net/url"
)

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

		tag := sf.Tag.Get("url")
		name, _ := parseTag(tag)
		if name == "" {
			name = sf.Name
		}

		sv := val.Field(i)
		values.Add(name, fmt.Sprint(sv.Interface()))
	}

	return values, nil
}
