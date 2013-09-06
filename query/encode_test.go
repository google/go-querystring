package query

import (
	"net/url"
	"reflect"
	"time"

	"testing"
)

func TestValues_types(t *testing.T) {
	str := "string"
	s := struct {
		A string
		B int
		C uint
		D float32
		E bool
		F *string
		G []string
		H [1]string
		I []string `url:",comma"`
		J []string `url:",space"`
		K time.Time
		L time.Time `url:",unix"`
	}{
		F: &str,
		G: []string{"a", "b"},
		H: [1]string{"a"},
		I: []string{"a", "b"},
		J: []string{"a", "b"},
		K: time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC),
		L: time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC),
	}

	v, err := Values(s)
	if err != nil {
		t.Errorf("Values(%q) returned error: %v", s, err)
	}

	want := url.Values{
		"A": {""},
		"B": {"0"},
		"C": {"0"},
		"D": {"0"},
		"E": {"false"},
		"F": {"string"},
		"G": {"a", "b"},
		"H": {"a"},
		"I": {"a,b"},
		"J": {"a b"},
		"K": {"2000-01-01T12:34:56Z"},
		"L": {"946730096"},
	}
	if !reflect.DeepEqual(want, v) {
		t.Errorf("Values(%q) returned %v, want %v", s, v, want)
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
