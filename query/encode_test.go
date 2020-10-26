// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package query

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"
)

type Nested struct {
	A   SubNested  `url:"a"`
	B   *SubNested `url:"b"`
	Ptr *SubNested `url:"ptr,omitempty"`
}

type SubNested struct {
	Value string `url:"value"`
}

func TestValues_types(t *testing.T) {
	str := "string"
	strPtr := &str
	timeVal := time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC)

	tests := []struct {
		in   interface{}
		want url.Values
	}{
		{
			// basic primitives
			struct {
				A string
				B int
				C uint
				D float32
				E bool
			}{},
			url.Values{
				"A": {""},
				"B": {"0"},
				"C": {"0"},
				"D": {"0"},
				"E": {"false"},
			},
		},
		{
			// pointers
			struct {
				A *string
				B *int
				C **string
				D *time.Time
			}{
				A: strPtr,
				C: &strPtr,
				D: &timeVal,
			},
			url.Values{
				"A": {str},
				"B": {""},
				"C": {str},
				"D": {"2000-01-01T12:34:56Z"},
			},
		},
		{
			// slices and arrays
			struct {
				A []string
				B []string `url:",comma"`
				C []string `url:",space"`
				D [2]string
				E [2]string `url:",comma"`
				F [2]string `url:",space"`
				G []*string `url:",space"`
				H []bool    `url:",int,space"`
				I []string  `url:",brackets"`
				J []string  `url:",semicolon"`
				K []string  `url:",numbered"`
			}{
				A: []string{"a", "b"},
				B: []string{"a", "b"},
				C: []string{"a", "b"},
				D: [2]string{"a", "b"},
				E: [2]string{"a", "b"},
				F: [2]string{"a", "b"},
				G: []*string{&str, &str},
				H: []bool{true, false},
				I: []string{"a", "b"},
				J: []string{"a", "b"},
				K: []string{"a", "b"},
			},
			url.Values{
				"A":   {"a", "b"},
				"B":   {"a,b"},
				"C":   {"a b"},
				"D":   {"a", "b"},
				"E":   {"a,b"},
				"F":   {"a b"},
				"G":   {"string string"},
				"H":   {"1 0"},
				"I[]": {"a", "b"},
				"J":   {"a;b"},
				"K0":  {"a"},
				"K1":  {"b"},
			},
		},
		{
			// other types
			struct {
				A time.Time
				B time.Time `url:",unix"`
				C bool      `url:",int"`
				D bool      `url:",int"`
			}{
				A: time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC),
				B: time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC),
				C: true,
				D: false,
			},
			url.Values{
				"A": {"2000-01-01T12:34:56Z"},
				"B": {"946730096"},
				"C": {"1"},
				"D": {"0"},
			},
		},
		{
			struct {
				Nest Nested `url:"nest"`
			}{
				Nested{
					A: SubNested{
						Value: "that",
					},
				},
			},
			url.Values{
				"nest[a][value]": {"that"},
				"nest[b]":        {""},
			},
		},
		{
			struct {
				Nest Nested `url:"nest"`
			}{
				Nested{
					Ptr: &SubNested{
						Value: "that",
					},
				},
			},
			url.Values{
				"nest[a][value]":   {""},
				"nest[b]":          {""},
				"nest[ptr][value]": {"that"},
			},
		},
		{
			nil,
			url.Values{},
		},
	}

	for i, tt := range tests {
		v, err := Values(tt.in)
		if err != nil {
			t.Errorf("%d. Values(%q) returned error: %v", i, tt.in, err)
		}

		if !reflect.DeepEqual(tt.want, v) {
			t.Errorf("%d. Values(%q) returned %v, want %v", i, tt.in, v, tt.want)
		}
	}
}

func TestValues_omitEmpty(t *testing.T) {
	str := ""
	s := struct {
		a string
		A string
		B string  `url:",omitempty"`
		C string  `url:"-"`
		D string  `url:"omitempty"` // actually named omitempty, not an option
		E *string `url:",omitempty"`
	}{E: &str}

	v, err := Values(s)
	if err != nil {
		t.Errorf("Values(%v) returned error: %v", s, err)
	}

	want := url.Values{
		"A":         {""},
		"omitempty": {""},
		"E":         {""}, // E is included because the pointer is not empty, even though the string being pointed to is
	}
	if !reflect.DeepEqual(want, v) {
		t.Errorf("Values(%v) returned %v, want %v", s, v, want)
	}
}

type A struct {
	B
}

type B struct {
	C string
}

type D struct {
	B
	C string
}

type e struct {
	B
	C string
}

type F struct {
	e
}

func TestValues_embeddedStructs(t *testing.T) {
	tests := []struct {
		in   interface{}
		want url.Values
	}{
		{
			A{B{C: "foo"}},
			url.Values{"C": {"foo"}},
		},
		{
			D{B: B{C: "bar"}, C: "foo"},
			url.Values{"C": {"foo", "bar"}},
		},
		{
			F{e{B: B{C: "bar"}, C: "foo"}}, // With unexported embed
			url.Values{"C": {"foo", "bar"}},
		},
	}

	for i, tt := range tests {
		v, err := Values(tt.in)
		if err != nil {
			t.Errorf("%d. Values(%q) returned error: %v", i, tt.in, err)
		}

		if !reflect.DeepEqual(tt.want, v) {
			t.Errorf("%d. Values(%q) returned %v, want %v", i, tt.in, v, tt.want)
		}
	}
}

func TestValues_invalidInput(t *testing.T) {
	_, err := Values("")
	if err == nil {
		t.Errorf("expected Values() to return an error on invalid input")
	}
}

type EncodedArgs []string

func (m EncodedArgs) EncodeValues(key string, v *url.Values) error {
	for i, arg := range m {
		v.Set(fmt.Sprintf("%s.%d", key, i), arg)
	}
	return nil
}

func TestValues_Marshaler(t *testing.T) {
	s := struct {
		Args EncodedArgs `url:"arg"`
	}{[]string{"a", "b", "c"}}
	v, err := Values(s)
	if err != nil {
		t.Errorf("Values(%q) returned error: %v", s, err)
	}

	want := url.Values{
		"arg.0": {"a"},
		"arg.1": {"b"},
		"arg.2": {"c"},
	}
	if !reflect.DeepEqual(want, v) {
		t.Errorf("Values(%q) returned %v, want %v", s, v, want)
	}
}

func TestValues_MarshalerWithNilPointer(t *testing.T) {
	s := struct {
		Args *EncodedArgs `url:"arg"`
	}{}
	v, err := Values(s)
	if err != nil {
		t.Errorf("Values(%v) returned error: %v", s, err)
	}

	want := url.Values{}
	if !reflect.DeepEqual(want, v) {
		t.Errorf("Values(%v) returned %v, want %v", s, v, want)
	}
}

func TestIsEmptyValue(t *testing.T) {
	str := "string"
	tests := []struct {
		value interface{}
		empty bool
	}{
		// slices, arrays, and maps
		{[]int{}, true},
		{[]int{0}, false},
		{[0]int{}, true},
		{[3]int{}, false},
		{[3]int{1}, false},
		{map[string]string{}, true},
		{map[string]string{"a": "b"}, false},

		// strings
		{"", true},
		{" ", false},
		{"a", false},

		// bool
		{true, false},
		{false, true},

		// ints of various types
		{(int)(0), true}, {(int)(1), false}, {(int)(-1), false},
		{(int8)(0), true}, {(int8)(1), false}, {(int8)(-1), false},
		{(int16)(0), true}, {(int16)(1), false}, {(int16)(-1), false},
		{(int32)(0), true}, {(int32)(1), false}, {(int32)(-1), false},
		{(int64)(0), true}, {(int64)(1), false}, {(int64)(-1), false},
		{(uint)(0), true}, {(uint)(1), false},
		{(uint8)(0), true}, {(uint8)(1), false},
		{(uint16)(0), true}, {(uint16)(1), false},
		{(uint32)(0), true}, {(uint32)(1), false},
		{(uint64)(0), true}, {(uint64)(1), false},

		// floats
		{(float32)(0), true}, {(float32)(0.0), true}, {(float32)(0.1), false},
		{(float64)(0), true}, {(float64)(0.0), true}, {(float64)(0.1), false},

		// pointers
		{(*int)(nil), true},
		{new([]int), false},
		{&str, false},

		// time
		{time.Time{}, true},
		{time.Now(), false},
	}

	for _, tt := range tests {
		got := isEmptyValue(reflect.ValueOf(tt.value))
		want := tt.empty
		if got != want {
			t.Errorf("isEmptyValue(%v) returned %t; want %t", tt.value, got, want)

		}
	}
}

func TestParseTag(t *testing.T) {
	name, opts := parseTag("field,foobar,foo")
	if name != "field" {
		t.Fatalf("name = %q, want field", name)
	}
	for _, tt := range []struct {
		opt  string
		want bool
	}{
		{"foobar", true},
		{"foo", true},
		{"bar", false},
		{"field", false},
	} {
		if opts.Contains(tt.opt) != tt.want {
			t.Errorf("Contains(%q) = %v", tt.opt, !tt.want)
		}
	}
}
