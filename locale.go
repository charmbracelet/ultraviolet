package uv

import (
	"fmt"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

// LocaleFromEnv detects the locale of the current environment. It returns
// the locale as a string, e.g., "en_US.UTF-8" or empty string if it cannot be
// determined.
//
// It checks the environment variables LC_ALL, LC_CTYPE, and LANG in that order.
func LocaleFromEnv(env []string) string {
	for _, e := range []string{
		"LC_ALL",
		"LC_CTYPE",
		"LANG",
	} {
		if l, ok := Environ(env).LookupEnv(e); ok {
			return l
		}
	}
	return ""
}

// DetectLocaleEncoding tries to detect the encoding of the locale based on the
// locale string.
func DetectLocaleEncoding(locale string) (encoding.Encoding, error) {
	charset := locale
	if locale == "POSIX" || locale == "C" {
		charset = "US-ASCII"
	}
	if i := strings.IndexRune(charset, '@'); i >= 0 {
		charset = charset[:i] // Remove any charset modifiers
	}
	if i := strings.IndexRune(charset, '.'); i >= 0 {
		charset = charset[i+1:] // Remove the locale prefix
	}
	enc, err := ianaindex.IANA.Encoding(charset)
	if err != nil {
		return nil, fmt.Errorf("unknown encoding %q: %w", charset, err)
	}
	return enc, nil
}
