package query

import (
	"net/url"
	"reflect"
	"time"

	"testing"
)

func TestValues_types(t *testing.T) {
	str := "string"

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
			struct{ A *string }{A: &str},
			url.Values{"A": {str}},
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
			}{
				A: []string{"a", "b"},
				B: []string{"a", "b"},
				C: []string{"a", "b"},
				D: [2]string{"a", "b"},
				E: [2]string{"a", "b"},
				F: [2]string{"a", "b"},
			},
			url.Values{
				"A": {"a", "b"},
				"B": {"a,b"},
				"C": {"a b"},
				"D": {"a", "b"},
				"E": {"a,b"},
				"F": {"a b"},
			},
		},
		{
			// other types
			struct {
				A time.Time
				B time.Time `url:",unix"`
			}{
				A: time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC),
				B: time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC),
			},
			url.Values{
				"A": {"2000-01-01T12:34:56Z"},
				"B": {"946730096"},
			},
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
		t.Errorf("Values(%q) returned error: %v", s, err)
	}

	want := url.Values{
		"A":         {""},
		"omitempty": {""},
		"E":         {""}, // E is included because the pointer is not empty, even though the string being pointed to is
	}
	if !reflect.DeepEqual(want, v) {
		t.Errorf("Values(%q) returned %v, want %v", s, v, want)
	}
}

func TestValues_invalidInput(t *testing.T) {
	_, err := Values("")
	if err == nil {
		t.Errorf("expected Values() to return an error on invalid input")
	}
}
