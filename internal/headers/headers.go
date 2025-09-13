package headers

import (
	"bytes"
	"errors"
	"slices"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Get(key string) (value string, ok bool) {
	key = strings.ToLower(key)
	value, ok = h[key]
	return value, ok
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		// get more data byte slice
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	headerText := string(data[:idx])
	headerName, headerValue, err := headerFromString(headerText)
	if err != nil {
		return 0, false, err
	}
	h.Set(headerName, headerValue)
	return idx + 2, false, nil
}

func (h Headers) Set(key string, value string) {
	key = strings.ToLower(key)
	prevValue, ok := h[key]
	if ok {
		h[key] = prevValue + ", " + value
	} else {
		h[key] = value
	}
}

var tokenChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

// validTokens checks if the data contains only valid tokens
// or characters that are allowed in a token
func validTokens(data []byte) bool {
	for _, c := range data {
		if !isTokenChar(c) {
			return false
		}
	}
	return true
}

func isTokenChar(c byte) bool {
	if c >= 'A' && c <= 'Z' ||
		c >= 'a' && c <= 'z' ||
		c >= '0' && c <= '9' {
		return true
	}

	return slices.Contains(tokenChars, c)
}

func headerFromString(s string) (string, string, error) {
	index := strings.Index(s, ":")
	if index == -1 {
		return "", "", errors.New("Malformed Header Text")
	}

	headerNameText := s[:index]
	headerValueText := s[index+1:]

	if strings.HasSuffix(headerNameText, " ") {
		return "", "", errors.New("Space between HeaderName and :")
	}

	headerName := strings.TrimSpace(headerNameText)
	if !validTokens([]byte(headerName)) {
		return "", "", errors.New("Invalid token found in HeaderName")
	}

	headerValue := strings.ReplaceAll(headerValueText, " ", "")
	return headerName, headerValue, nil
}
