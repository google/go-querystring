package query

import (
	"net/url"
	"reflect"

	"testing"
)

func TestValues(t *testing.T) {
	s := struct {
		A string
		B int
	}{"abc", 1}
	v, err := Values(s)
	if err != nil {
		t.Errorf("Values(%q) returned error: %v", s, err)
	}

	want := url.Values{
		"A": {"abc"},
		"B": {"1"},
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
