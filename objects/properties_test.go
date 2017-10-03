package objects

import (
	"testing"
)

func TestPropEscape(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"empty", "", ""},
		{"no escape", "foo:bar_BAZ-123", "foo:bar_BAZ-123"},
		{"reserved chars", "foo=bar%baz%%=", "foo%3dbar%25baz%25%25%3d"},
	}

	for _, subtest := range tests {
		have := string(escapePropertyString(subtest.in))
		if have != subtest.want {
			t.Errorf("%s: want: '%s', have: '%s'", subtest.name, subtest.want, have)
		}
	}
}

func TestPropertyMarshalling(t *testing.T) {
	tests := []struct {
		name string
		in   Properties
		want string
	}{
		{"empty", Properties{}, ""},
		{"single", Properties{"foo": "bar"}, "foo=bar"},
		{"simple", Properties{"foo": "bar", "bar": "baz"}, "bar=baz&foo=bar"},
		{"escapes", Properties{"foo&bar": "%=baz", "?": "!"}, "%3f=%21&foo%26bar=%25%3dbaz"},
	}

	for _, subtest := range tests {
		have, err := subtest.in.MarshalText()
		if err != nil {
			t.Errorf("%s: Got an error: %s", err)
			continue
		}

		if string(have) != subtest.want {
			t.Errorf("%s: want: '%s', have: '%s'", subtest.name, subtest.want, have)
		}
	}
}

func TestPropertyUnmarshalling(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Properties
	}{
		{"empty", "", Properties{}},
		{"single", "foo=bar", Properties{"foo": "bar"}},
		{"simple", "bar=baz&foo=bar", Properties{"foo": "bar", "bar": "baz"}},
		{"escapes", "%3f=%21&foo%26bar=%25%3dbaz", Properties{"foo&bar": "%=baz", "?": "!"}},
	}

	for _, subtest := range tests {
		have := make(Properties)
		err := have.UnmarshalText([]byte(subtest.in))

		if err != nil {
			t.Errorf("%s: Got an error: %s", err)
			continue
		}

		if !have.Equals(subtest.want) {
			t.Errorf("%s: want: '%v', have: '%v'", subtest.name, subtest.want, have)
		}
	}
}
