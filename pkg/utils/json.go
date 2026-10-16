package utils

import (
	"encoding/json"
	"strconv"
	"strings"
)

// NormalizeJSONBytes keeps valid JSON as-is and wraps plain text as a JSON string.
func NormalizeJSONBytes(data []byte) []byte {
	if len(data) == 0 || json.Valid(data) {
		return data
	}

	normalized, err := json.Marshal(string(data))
	if err != nil {
		return data
	}
	return normalized
}

type JsonStringOrStrings []string

// MarshalJSON string or strings
// if len(u) == 0, return null
// if len(u) == 1, return string
// if len(u) > 1, return strings
func (u JsonStringOrStrings) MarshalJSON() ([]byte, error) {
	if len(u) == 0 {
		return []byte("null"), nil
	} else if len(u) == 1 {
		return json.Marshal(u[0])
	}

	return json.Marshal(u)
}

// UnmarshalJSON string or strings
// if data is null, set u to nil
// if data is string, set u to [url]
// if data is strings, set u to urls
func (u *JsonStringOrStrings) UnmarshalJSON(data []byte) error {
	var urls []string
	if err := json.Unmarshal(data, &urls); err == nil {
		*u = urls
		return nil
	}

	var url string
	if err := json.Unmarshal(data, &url); err != nil {
		return err
	}
	if url == "" {
		*u = nil
		return nil
	}

	*u = JsonStringOrStrings{url}
	return nil
}

type JsonInt64String int64

// MarshalJSON int64 string, always return int64
func (i JsonInt64String) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(i), 10)), nil
}

// UnmarshalJSON int64 string to int64, int64 to int64
func (i *JsonInt64String) UnmarshalJSON(b []byte) error {

	s := strings.TrimSpace(string(b))

	if s == "null" {
		*i = 0
	} else {
		s = strings.Trim(s, "\"")
		v, _ := strconv.ParseInt(s, 10, 64)
		*i = JsonInt64String(v)
	}
	return nil
}
