package query

import "testing"

func TestParseTag(t *testing.T) {
	parsedTag := parseTag("field,foobar,foo")
	if parsedTag.name != "field" {

		t.Fatalf("name = %q, want field", parsedTag.name)
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
		if parsedTag.options.Contains(tt.opt) != tt.want {
			t.Errorf("Contains(%q) = %v", tt.opt, !tt.want)
		}
	}
}
