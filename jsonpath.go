package jsonpath

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var rangeRegex = regexp.MustCompile(`^(-?\d+)?:(-?\d+)?$`)

type Error struct {
	Code string
	Msg  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Msg)
}

const (
	NotFound    = "not_found"
	InvalidPath = "invalid_path"
)

type segment struct {
	raw        string
	keys       []string
	indexes    []index
	isKey      bool
	isIndex    bool
	isWildcard bool
}

type index struct {
	idx      int
	start    int
	hasStart bool
	end      int
	hasEnd   bool
}

func Set(object interface{}, path string, value interface{}) error {
	pathParts, err := parsePath(path)
	if err != nil {
		return err
	}
	_, err = setNestedValues(object, pathParts, value)
	if err != nil {
		return err
	}
	return nil
}

func setNestedValues(object interface{}, path []segment, value interface{}) (interface{}, error) {
	var err error
	final := len(path) == 0
	if final {
		return value, nil
	}
	segment := path[0]
	fullKey := segment.raw

	switch obj := object.(type) {
	case map[string]interface{}:
		keys := []string{}
		if segment.isWildcard {
			for k := range obj {
				keys = append(keys, k)
			}
		} else {
			if segment.isIndex {
				return nil, &Error{NotFound, fmt.Sprintf("cannot set map with an index (%s)", fullKey)}
			}
			keys = segment.keys
		}

		for _, k := range keys {
			obj[k], err = setNestedValues(obj[k], path[1:], value)
		}
		return obj, err

	case []interface{}:

		var idxs []int
		if segment.isWildcard {
			idxs = makeRange(0, len(obj)-1)
		} else {
			if segment.isKey {
				return nil, &Error{NotFound, fmt.Sprintf("cannot set array with a key (%s)", fullKey)}
			}
			idxs, err = parseIndexes(segment.indexes, len(obj))
			if err != nil {
				return nil, err
			}
			obj = fillSlice(obj, idxs[len(idxs)-1])
		}

		for _, i := range idxs {
			obj[i], err = setNestedValues(obj[i], path[1:], value)
		}

		return obj, err

	default:
		if segment.isWildcard {
			return nil, &Error{NotFound, fmt.Sprintf("cannot set using a wildcard on a non-existing path (%s)", fullKey)}
		}
		if segment.isIndex {
			new := []interface{}{}
			parsed, err := parseIndexes(segment.indexes, 0)
			if err != nil {
				return nil, err
			}
			new = fillSlice(new, parsed[len(parsed)-1])
			for _, i := range parsed {
				new[i], err = setNestedValues(nil, path[1:], value)
			}
			return new, err

		} else {
			new := map[string]interface{}{}
			for _, k := range segment.keys {
				new[k], err = setNestedValues(nil, path[1:], value)
			}
			return new, err

		}
	}
}

func Get(object interface{}, path string) (interface{}, error) {
	pathParts, err := parsePath(path)
	if err != nil {
		return nil, err
	}
	value, err := getNestedValues(object, pathParts)
	if err != nil {
		return nil, err

	}
	return value, nil
}

func getNestedValues(object interface{}, path []segment) (interface{}, error) {
	final := len(path) == 0
	if final {
		return object, nil
	}
	var err error
	segment := path[0]
	fullKey := segment.raw

	result := []interface{}{}

	switch obj := object.(type) {
	case map[string]interface{}:
		var keys []string
		if segment.isWildcard {
			for k := range obj {
				keys = append(keys, k)
			}
		} else {
			if segment.isIndex {
				return nil, &Error{NotFound, fmt.Sprintf("cannot access map with an index (%s)", fullKey)}
			}
			keys = segment.keys
		}

		for _, k := range keys {
			if _, ok := obj[k]; !ok {
				return nil, &Error{NotFound, fmt.Sprintf("key does not exist (%s)", fullKey)}
			}
			temp, err := getNestedValues(obj[k], path[1:])
			if err != nil {
				return nil, err
			}
			result = append(result, temp)
		}

	case []interface{}:
		var idxs []int
		if segment.isWildcard {
			idxs = makeRange(0, len(obj)-1)
		} else {
			if segment.isKey {
				return nil, &Error{NotFound, fmt.Sprintf("cannot access array with a key (%s)", fullKey)}
			}
			idxs, err = parseIndexes(segment.indexes, len(obj))
			if err != nil {
				return nil, err
			}
		}

		for _, i := range idxs {
			if i >= len(obj) || i < 0 {
				return nil, &Error{NotFound, fmt.Sprintf("index out of range (%s)", fullKey)}
			}
			temp, err := getNestedValues(obj[i], path[1:])
			if err != nil {
				return nil, err
			}
			result = append(result, temp)
		}

	default:
		if !final {
			return nil, &Error{NotFound, "path not found"}
		}
		return obj, nil
	}

	if len(result) == 1 {
		return result[0], nil
	}
	return result, nil
}

func parsePath(path string) ([]segment, error) {
	segments := []segment{}
	var err error
	var key string
	var keyEnd bool
	var inBracket bool
	var inQuote bool
	var quoteChar rune

	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")
	for i, c := range path {
		if inQuote && c == quoteChar && lastChar(key) != "\\" {
			inQuote = false

		} else if !inQuote && (c == '\'' || c == '"') {
			if !inBracket {
				return nil, &Error{InvalidPath, "cannot use quotes outside of brackets"}
			}
			inQuote = true
			quoteChar = c
		}

		if c == '.' && !inQuote {
			if i == len(path)-1 {
				return nil, &Error{InvalidPath, "path cannot end with '.' separator"}
			}
			keyEnd = true
		}

		if c == '[' && !inQuote {
			if inBracket {
				return nil, &Error{InvalidPath, "missing closing bracket"}
			}
			inBracket = true
			if i != 0 {
				keyEnd = true
			}
		}

		if c == ']' && !inQuote {
			if !inBracket {
				return nil, &Error{InvalidPath, "missing opening bracket"}
			}
			inBracket = false
		}

		if unicode.IsSpace(c) && !inQuote && !inBracket {
			return nil, &Error{InvalidPath, "cannot use whitespace characters outside quotes and brackets"}
		}

		if keyEnd {
			segments, err = addSegment(key, segments)
			if err != nil {
				return nil, err
			}
			key = ""
			keyEnd = false
			if c == '.' {
				continue
			}
		}

		key += string(c)
	}

	if key != "" {
		segments, err = addSegment(key, segments)
		if err != nil {
			return nil, err
		}
	}

	if inBracket {
		return nil, &Error{InvalidPath, "missing closing bracket"}
	}
	if inQuote {
		return nil, &Error{InvalidPath, "missing closing quote"}
	}

	return segments, nil
}

func addSegment(key string, segments []segment) ([]segment, error) {
	keys, indexes, indexed, wildcard, err := parseKey(key)
	if err != nil {
		return segments, err
	}
	segments = append(segments, segment{
		raw:        key,
		keys:       keys,
		indexes:    indexes,
		isKey:      !indexed,
		isIndex:    indexed,
		isWildcard: wildcard,
	})
	return segments, nil
}

// Parses path keys
func parseKey(fullKey string) ([]string, []index, bool, bool, error) {
	var err error
	keys := []string{}
	indexes := []index{}

	if fullKey == "" {
		return keys, indexes, false, false, &Error{InvalidPath, "empty path segment"}
	}

	// Is a wildcard
	if fullKey == "*" {
		return keys, indexes, false, true, nil
	}

	// Check for square brackets
	if string(fullKey[0]) != "[" || string(fullKey[len(fullKey)-1]) != "]" {
		return []string{fullKey}, indexes, false, false, nil
	}

	key := fullKey[1 : len(fullKey)-1]

	if key == "" {
		return keys, indexes, false, false, &Error{InvalidPath, "empty path segment"}
	}

	// Split the key into it's parts
	var segment string
	var readSegment bool
	var quoted bool
	var quoteChar rune
	for _, c := range key {
		if readSegment {
			if !quoted {
				if unicode.IsSpace(c) {
					continue
				}
				if c == ',' {
					readSegment = false
					keys = append(keys, segment)
					segment = ""
					continue
				}
			}
			if quoted && c == quoteChar && lastChar(segment) != "\\" {
				segment = strings.ReplaceAll(segment, "\\"+string(quoteChar), string(quoteChar))
				quoted = false
			}
			segment += string(c)

		} else if !unicode.IsSpace(c) {
			readSegment = true
			if c == '\'' || c == '"' {
				quoteChar = c
				quoted = true
			}
			segment += string(c)
		}
	}

	if readSegment {
		if quoted {
			return keys, indexes, false, false, &Error{InvalidPath, "missing closing quote"}
		}
		keys = append(keys, segment)
	}

	for i, k := range keys {
		// Check for a wildcard
		if k == "*" {
			if len(keys) > 1 {
				return keys, indexes, false, false, &Error{InvalidPath, "cannot use a wildcard with a multi-select"}
			}
			return keys, indexes, false, true, nil
		}

		// If quoted string (treat as a map key)
		if len(k) >= 2 && string(k[0]) == "\"" && string(k[len(k)-1]) == "\"" {
			keys[i] = k[1 : len(k)-1]
			continue
		}
		if len(k) >= 2 && string(k[0]) == "'" && string(k[len(k)-1]) == "'" {
			keys[i] = k[1 : len(k)-1]
			continue
		}

		// Check if the key is an index
		idx, err := strconv.Atoi(k)
		if err == nil {
			indexes = append(indexes, index{idx: idx})
			continue
		}

		// Check if the key is a range
		rangeKey := rangeRegex.FindStringSubmatch(k)
		if len(rangeKey) > 0 {
			idx := index{}
			if rangeKey[1] != "" {
				start, err := strconv.Atoi(rangeKey[1])
				if err != nil {
					return keys, indexes, false, false, &Error{InvalidPath, "invalid range"}
				}
				idx.start = start
				idx.hasStart = true
			}
			if rangeKey[2] != "" {
				end, err := strconv.Atoi(rangeKey[2])
				if err != nil {
					return keys, indexes, false, false, &Error{InvalidPath, "invalid range"}
				}
				idx.end = end
				idx.hasEnd = true
			}
			indexes = append(indexes, idx)
		}
	}

	if len(indexes) == 0 {
		return keys, indexes, false, false, err
	}

	if len(indexes) != len(keys) {
		return keys, indexes, false, false, &Error{InvalidPath, "cannot specify both array indexes and map keys in a multi-select"}
	}

	return keys, indexes, true, false, err
}

func parseIndexes(indexes []index, length int) ([]int, error) {
	temp := map[int]struct{}{}
	parsed := []int{}
	for _, idx := range indexes {
		if !idx.hasStart && !idx.hasEnd {
			idx := wrapIndex(idx.idx, length)
			temp[idx] = struct{}{}
			continue
		}
		var start int
		var end int
		if idx.hasStart {
			start = wrapIndex(idx.start, length)
		}
		if idx.hasEnd {
			end = wrapIndex(idx.end, length) - 1
		} else {
			end = length - 1
		}
		if start == end {
			temp[start] = struct{}{}
			continue
		}
		if start > end {
			return parsed, &Error{InvalidPath, fmt.Sprintf("indexes out of range [%d:%d]", idx.start, idx.end)}
		}
		for _, i := range makeRange(start, end) {
			temp[i] = struct{}{}
		}
	}

	for i := range temp {
		parsed = append(parsed, i)
	}

	sort.Ints(parsed)
	return parsed, nil
}

func wrapIndex(idx, length int) int {
	if idx < 0 {
		idx = length + idx
	}
	return idx
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func fillSlice(slice []interface{}, max int) []interface{} {
	for i := len(slice); i <= max; i++ {
		slice = append(slice, nil)
	}
	return slice
}

func lastChar(val string) string {
	if len(val) == 0 {
		return ""
	}
	return string(val[len(val)-1])
}
