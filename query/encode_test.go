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

	"github.com/google/go-cmp/cmp"
)

// test that Values(input) matches want.  If not, report an error on t.
func testValue(t *testing.T, input interface{}, want url.Values) {
	v, err := Values(input)
	if err != nil {
		t.Errorf("Values(%q) returned error: %v", input, err)
	}
	if diff := cmp.Diff(want, v); diff != "" {
		t.Errorf("Values(%q) mismatch:\n%s", input, diff)
	}
}

func TestValues_BasicTypes(t *testing.T) {
	tests := []struct {
		input interface{}
		want  url.Values
	}{
		// zero values
		{struct{ V string }{}, url.Values{"V": {""}}},
		{struct{ V int }{}, url.Values{"V": {"0"}}},
		{struct{ V uint }{}, url.Values{"V": {"0"}}},
		{struct{ V float32 }{}, url.Values{"V": {"0"}}},
		{struct{ V bool }{}, url.Values{"V": {"false"}}},

		// simple non-zero values
		{struct{ V string }{"v"}, url.Values{"V": {"v"}}},
		{struct{ V int }{1}, url.Values{"V": {"1"}}},
		{struct{ V uint }{1}, url.Values{"V": {"1"}}},
		{struct{ V float32 }{0.1}, url.Values{"V": {"0.1"}}},
		{struct{ V bool }{true}, url.Values{"V": {"true"}}},

		// bool-specific options
		{
			struct {
				V bool `url:",int"`
			}{false},
			url.Values{"V": {"0"}},
		},
		{
			struct {
				V bool `url:",int"`
			}{true},
			url.Values{"V": {"1"}},
		},

		// time values
		{
			struct {
				V time.Time
			}{time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC)},
			url.Values{"V": {"2000-01-01T12:34:56Z"}},
		},
		{
			struct {
				V time.Time `url:",unix"`
			}{time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC)},
			url.Values{"V": {"946730096"}},
		},
	}

	for _, tt := range tests {
		testValue(t, tt.input, tt.want)
	}
}

func TestValues_Pointers(t *testing.T) {
	str := "s"
	strPtr := &str

	tests := []struct {
		input interface{}
		want  url.Values
	}{
		// nil pointers (zero values)
		{struct{ V *string }{}, url.Values{"V": {""}}},
		{struct{ V *int }{}, url.Values{"V": {""}}},

		// non-zero pointer values
		{struct{ V *string }{&str}, url.Values{"V": {"s"}}},
		{struct{ V **string }{&strPtr}, url.Values{"V": {"s"}}},

		// slices of pointer values
		{struct{ V []*string }{}, url.Values{}},
		{struct{ V []*string }{[]*string{&str, &str}}, url.Values{"V": {"s", "s"}}},

		// pointer values for the input struct itself
		{(*struct{})(nil), url.Values{}},
		{&struct{}{}, url.Values{}},
		{&struct{ V string }{}, url.Values{"V": {""}}},
		{&struct{ V string }{"v"}, url.Values{"V": {"v"}}},
	}

	for _, tt := range tests {
		testValue(t, tt.input, tt.want)
	}
}

func TestValues_Slices(t *testing.T) {
	tests := []struct {
		input interface{}
		want  url.Values
	}{
		// slices of strings
		{
			struct{ V []string }{},
			url.Values{},
		},
		{
			struct{ V []string }{[]string{"a", "b"}},
			url.Values{"V": {"a", "b"}},
		},
		{
			struct {
				V []string `url:",comma"`
			}{[]string{"a", "b"}},
			url.Values{"V": {"a,b"}},
		},
		{
			struct {
				V []string `url:",space"`
			}{[]string{"a", "b"}},
			url.Values{"V": {"a b"}},
		},
		{
			struct {
				V []string `url:",semicolon"`
			}{[]string{"a", "b"}},
			url.Values{"V": {"a;b"}},
		},
		{
			struct {
				V []string `url:",brackets"`
			}{[]string{"a", "b"}},
			url.Values{"V[]": {"a", "b"}},
		},
		{
			struct {
				V []string `url:",numbered"`
			}{[]string{"a", "b"}},
			url.Values{"V0": {"a"}, "V1": {"b"}},
		},

		// arrays of strings
		{
			struct{ V [2]string }{},
			url.Values{"V": {"", ""}},
		},
		{
			struct{ V [2]string }{[2]string{"a", "b"}},
			url.Values{"V": {"a", "b"}},
		},
		{
			struct {
				V [2]string `url:",comma"`
			}{[2]string{"a", "b"}},
			url.Values{"V": {"a,b"}},
		},
		{
			struct {
				V [2]string `url:",space"`
			}{[2]string{"a", "b"}},
			url.Values{"V": {"a b"}},
		},
		{
			struct {
				V [2]string `url:",semicolon"`
			}{[2]string{"a", "b"}},
			url.Values{"V": {"a;b"}},
		},
		{
			struct {
				V [2]string `url:",brackets"`
			}{[2]string{"a", "b"}},
			url.Values{"V[]": {"a", "b"}},
		},
		{
			struct {
				V [2]string `url:",numbered"`
			}{[2]string{"a", "b"}},
			url.Values{"V0": {"a"}, "V1": {"b"}},
		},

		// slice of bools with additional options
		{
			struct {
				V []bool `url:",space,int"`
			}{[]bool{true, false}},
			url.Values{"V": {"1 0"}},
		},
	}

	for _, tt := range tests {
		testValue(t, tt.input, tt.want)
	}
}

func TestValues_NestedTypes(t *testing.T) {
	type SubNested struct {
		Value string `url:"value"`
	}

	type Nested struct {
		A   SubNested  `url:"a"`
		B   *SubNested `url:"b"`
		Ptr *SubNested `url:"ptr,omitempty"`
	}

	tests := []struct {
		input interface{}
		want  url.Values
	}{
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

	for _, tt := range tests {
		testValue(t, tt.input, tt.want)
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

		// unknown type - always false unless a nil pointer, which are always empty.
		{(*struct{ int })(nil), true},
		{struct{ int }{}, false},
		{struct{ int }{0}, false},
		{struct{ int }{1}, false},
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
