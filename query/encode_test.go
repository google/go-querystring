// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package query

import (
	"errors"
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
		t.Errorf("Values(%#v) mismatch:\n%s", input, diff)
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
		{
			struct {
				V time.Time `url:",unixmilli"`
			}{time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC)},
			url.Values{"V": {"946730096000"}},
		},
		{
			struct {
				V time.Time `url:",unixnano"`
			}{time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC)},
			url.Values{"V": {"946730096000000000"}},
		},
		{
			struct {
				V time.Time `layout:"2006-01-02"`
			}{time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC)},
			url.Values{"V": {"2000-01-01"}},
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

		// pointer to slice
		{struct{ V *[]string }{}, url.Values{"V": {""}}},
		{struct{ V *[]string }{&[]string{"a", "b"}}, url.Values{"V": {"a", "b"}}},

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
			struct{ V []string }{[]string{}},
			url.Values{},
		},
		{
			struct{ V []string }{[]string{""}},
			url.Values{"V": {""}},
		},
		{
			struct{ V []string }{[]string{"a", "b"}},
			url.Values{"V": {"a", "b"}},
		},
		{
			struct {
				V []string `url:",comma"`
			}{[]string{}},
			url.Values{},
		},
		{
			struct {
				V []string `url:",comma"`
			}{[]string{""}},
			url.Values{"V": {""}},
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

		// custom delimiters
		{
			struct {
				V []string `del:","`
			}{[]string{"a", "b"}},
			url.Values{"V": {"a,b"}},
		},
		{
			struct {
				V []string `del:"|"`
			}{[]string{"a", "b"}},
			url.Values{"V": {"a|b"}},
		},
		{
			struct {
				V []string `del:"🥑"`
			}{[]string{"a", "b"}},
			url.Values{"V": {"a🥑b"}},
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
						Value: "v",
					},
				},
			},
			url.Values{
				"nest[a][value]": {"v"},
				"nest[b]":        {""},
			},
		},
		{
			struct {
				Nest Nested `url:"nest"`
			}{
				Nested{
					Ptr: &SubNested{
						Value: "v",
					},
				},
			},
			url.Values{
				"nest[a][value]":   {""},
				"nest[b]":          {""},
				"nest[ptr][value]": {"v"},
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

func TestValues_OmitEmpty(t *testing.T) {
	str := ""

	tests := []struct {
		input interface{}
		want  url.Values
	}{
		{struct{ v string }{}, url.Values{}}, // non-exported field
		{
			struct {
				V string `url:",omitempty"`
			}{},
			url.Values{},
		},
		{
			struct {
				V string `url:"-"`
			}{},
			url.Values{},
		},
		{
			struct {
				V string `url:"omitempty"` // actually named omitempty
			}{},
			url.Values{"omitempty": {""}},
		},
		{
			// include value for a non-nil pointer to an empty value
			struct {
				V *string `url:",omitempty"`
			}{&str},
			url.Values{"V": {""}},
		},
	}

	for _, tt := range tests {
		testValue(t, tt.input, tt.want)
	}
}

func TestValues_EmbeddedStructs(t *testing.T) {
	type Inner struct {
		V string
	}
	type Outer struct {
		Inner
	}
	type OuterPtr struct {
		*Inner
	}
	type Mixed struct {
		Inner
		V string
	}
	type unexported struct {
		Inner
		V string
	}
	type Exported struct {
		unexported
	}

	tests := []struct {
		input interface{}
		want  url.Values
	}{
		{
			Outer{Inner{V: "a"}},
			url.Values{"V": {"a"}},
		},
		{
			OuterPtr{&Inner{V: "a"}},
			url.Values{"V": {"a"}},
		},
		{
			Mixed{Inner: Inner{V: "a"}, V: "b"},
			url.Values{"V": {"b", "a"}},
		},
		{
			// values from unexported embed are still included
			Exported{
				unexported{
					Inner: Inner{V: "bar"},
					V:     "foo",
				},
			},
			url.Values{"V": {"foo", "bar"}},
		},
	}

	for _, tt := range tests {
		testValue(t, tt.input, tt.want)
	}
}

func TestValues_InvalidInput(t *testing.T) {
	_, err := Values("")
	if err == nil {
		t.Errorf("expected Values() to return an error on invalid input")
	}
}

// customEncodedStrings is a slice of strings with a custom URL encoding
type customEncodedStrings []string

// EncodeValues using key name of the form "{key}.N" where N increments with
// each value.  A value of "err" will return an error.
func (m customEncodedStrings) EncodeValues(key string, v *url.Values) error {
	for i, arg := range m {
		if arg == "err" {
			return errors.New("encoding error")
		}
		v.Set(fmt.Sprintf("%s.%d", key, i), arg)
	}
	return nil
}

func TestValues_CustomEncodingSlice(t *testing.T) {
	tests := []struct {
		input interface{}
		want  url.Values
	}{
		{
			struct {
				V customEncodedStrings `url:"v"`
			}{},
			url.Values{},
		},
		{
			struct {
				V customEncodedStrings `url:"v"`
			}{[]string{"a", "b"}},
			url.Values{"v.0": {"a"}, "v.1": {"b"}},
		},

		// pointers to custom encoded types
		{
			struct {
				V *customEncodedStrings `url:"v"`
			}{},
			url.Values{},
		},
		{
			struct {
				V *customEncodedStrings `url:"v"`
			}{(*customEncodedStrings)(&[]string{"a", "b"})},
			url.Values{"v.0": {"a"}, "v.1": {"b"}},
		},

		// custom encoded type held in an interface field
		{
			struct {
				V interface{} `url:"v"`
			}{customEncodedStrings{"a", "b"}},
			url.Values{"v.0": {"a"}, "v.1": {"b"}},
		},
	}

	for _, tt := range tests {
		testValue(t, tt.input, tt.want)
	}
}

// One of the few ways reflectValues will return an error is if a custom
// encoder returns an error.  Test all of the various ways that can happen.
func TestValues_CustomEncoding_Error(t *testing.T) {
	type st struct {
		V customEncodedStrings
	}
	tests := []struct {
		input interface{}
	}{
		{
			st{[]string{"err"}},
		},
		{ // struct field
			struct{ S st }{st{[]string{"err"}}},
		},
		{ // embedded struct
			struct{ st }{st{[]string{"err"}}},
		},
	}
	for _, tt := range tests {
		_, err := Values(tt.input)
		if err == nil {
			t.Errorf("Values(%q) did not return expected encoding error", tt.input)
		}
	}
}

// customEncodedInt is an int with a custom URL encoding
type customEncodedInt int

// EncodeValues encodes values with leading underscores
func (m customEncodedInt) EncodeValues(key string, v *url.Values) error {
	v.Set(key, fmt.Sprintf("_%d", m))
	return nil
}

func TestValues_CustomEncodingInt(t *testing.T) {
	var zero customEncodedInt = 0
	var one customEncodedInt = 1
	tests := []struct {
		input interface{}
		want  url.Values
	}{
		{
			struct {
				V customEncodedInt `url:"v"`
			}{},
			url.Values{"v": {"_0"}},
		},
		{
			struct {
				V customEncodedInt `url:"v,omitempty"`
			}{zero},
			url.Values{},
		},
		{
			struct {
				V customEncodedInt `url:"v"`
			}{one},
			url.Values{"v": {"_1"}},
		},

		// pointers to custom encoded types
		{
			struct {
				V *customEncodedInt `url:"v"`
			}{},
			url.Values{"v": {"_0"}},
		},
		{
			struct {
				V *customEncodedInt `url:"v,omitempty"`
			}{},
			url.Values{},
		},
		{
			struct {
				V *customEncodedInt `url:"v,omitempty"`
			}{&zero},
			url.Values{"v": {"_0"}},
		},
		{
			struct {
				V *customEncodedInt `url:"v"`
			}{&one},
			url.Values{"v": {"_1"}},
		},
	}

	for _, tt := range tests {
		testValue(t, tt.input, tt.want)
	}
}

// customEncodedInt is an int with a custom URL encoding defined on its pointer
// value.
type customEncodedIntPtr int

// EncodeValues encodes a 0 as false, 1 as true, and nil as unknown.  All other
// values cause an error.
func (m *customEncodedIntPtr) EncodeValues(key string, v *url.Values) error {
	if m == nil {
		v.Set(key, "undefined")
	} else {
		v.Set(key, fmt.Sprintf("_%d", *m))
	}
	return nil
}

// Test behavior when encoding is defined for a pointer of a custom type.
// Custom type should be able to encode values for nil pointers.
func TestValues_CustomEncodingPointer(t *testing.T) {
	var zero customEncodedIntPtr = 0
	var one customEncodedIntPtr = 1
	tests := []struct {
		input interface{}
		want  url.Values
	}{
		// non-pointer values do not get the custom encoding because
		// they don't implement the encoder interface.
		{
			struct {
				V customEncodedIntPtr `url:"v"`
			}{},
			url.Values{"v": {"0"}},
		},
		{
			struct {
				V customEncodedIntPtr `url:"v,omitempty"`
			}{},
			url.Values{},
		},
		{
			struct {
				V customEncodedIntPtr `url:"v"`
			}{one},
			url.Values{"v": {"1"}},
		},

		// pointers to custom encoded types.
		{
			struct {
				V *customEncodedIntPtr `url:"v"`
			}{},
			url.Values{"v": {"undefined"}},
		},
		{
			struct {
				V *customEncodedIntPtr `url:"v,omitempty"`
			}{},
			url.Values{},
		},
		{
			struct {
				V *customEncodedIntPtr `url:"v"`
			}{&zero},
			url.Values{"v": {"_0"}},
		},
		{
			struct {
				V *customEncodedIntPtr `url:"v,omitempty"`
			}{&zero},
			url.Values{"v": {"_0"}},
		},
		{
			struct {
				V *customEncodedIntPtr `url:"v"`
			}{&one},
			url.Values{"v": {"_1"}},
		},
	}

	for _, tt := range tests {
		testValue(t, tt.input, tt.want)
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

func TestValues_EncodingCase(t *testing.T) {
	type TestStruct struct {
		FirstName  string
		UserID     int
		HTTPServer string
		URL        string
		User2Name  string
		FieldX     string
		CustomName string `url:"custom_param"`
		OmitEmpty  string `url:",omitempty"`
	}

	input := TestStruct{
		FirstName:  "Alice",
		UserID:     123,
		HTTPServer: "nginx",
		URL:        "example.com",
		User2Name:  "bob",
		FieldX:     "x",
		CustomName: "custom",
	}

	// CaseDefault (and without options)
	wantDefault := url.Values{
		"FirstName":    {"Alice"},
		"UserID":       {"123"},
		"HTTPServer":   {"nginx"},
		"URL":          {"example.com"},
		"User2Name":    {"bob"},
		"FieldX":       {"x"},
		"custom_param": {"custom"},
	}
	testValue(t, input, wantDefault)

	vDefault, err := Values(input, Options{EncodingCase: CaseDefault})
	if err != nil {
		t.Fatalf("Values with CaseDefault failed: %v", err)
	}
	if diff := cmp.Diff(wantDefault, vDefault); diff != "" {
		t.Errorf("Values with CaseDefault mismatch:\n%s", diff)
	}

	// CaseSnake and CASE_SNAKE
	wantSnake := url.Values{
		"first_name":   {"Alice"},
		"user_id":      {"123"},
		"http_server":  {"nginx"},
		"url":          {"example.com"},
		"user2_name":   {"bob"},
		"field_x":      {"x"},
		"custom_param": {"custom"},
	}
	vSnake, err := Values(input, Options{EncodingCase: CaseSnake})
	if err != nil {
		t.Fatalf("Values with CaseSnake failed: %v", err)
	}
	if diff := cmp.Diff(wantSnake, vSnake); diff != "" {
		t.Errorf("Values with CaseSnake mismatch:\n%s", diff)
	}
	vSnakeCompat, err := ValuesWithOptions(input, Options{EncodingCase: CASE_SNAKE})
	if err != nil {
		t.Fatalf("ValuesWithOptions with CASE_SNAKE failed: %v", err)
	}
	if diff := cmp.Diff(wantSnake, vSnakeCompat); diff != "" {
		t.Errorf("ValuesWithOptions with CASE_SNAKE mismatch:\n%s", diff)
	}

	// CaseCamel and CASE_CAMEL
	wantCamel := url.Values{
		"firstName":    {"Alice"},
		"userId":       {"123"},
		"httpServer":   {"nginx"},
		"url":          {"example.com"},
		"user2Name":    {"bob"},
		"fieldX":       {"x"},
		"custom_param": {"custom"},
	}
	vCamel, err := Values(input, Options{EncodingCase: CaseCamel})
	if err != nil {
		t.Fatalf("Values with CaseCamel failed: %v", err)
	}
	if diff := cmp.Diff(wantCamel, vCamel); diff != "" {
		t.Errorf("Values with CaseCamel mismatch:\n%s", diff)
	}
	vCamelCompat, err := Values(input, Options{EncodingCase: CASE_CAMEL})
	if err != nil {
		t.Fatalf("Values with CASE_CAMEL failed: %v", err)
	}
	if diff := cmp.Diff(wantCamel, vCamelCompat); diff != "" {
		t.Errorf("Values with CASE_CAMEL mismatch:\n%s", diff)
	}

	// CaseKebab and CASE_KEBAB
	wantKebab := url.Values{
		"first-name":   {"Alice"},
		"user-id":      {"123"},
		"http-server":  {"nginx"},
		"url":          {"example.com"},
		"user2-name":   {"bob"},
		"field-x":      {"x"},
		"custom_param": {"custom"},
	}
	vKebab, err := Values(input, Options{EncodingCase: CaseKebab})
	if err != nil {
		t.Fatalf("Values with CaseKebab failed: %v", err)
	}
	if diff := cmp.Diff(wantKebab, vKebab); diff != "" {
		t.Errorf("Values with CaseKebab mismatch:\n%s", diff)
	}
	vKebabCompat, err := Values(input, Options{EncodingCase: CASE_KEBAB})
	if err != nil {
		t.Fatalf("Values with CASE_KEBAB failed: %v", err)
	}
	if diff := cmp.Diff(wantKebab, vKebabCompat); diff != "" {
		t.Errorf("Values with CASE_KEBAB mismatch:\n%s", diff)
	}

	// CasePascal and CASE_PASCAL
	wantPascal := url.Values{
		"FirstName":    {"Alice"},
		"UserId":       {"123"},
		"HttpServer":   {"nginx"},
		"Url":          {"example.com"},
		"User2Name":    {"bob"},
		"FieldX":       {"x"},
		"custom_param": {"custom"},
	}
	vPascal, err := Values(input, Options{EncodingCase: CasePascal})
	if err != nil {
		t.Fatalf("Values with CasePascal failed: %v", err)
	}
	if diff := cmp.Diff(wantPascal, vPascal); diff != "" {
		t.Errorf("Values with CasePascal mismatch:\n%s", diff)
	}
	vPascalCompat, err := Values(input, Options{EncodingCase: CASE_PASCAL})
	if err != nil {
		t.Fatalf("Values with CASE_PASCAL failed: %v", err)
	}
	if diff := cmp.Diff(wantPascal, vPascalCompat); diff != "" {
		t.Errorf("Values with CASE_PASCAL mismatch:\n%s", diff)
	}

	// Unknown Case fallback
	vUnknown, err := Values(input, Options{EncodingCase: Case(99)})
	if err != nil {
		t.Fatalf("Values with unknown Case failed: %v", err)
	}
	if len(vUnknown) == 0 {
		t.Errorf("Values with unknown Case returned empty")
	}
}

func TestValues_TagCasing(t *testing.T) {
	type TaggedStruct struct {
		FieldOne   string `url:",snake"`
		FieldTwo   string `url:",camel"`
		FieldThree string `url:",kebab"`
		FieldFour  string `url:",pascal"`
		FieldFive  string `url:"explicit_name,camel"`
	}

	input := TaggedStruct{
		FieldOne:   "one",
		FieldTwo:   "two",
		FieldThree: "three",
		FieldFour:  "four",
		FieldFive:  "five",
	}

	want := url.Values{
		"field_one":     {"one"},
		"fieldTwo":      {"two"},
		"field-three":   {"three"},
		"FieldFour":     {"four"},
		"explicit_name": {"five"},
	}

	testValue(t, input, want)
}

func TestValues_NestedCasing(t *testing.T) {
	type Address struct {
		StreetName string
		PostalCode int
	}
	type User struct {
		HomeAddress Address
		UserProfile struct {
			DisplayName string
		}
	}

	input := User{
		HomeAddress: Address{
			StreetName: "Main St",
			PostalCode: 12345,
		},
		UserProfile: struct {
			DisplayName string
		}{
			DisplayName: "Alice",
		},
	}

	wantSnake := url.Values{
		"home_address[street_name]":  {"Main St"},
		"home_address[postal_code]":  {"12345"},
		"user_profile[display_name]": {"Alice"},
	}

	vSnake, err := Values(input, Options{EncodingCase: CaseSnake})
	if err != nil {
		t.Fatalf("Values nested with CaseSnake failed: %v", err)
	}
	if diff := cmp.Diff(wantSnake, vSnake); diff != "" {
		t.Errorf("Values nested with CaseSnake mismatch:\n%s", diff)
	}

	wantCamel := url.Values{
		"homeAddress[streetName]":  {"Main St"},
		"homeAddress[postalCode]":  {"12345"},
		"userProfile[displayName]": {"Alice"},
	}

	vCamel, err := Values(input, Options{EncodingCase: CaseCamel})
	if err != nil {
		t.Fatalf("Values nested with CaseCamel failed: %v", err)
	}
	if diff := cmp.Diff(wantCamel, vCamel); diff != "" {
		t.Errorf("Values nested with CaseCamel mismatch:\n%s", diff)
	}
}

func TestValues_EmbeddedCasing(t *testing.T) {
	type Inner struct {
		StreetName string
	}
	type Outer struct {
		Inner
		UserName string
	}

	input := Outer{
		Inner:    Inner{StreetName: "Main St"},
		UserName: "Alice",
	}

	wantSnake := url.Values{
		"street_name": {"Main St"},
		"user_name":   {"Alice"},
	}

	vSnake, err := Values(input, Options{EncodingCase: CaseSnake})
	if err != nil {
		t.Fatalf("Values embedded with CaseSnake failed: %v", err)
	}
	if diff := cmp.Diff(wantSnake, vSnake); diff != "" {
		t.Errorf("Values embedded with CaseSnake mismatch:\n%s", diff)
	}
}

func TestApplyCase_EdgeCases(t *testing.T) {
	if got := applyCase("", CaseSnake); got != "" {
		t.Errorf("applyCase(\"\") = %q, want \"\"", got)
	}
	if got := applyCase("___", CaseSnake); got != "" {
		t.Errorf("applyCase(\"___\") = %q, want \"\"", got)
	}
	if got := applyCase("a_b-c d", CaseSnake); got != "a_b_c_d" {
		t.Errorf("applyCase(\"a_b-c d\") = %q, want \"a_b_c_d\"", got)
	}
	if got := applyCase("A_B", CaseCamel); got != "aB" {
		t.Errorf("applyCase(\"A_B\") = %q, want \"aB\"", got)
	}
	if got := applyCase("a_b", CasePascal); got != "AB" {
		t.Errorf("applyCase(\"a_b\") = %q, want \"AB\"", got)
	}
	if got := formatWords(nil, CaseSnake); got != "" {
		t.Errorf("formatWords(nil) = %q, want \"\"", got)
	}
}
