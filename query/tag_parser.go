package query

import "strings"

type urlTag struct {
	name    string
	options tagOptions
}

type tagOptions []string

// Contains checks whether the tagOptions contains the specified option.
func (o tagOptions) Contains(option string) bool {
	for _, s := range o {
		if s == option {
			return true
		}
	}
	return false
}

const tagStringComma = "comma"
const tagStringSpace = "space"
const tagStringSemicolon = "semicolon"
const tagStringBrackets = "brackets"
const tagStringNumbered = "numbered"
const tagStringIndexed = "indexed"
const tagStringInt = "int"
const tagStringUnix = "unix"
const tagStringUnixMilli = "unixmilli"
const tagStringUnixNano = "unixnano"
const tagStringOmitEmpty = "omitempty"

// parseTag splits a struct field's url tag into its name and comma-separated
// options.
func parseTag(tag string) urlTag {
	s := strings.Split(tag, ",")
	return urlTag{s[0], s[1:]}
}
