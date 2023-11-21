package jsonpath

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var indexRegex = regexp.MustCompile(`\[[^]]+\]`)
var invalidPathRegex = regexp.MustCompile(`^\]|\[\[|\]\]|\[\]|\][^\[]|\[[^]]*$`)

func Set(object interface{}, path string, value interface{}) error {
	pathParts, err := tokenizePath(path)
	if err != nil {
		return err
	}
	_, err = setNestedValues(object, pathParts, value)
	if err != nil {
		return fmt.Errorf("%w: %s", err, path)
	}
	return nil
}

func setNestedValues(object interface{}, path []string, value interface{}) (interface{}, error) {
	var err error
	final := len(path) == 0
	if final {
		return value, nil
	}
	key := path[0]
	isIndex, idx, key := isIndexKey(key)

	switch obj := object.(type) {
	case map[string]interface{}:
		if isIndex {
			return nil, fmt.Errorf("map type cannot be set with index (%s)", key)
		}
		obj[key], err = setNestedValues(obj[key], path[1:], value)
		return obj, err

	case []interface{}:
		if !isIndex {
			return nil, fmt.Errorf("slice type cannot be set with key (%s)", key)
		}
		idx, err = wrapIndex(idx, len(obj))
		if err != nil {
			return nil, err
		}
		obj = fillSlice(obj, idx)
		obj[idx], err = setNestedValues(obj[idx], path[1:], value)
		return obj, err

	default:
		if isIndex {
			idx, err = wrapIndex(idx, 0)
			if err != nil {
				return nil, err
			}
			new := fillSlice([]interface{}{}, idx)
			new[idx], err = setNestedValues(nil, path[1:], value)
			return new, err

		} else {
			new := map[string]interface{}{}
			new[key], err = setNestedValues(nil, path[1:], value)
			return new, err

		}
	}
}

func Get(object interface{}, path string) (interface{}, error) {
	pathParts, err := tokenizePath(path)
	if err != nil {
		return nil, err
	}
	value, err := getNestedValues(object, pathParts)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, path)

	}
	return value, nil
}

func getNestedValues(object interface{}, path []string) (interface{}, error) {
	final := len(path) == 0
	if final {
		return object, nil
	}
	key := path[0]
	isIndex, idx, key := isIndexKey(key)

	switch obj := object.(type) {
	case map[string]interface{}:
		if isIndex {
			return nil, fmt.Errorf("map type cannot be accessed with index (%s)", key)
		}
		if _, ok := obj[key]; !ok {
			return nil, fmt.Errorf("key does not exist (%s)", key)
		}
		return getNestedValues(obj[key], path[1:])

	case []interface{}:
		if !isIndex {
			return nil, fmt.Errorf("slice type cannot be accessed with key (%s)", key)
		}
		if idx >= len(obj) {
			return nil, fmt.Errorf("index out of range (%s)", key)
		}
		idx, err := wrapIndex(idx, len(obj))
		if err != nil {
			return nil, err
		}
		return getNestedValues(obj[idx], path[1:])

	default:
		if !final {
			return nil, errors.New("path does not exist")
		}
		return obj, nil
	}
}

func tokenizePath(path string) ([]string, error) {
	var tokens []string
	for _, stem := range strings.Split(path, ".") {
		if stem == "" {
			continue
		}
		found := indexRegex.FindAllString(stem, -1)
		if found == nil {
			tokens = append(tokens, stem)
		} else {
			indexStart := strings.Index(stem, "[")
			if indexStart > 0 {
				tokens = append(tokens, stem[:indexStart])
			}
			if invalidPathRegex.MatchString(stem[indexStart:]) {
				return nil, fmt.Errorf("invalid path: %s", path)
			}
			tokens = append(tokens, found...)
		}

	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty path: %s", path)
	}
	return tokens, nil
}

// Check for keys in square brackets
func isIndexKey(key string) (bool, int, string) {
	if len(key) < 3 {
		return false, 0, key
	}

	// Check for square brackets
	if string(key[0]) != "[" || string(key[len(key)-1]) != "]" {
		return false, 0, key
	}

	key = key[1 : len(key)-1]

	// If quoted string (treated as a map key)
	if len(key) >= 3 && string(key[0]) == "\"" && string(key[len(key)-1]) == "\"" {
		return false, 0, key[1 : len(key)-1]
	}
	if len(key) >= 3 && string(key[0]) == "'" && string(key[len(key)-1]) == "'" {
		return false, 0, key[1 : len(key)-1]
	}

	idx, err := strconv.Atoi(key)
	if err != nil {
		return false, 0, key
	}
	return true, idx, key
}

func fillSlice(slice []interface{}, max int) []interface{} {
	for i := len(slice); i <= max; i++ {
		slice = append(slice, nil)
	}
	return slice
}

func wrapIndex(idx, length int) (int, error) {
	if idx < 0 {
		idxWrapped := length + idx
		if idxWrapped < 0 {
			return idx, fmt.Errorf("index out of range (%d)", idx)
		}
		idx = idxWrapped
	}
	return idx, nil
}
