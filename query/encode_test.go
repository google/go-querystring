package query

import (
	"net/url"
	"reflect"

	"testing"
)

func TestValues(t *testing.T) {
	s := struct {
		A string `url:"a"`
		B int
		C string
		D string `url:"-"`
		E string `url:",omitempty"`
		F string `url:",omitempty"`
	}{
		A: "abc",
		B: 1,
		F: "foo",
	}
	v, err := Values(s)
	if err != nil {
		t.Errorf("Values(%q) returned error: %v", s, err)
	}

	want := url.Values{
		"a": {"abc"},
		"B": {"1"},
		"C": {""},
		"F": {"foo"},
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
