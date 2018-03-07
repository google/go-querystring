// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package query

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/k0kubun/pp"
)

func TestDecode_types(t *testing.T) {
	str := "string"
	strPtr := &str
	timeVal := time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC)

	tests := []struct {
		in   url.Values
		want interface{}
	}{
		{
			// empty
			url.Values{
				"A": {""},
				"B": {"0"},
				"C": {"0"},
				"D": {"0"},
				"E": {"false"},
			},
			struct {
				A string
				B int
				C uint
				D float32
				E bool
			}{},
		},
		{
			// basic primitives
			url.Values{
				"A": {"a"},
				"B": {"-1"},
				"C": {"2"},
				"D": {"0.25"},
				"E": {"true"},
			},
			struct {
				A string
				B int
				C uint
				D float32
				E bool
			}{
				A: "a",
				B: -1,
				C: 2,
				D: 0.25,
				E: true,
			},
		},
		{
			// pointers
			url.Values{
				"A": {str},
				"C": {str},
				"D": {"2000-01-01T12:34:56Z"},
			},
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
		},
		{
			// slices and arrays
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
		},
		{
			// other types
			url.Values{
				"A": {"2000-01-01T12:34:56Z"},
				"B": {"946730096"},
				"C": {"1"},
				"D": {"0"},
			},
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
		},
		{
			url.Values{
				"nest[a][value]": {"that"},
				"nest[b]":        {""},
			},
			struct {
				Nest Nested `url:"nest"`
			}{
				Nested{
					A: SubNested{
						Value: "that",
					},
				},
			},
		},
		{
			url.Values{
				"nest[a][value]":   {""},
				"nest[b]":          {""},
				"nest[ptr][value]": {"that"},
			},
			struct {
				Nest Nested `url:"nest"`
			}{
				Nested{
					Ptr: &SubNested{
						Value: "that",
					},
				},
			},
		},
		{
			// skip empty
			url.Values{},
			struct {
				A string
				B int
				C uint
				D float32
				E bool
			}{},
		},
	}

	for i, tt := range tests {
		v := reflect.New(reflect.ValueOf(tt.want).Type()).Elem().Interface()
		err := NewDecoder(tt.in).Decode(&v)
		if err != nil {
			t.Errorf("%d. Decode(%q) returned error: %v", i, tt.in, err)
			continue
		}

		if df := cmp.Diff(tt.want, v); df != "" {
			t.Errorf("%d. Decode(%q) diff = %s", i, tt.in, df)
		}
	}
}

func TestDecode_embeddedStructs(t *testing.T) {
	tests := []struct {
		in   url.Values
		want interface{}
	}{
		{
			url.Values{"C": {"foo"}},
			A{B{C: "foo"}},
		},
		{
			url.Values{"C": {"foo"}},
			D{B: B{C: "foo"}, C: "foo"},
		},
		{
			url.Values{"C": {"foo", "bar"}},
			F{e{B: B{C: "bar"}, C: "foo"}}, // With unexported embed
		},
	}

	for i, tt := range tests {
		v := reflect.New(reflect.ValueOf(tt.want).Type()).Elem().Interface()
		err := NewDecoder(tt.in).Decode(&v)
		if err != nil {
			t.Errorf("%d. Decode(%q) returned error: %v", i, tt.in, err)
			continue
		}

		if df := cmp.Diff(tt.want, v, cmpopts.IgnoreUnexported(F{})); df != "" {
			t.Errorf("%d. Decode(%q) diff = %s", i, tt.in, df)
			pp.Println(tt.want)
			pp.Println(v)
		}
	}
}
