package query

import (
	"net/url"
	"reflect"

	"testing"
)

func TestValues_types(t *testing.T) {
	s := struct {
		A string
		B int
		C uint
		D float32
		E bool
	}{}
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
	}
	if !reflect.DeepEqual(want, v) {
		t.Errorf("Values(%q) returned %v, want %v", s, v, want)
	}
}

func TestValues_tags(t *testing.T) {
	s := struct {
		A string
		B string `url:"so,omitempty"`
		C string `url:"-"`
		D string `url:"omitempty"` // actually named omitempty, not an option
	}{}
	v, err := Values(s)
	if err != nil {
		t.Errorf("Values(%q) returned error: %v", s, err)
	}

	want := url.Values{
		"A":         {""},
		"omitempty": {""},
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
