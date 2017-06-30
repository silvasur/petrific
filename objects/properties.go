package objects

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
)

// properties are mappings from strings to strings that are encoded as a restricted version of URL query strings
// (only the characters [a-zA-Z0-9.:_-] are allowed, values are ordered by their key)
type properties map[string]string

// escapePropertyString escapes all bytes not in [a-zA-Z0-9.,:_-] as %XX, where XX represents the hexadecimal value of the byte.
// Compatible with URL query strings
func escapePropertyString(s string) []byte {
	out := []byte{}
	esc := []byte("%XX")

	for _, b := range []byte(s) {
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '.' || b == ',' || b == ':' || b == '_' || b == '-' {
			out = append(out, b)
		} else {
			hex.Encode(esc[1:], []byte{b})
			out = append(out, esc...)
		}
	}

	return out
}

func (p properties) MarshalText() ([]byte, error) { // Guaranteed to not fail, error is only here to satisfy encoding.TextMarshaler
	keys := make([]string, len(p))
	i := 0
	for k := range p {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	out := []byte{}

	first := true
	for _, k := range keys {
		if first {
			first = false
		} else {
			out = append(out, '&')
		}

		out = append(out, escapePropertyString(k)...)
		out = append(out, '=')
		out = append(out, escapePropertyString(p[k])...)
	}

	return out, nil
}

func (p properties) UnmarshalText(text []byte) error {
	vals, err := url.ParseQuery(string(text))
	if err != nil {
		return err
	}

	for k, v := range vals {
		if len(v) != 1 {
			return fmt.Errorf("Got %d values for key %s, expected 1", len(v), k)
		}

		p[k] = v[0]
	}

	return nil
}

func (a properties) Equals(b properties) bool {
	for k, va := range a {
		vb, ok := b[k]
		if !ok || vb != va {
			return false
		}
	}

	for k := range b {
		_, ok := a[k]
		if !ok {
			return false
		}
	}

	return true
}
