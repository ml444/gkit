package main

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
)

func pluckFields(v interface{}) map[string]string {
	m := map[string]string{}
	_v := reflect.Indirect(reflect.ValueOf(v))
	switch _v.Kind() {
	case reflect.Map:
		if vv, ok := v.(map[string][]string); ok {
			for k, v := range vv {
				m[k] = strings.Join(v, ",")
			}
		} else if vv, ok := v.(map[string]string); ok {
			for k, v := range vv {
				m[k] = v
			}
		}
	case reflect.Struct:
		for i := 0; i < _v.NumField(); i++ {
			field := reflect.Indirect(_v.Field(i))
			if !field.IsValid() || !field.CanInterface() {
				continue
			}
			key := dashCase(_v.Type().Field(i).Name)
			if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.String {
				if field.Len() == 0 {
					continue
				}
				m[key] = strings.Join(field.Interface().([]string), ",")
			} else if field.Kind() == reflect.String {
				if field.String() == "" {
					continue
				}
				m[key] = field.String()
			} else {
				m[key] = fmt.Sprintf("%v", field.Interface())
			}

		}
	}
	return m
}

func dashCase(s string) string {
	b := strings.Builder{}
	for i := 0; i < len(s); i++ {
		v := s[i]
		if i != 0 && v >= 'A' && v <= 'Z' {
			b.WriteByte('-')
		}
		b.WriteByte(v)
	}
	return b.String()
}

func buildPathVars(path string) (res map[string]*string) {
	if strings.HasSuffix(path, "/") {
		fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: Path %s should not end with \"/\" \n", path)
	}
	pattern := regexp.MustCompile(`(?i){([a-z.0-9_\s]*)=?([^{}]*)}`)
	matches := pattern.FindAllStringSubmatch(path, -1)
	res = make(map[string]*string, len(matches))
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if len(name) > 1 && len(m[2]) > 0 {
			res[name] = &m[2]
		} else {
			res[name] = nil
		}
	}
	return
}

func replacePath(name string, value string, path string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`(?i){([\s]*%s\b[\s]*)=?([^{}]*)}`, name))
	idx := pattern.FindStringIndex(path)
	if len(idx) > 0 {
		path = fmt.Sprintf("%s{%s:%s}%s",
			path[:idx[0]], // The start of the match
			name,
			strings.ReplaceAll(value, "*", ".*"),
			path[idx[1]:],
		)
	}
	return path
}

func camelCaseVars(s string) string {
	subs := strings.Split(s, ".")
	vars := make([]string, 0, len(subs))
	for _, sub := range subs {
		vars = append(vars, camelCase(sub))
	}
	return strings.Join(vars, ".")
}

// camelCase returns the CamelCased name.
// If there is an interior underscore followed by a lower case letter,
// drop the underscore and convert the letter to upper case.
// There is a remote possibility of this rewrite causing a name collision,
// but it's so remote we're prepared to pretend it's nonexistent - since the
// C++ generator lowercase names, it's extremely unlikely to have two fields
// with different capitalization.
// In short, _my_field_name_2 becomes XMyFieldName_2.
func camelCase(s string) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// Need a capital letter; drop the '_'.
		t = append(t, 'X')
		i++
	}
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier.
		// The next word is a sequence of characters that must start upper case.
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}
