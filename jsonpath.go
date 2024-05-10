package jsonpath

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var rangeRegex = regexp.MustCompile(`^(-?\d+)?:(-?\d+)?$`)

type Compiled struct {
	raw      string
	segments []segment
	hasMulti bool

	// only allow setting values on existing paths
	strictPaths bool
	// query struct based off a tag instead of field names
	structTag    string
	structTagSet bool
}

type segment struct {
	raw         string
	keys        []string
	keysRefl    []reflect.Value
	indexes     []index
	isKey       bool
	isIndex     bool
	isWildcard  bool
	isRecursive bool
	isMulti     bool
}

type index struct {
	idx      int
	start    int
	hasStart bool
	end      int
	hasEnd   bool
}

type Error struct {
	Code string
	Msg  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Msg)
}

const (
	NotFound      = "not_found"
	InvalidPath   = "invalid_path"
	RecursiveMiss = "recursive_miss"
)

func (c *Compiled) EnableStrictPaths() {
	c.strictPaths = true
}

func (c *Compiled) UseStructTag(tag string) {
	c.structTag = tag
	c.structTagSet = true
}

func EnableStrictPaths() func(c *Compiled) {
	return func(c *Compiled) {
		c.EnableStrictPaths()
	}
}

func UseStructTag(tag string) func(c *Compiled) {
	return func(c *Compiled) {
		c.UseStructTag(tag)
	}
}

func (c *Compiled) Set(object interface{}, value interface{}) error {
	var valueSet bool
	_, err := c.setNestedValues(reflect.ValueOf(object), nil, c.segments, value, &valueSet)
	if err != nil {
		if err.Code != RecursiveMiss {
			return err
		}
		if !valueSet {
			return &Error{NotFound, err.Msg}
		}
	}
	return nil
}

func (c *Compiled) Get(object interface{}) (interface{}, error) {
	value, err := c.getNestedValues(reflect.ValueOf(object), c.segments)
	if err != nil {
		if err.Code != RecursiveMiss {
			return nil, err
		}
		if len(value) == 0 {
			return nil, &Error{NotFound, "path not found"}
		}
	}
	if !c.hasMulti && len(value) == 1 {
		return value[0], nil
	}
	return value, nil
}

func Set(object interface{}, path string, value interface{}, options ...func(*Compiled)) error {
	compiled, err := Compile(path, options...)
	if err != nil {
		return err
	}
	return compiled.Set(object, value)
}

func Get(object interface{}, path string, options ...func(*Compiled)) (interface{}, error) {
	compiled, err := Compile(path, options...)
	if err != nil {
		return nil, err
	}
	return compiled.Get(object)
}

func (c *Compiled) setNestedValues(object reflect.Value, objectType reflect.Type, path []segment, value interface{}, valueSet *bool) (reflect.Value, *Error) {
	var err *Error
	var temp reflect.Value

	final := len(path) == 0
	if final {
		*valueSet = true
		return reflect.ValueOf(value), nil
	}
	seg := path[0]
	fullKey := seg.raw

	if !object.IsValid() && objectType != nil {
		object = initNewValue(objectType).Elem()
	}

	var derefenced bool
	objectRef := object
	for objectRef.Kind() == reflect.Ptr || objectRef.Kind() == reflect.Interface {
		if objectRef.Kind() == reflect.Ptr {
			derefenced = true
			if objectRef.IsNil() {
				if c.strictPaths {
					return temp, &Error{NotFound, fmt.Sprintf("path not found (%s)", fullKey)}
				}
				objectRef.Set(initNewValue(objectRef.Type().Elem()))
			}
		}
		objectRef = objectRef.Elem()
	}

	if objectRef.IsValid() && objectRef.IsZero() {
		if c.strictPaths {
			return temp, &Error{NotFound, fmt.Sprintf("path not found (%s)", fullKey)}
		}
		if !objectRef.CanSet() {
			return temp, &Error{NotFound, fmt.Sprintf("object is not addressable (%s)", fullKey)}
		}
		objectRef.Set(initNewValue(objectRef.Type()).Elem())
	}

	switch objectRef.Kind() {
	case reflect.Map:
		var keys []reflect.Value
		if !objectRef.IsValid() {
			return temp, &Error{NotFound, fmt.Sprintf("map invalid (%s)", fullKey)}
		}
		elemType := objectRef.Type().Elem()
		keys, err = c.mapKeys(objectRef, seg)
		if err != nil {
			return temp, err
		}

		for _, k := range keys {
			nextObject := objectRef.MapIndex(k)
			if c.strictPaths && !nextObject.IsValid() {
				return temp, &Error{NotFound, fmt.Sprintf("key does not exist (%s)", fullKey)}
			}
			err = c.setCommon(nextObject, path, seg, value, valueSet, elemType,
				func(val reflect.Value) *Error {
					objectRef.SetMapIndex(k, val)
					return nil
				},
				func() bool {
					return contains(seg.keysRefl, k)
				},
			)
		}

	case reflect.Struct:
		var fields []string
		var segFields []string
		fields, segFields, err = c.structFields(objectRef, seg)
		if err != nil {
			return temp, err
		}
		for _, f := range fields {
			nextObject := objectRef.FieldByName(f)
			if !nextObject.IsValid() {
				return temp, &Error{NotFound, fmt.Sprintf("field does not exist (%s)", seg.raw)}
			}
			elemType, _ := objectRef.Type().FieldByName(f)
			err = c.setCommon(nextObject, path, seg, value, valueSet, elemType.Type,
				func(val reflect.Value) *Error {
					if !nextObject.CanSet() {
						return &Error{NotFound, fmt.Sprintf("struct field is not addressable (%s)", fullKey)}
					}
					nextObject.Set(val)
					return nil
				},
				func() bool {
					return slices.Contains(segFields, f)
				},
			)
		}

	case reflect.Slice, reflect.Array:
		var idxs []int
		var segIdxs []int
		elemType := objectRef.Type().Elem()
		idxs, segIdxs, err = c.sliceIndexes(objectRef, seg, c.strictPaths)
		if err != nil {
			return temp, err
		}
		objectRef = fillSlice(objectRef, idxs[len(idxs)-1])
		for _, i := range idxs {
			nextObject := objectRef.Index(i)
			if !nextObject.IsValid() {
				return temp, &Error{NotFound, fmt.Sprintf("index out of range (%d)", i)}
			}
			err = c.setCommon(nextObject, path, seg, value, valueSet, elemType,
				func(val reflect.Value) *Error {
					if !nextObject.CanSet() {
						return &Error{NotFound, fmt.Sprintf("slice index is not addressable (%s)", fullKey)}
					}
					nextObject.Set(val)
					return nil
				},
				func() bool {
					return slices.Contains(segIdxs, i)
				},
			)
		}

	default:
		if seg.isRecursive {
			return temp, &Error{RecursiveMiss, fmt.Sprintf("path not found (%s)", fullKey)}
		}
		if c.strictPaths || seg.isWildcard {
			return temp, &Error{NotFound, fmt.Sprintf("path not found (%s)", fullKey)}
		}
		if seg.isIndex {
			new := reflect.ValueOf([]interface{}{})
			parsed, err := parseIndexes(seg.indexes, 0, false)
			if err != nil {
				return temp, err
			}
			new = fillSlice(new, parsed[len(parsed)-1])
			for _, i := range parsed {
				nextObject := new.Index(i)
				temp, err = c.setNestedValues(nextObject, nil, path[1:], value, valueSet)
				if err != nil {
					return temp, err
				}
				if temp.IsValid() {
					nextObject.Set(temp)
				}
			}
			return new, err

		} else {
			new := reflect.ValueOf(map[string]interface{}{})
			for _, k := range seg.keysRefl {
				temp, err = c.setNestedValues(new.MapIndex(k), nil, path[1:], value, valueSet)
				if err != nil {
					return temp, err
				}
				if temp.IsValid() {
					new.SetMapIndex(k, temp)
				}
			}
			return new, err
		}
	}

	if derefenced {
		return object, err
	}
	return objectRef, err
}

func initNewValue(t reflect.Type) reflect.Value {
	switch t.Kind() {
	case reflect.Map:
		ptr := reflect.New(t)
		ptr.Elem().Set(reflect.MakeMap(t))
		return ptr
	case reflect.Slice, reflect.Array:
		ptr := reflect.New(t)
		ptr.Elem().Set(reflect.MakeSlice(t, 0, 0))
		return ptr
	default:
		return reflect.New(t)
	}
}

func (c *Compiled) getNestedValues(object reflect.Value, path []segment) ([]interface{}, *Error) {
	var err *Error
	var temp []interface{}

	final := len(path) == 0
	if final {
		if object.IsValid() {
			return []interface{}{object.Interface()}, nil
		}
		return []interface{}{nil}, nil
	}
	seg := path[0]
	fullKey := seg.raw

	for object.Kind() == reflect.Ptr || object.Kind() == reflect.Interface {
		object = object.Elem()
	}

	result := []interface{}{}

	if !object.IsValid() {
		return result, &Error{NotFound, fmt.Sprintf("path not found (%s)", seg.raw)}
	}

	switch object.Kind() {
	case reflect.Map:
		var keys []reflect.Value
		keys, err = c.mapKeys(object, seg)
		if err != nil {
			return temp, err
		}
		for _, k := range keys {
			nextObject := object.MapIndex(k)
			if !nextObject.IsValid() {
				return temp, &Error{NotFound, fmt.Sprintf("key does not exist (%s)", seg.raw)}
			}
			result, err = c.getCommon(nextObject, path, seg, result, func() bool {
				return contains(seg.keysRefl, k)
			})
		}

	case reflect.Struct:
		var fields []string
		var segFields []string
		fields, segFields, err = c.structFields(object, seg)
		if err != nil {
			return temp, err
		}
		for _, f := range fields {
			nextObject := object.FieldByName(f)
			if !nextObject.IsValid() {
				return temp, &Error{NotFound, fmt.Sprintf("field does not exist (%s)", seg.raw)}
			}
			result, err = c.getCommon(nextObject, path, seg, result, func() bool {
				return slices.Contains(segFields, f)
			})
		}

	case reflect.Slice, reflect.Array:
		var idxs []int
		var segIdxs []int
		idxs, segIdxs, err = c.sliceIndexes(object, seg, true)
		if err != nil {
			return temp, err
		}
		for _, i := range idxs {
			nextObject := object.Index(i)
			if !nextObject.IsValid() {
				return temp, &Error{NotFound, fmt.Sprintf("index out of range (%d)", i)}
			}
			result, err = c.getCommon(nextObject, path, seg, result, func() bool {
				return slices.Contains(segIdxs, i)
			})
		}

	default:
		if seg.isRecursive {
			return nil, &Error{RecursiveMiss, fmt.Sprintf("path not found (%s)", fullKey)}
		}
		return nil, &Error{NotFound, fmt.Sprintf("path not found (%s)", fullKey)}
	}

	return result, err
}

func (c *Compiled) setCommon(
	nextObject reflect.Value,
	path []segment,
	seg segment,
	value interface{},
	valueSet *bool,
	elemType reflect.Type,
	setValue func(reflect.Value) *Error,
	inSegment func() bool,
) *Error {
	var err *Error
	var temp reflect.Value
	nextPath := path[1:]
	if seg.isRecursive && !inSegment() {
		nextPath = path
	}
	temp, err = c.setNestedValues(nextObject, elemType, nextPath, value, valueSet)
	if err != nil && err.Code != RecursiveMiss {
		return err
	}
	if temp.IsValid() {
		if !temp.Type().AssignableTo(elemType) {
			return &Error{NotFound, fmt.Sprintf("cannot assign type %s to type %s", temp.Type().String(), elemType.String())}
		}
		err := setValue(temp)
		if err != nil {
			return err
		}
	}
	return err
}

func (c *Compiled) getCommon(
	nextObject reflect.Value,
	path []segment,
	seg segment,
	result []interface{},
	inSegment func() bool,
) ([]interface{}, *Error) {
	nextPaths := [][]segment{}
	if seg.isRecursive {
		nextPaths = append(nextPaths, path)
	}
	if !seg.isRecursive || seg.isWildcard || inSegment() {
		nextPaths = append(nextPaths, path[1:])
	}
	var err *Error
	var temp []interface{}
	for _, p := range nextPaths {
		temp, err = c.getNestedValues(nextObject, p)
		if err != nil && err.Code != RecursiveMiss {
			return result, err
		}
		if err == nil || temp != nil {
			result = append(result, temp...)
		}
	}
	return result, err
}

func (c *Compiled) mapKeys(object reflect.Value, seg segment) ([]reflect.Value, *Error) {
	if seg.isWildcard || seg.isRecursive {
		return object.MapKeys(), nil
	} else {
		if seg.isIndex {
			return nil, &Error{NotFound, fmt.Sprintf("cannot access map with an index (%s)", seg.raw)}
		}
		return seg.keysRefl, nil
	}
}

func (c *Compiled) sliceIndexes(object reflect.Value, seg segment, capLength bool) ([]int, []int, *Error) {
	var err *Error
	var idxs []int
	var segIdxs []int
	if seg.isWildcard || seg.isRecursive {
		idxs = makeRange(0, object.Len()-1)
	}
	if !seg.isWildcard {
		if !seg.isRecursive && seg.isKey {
			return nil, nil, &Error{NotFound, fmt.Sprintf("cannot access array with a key (%s)", seg.raw)}
		}
		segIdxs, err = parseIndexes(seg.indexes, object.Len(), capLength)
		if err != nil {
			return nil, nil, err
		}
		if !seg.isRecursive {
			idxs = segIdxs
		}
	}
	return idxs, segIdxs, nil
}

func (c *Compiled) structFields(object reflect.Value, seg segment) ([]string, []string, *Error) {
	var fields []string
	var segFields []string
	tagMap := map[string]string{}
	if seg.isWildcard || seg.isRecursive || c.structTagSet {
		objType := object.Type()
		for i := 0; i < object.NumField(); i += 1 {
			field := objType.Field(i)
			fields = append(fields, field.Name)
			if c.structTagSet {
				if val, ok := field.Tag.Lookup(c.structTag); ok {
					tagMap[val] = field.Name
				}
			}
		}
	}
	if !seg.isWildcard {
		if seg.isIndex {
			return nil, nil, &Error{NotFound, fmt.Sprintf("cannot access struct field with an index (%s)", seg.raw)}
		}
		segFields = seg.keys
		if c.structTagSet {
			for i, k := range segFields {
				segFields[i] = tagMap[k]
			}
		}
		if !seg.isRecursive {
			fields = segFields
		}
	}
	return fields, segFields, nil
}

func Compile(path string, options ...func(*Compiled)) (*Compiled, error) {
	compiled := Compiled{
		raw:      path,
		segments: []segment{},
	}

	for _, option := range options {
		option(&compiled)
	}

	var key string
	var keyEnd bool
	var inBracket bool
	var inQuote bool
	var quoteChar rune

	if path == "" {
		return &compiled, &Error{InvalidPath, "empty path"}
	}

	path = strings.TrimPrefix(path, "$")
	if path == "." {
		return &compiled, nil
	}

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

		if c == '.' && !inQuote && key != "" && key != "." {
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
			if i != 0 && key != "." && key != ".." {
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
			segment, err := parseKey(key)
			if err != nil {
				return nil, err
			}
			compiled.segments = append(compiled.segments, segment)
			compiled.hasMulti = compiled.hasMulti || segment.isMulti

			key = ""
			keyEnd = false
		}

		key += string(c)
	}

	if key != "" {
		segment, err := parseKey(key)
		if err != nil {
			return nil, err
		}
		compiled.segments = append(compiled.segments, segment)
		compiled.hasMulti = compiled.hasMulti || segment.isMulti
	}

	if inBracket {
		return nil, &Error{InvalidPath, "missing closing bracket"}
	}
	if inQuote {
		return nil, &Error{InvalidPath, "missing closing quote"}
	}

	return &compiled, nil
}

// Parses path keys
func parseKey(fullKey string) (segment, error) {
	var err error
	result := segment{
		raw:      fullKey,
		keys:     []string{},
		keysRefl: []reflect.Value{},
		indexes:  []index{},
	}

	fullKey = strings.TrimPrefix(fullKey, ".")

	if fullKey == "" {
		return result, &Error{InvalidPath, "empty path segment"}
	}

	// Is a wildcard
	if fullKey == "*" {
		result.isWildcard = true
		result.isMulti = true
		return result, nil
	}

	// Is recursive
	if string(fullKey[0]) == "." {
		result.isRecursive = true
		result.isMulti = true
		fullKey = strings.TrimPrefix(fullKey, ".")
		if fullKey == "" || string(fullKey[0]) == "." {
			return result, &Error{InvalidPath, "invalid recursive path"}
		}
	}

	// Check for square brackets
	if string(fullKey[0]) != "[" || string(fullKey[len(fullKey)-1]) != "]" {
		result.isKey = true
		result.addKeys([]string{fullKey})
		return result, nil
	}

	key := strings.TrimSpace(fullKey[1 : len(fullKey)-1])

	if key == "" {
		return result, &Error{InvalidPath, "empty path segment"}
	}

	keys := []string{}

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
			return result, &Error{InvalidPath, "missing closing quote"}
		}
		keys = append(keys, segment)
	}

	for i, k := range keys {
		// Check for a wildcard
		if k == "*" {
			if len(keys) > 1 {
				return result, &Error{InvalidPath, "cannot use a wildcard with a multi-select"}
			}
			result.isWildcard = true
			result.isMulti = true
			return result, nil
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
			result.indexes = append(result.indexes, index{idx: idx})
			continue
		}

		// Check if the key is a range
		rangeKey := rangeRegex.FindStringSubmatch(k)
		if len(rangeKey) > 0 {
			idx := index{}
			if rangeKey[1] != "" {
				start, err := strconv.Atoi(rangeKey[1])
				if err != nil {
					return result, &Error{InvalidPath, "invalid range"}
				}
				idx.start = start
				idx.hasStart = true
			}
			if rangeKey[2] != "" {
				end, err := strconv.Atoi(rangeKey[2])
				if err != nil {
					return result, &Error{InvalidPath, "invalid range"}
				}
				idx.end = end
				idx.hasEnd = true
			}
			result.indexes = append(result.indexes, idx)
			result.isMulti = true
			if idx.hasStart && idx.hasEnd && idx.start == idx.end {
				return result, &Error{InvalidPath, fmt.Sprintf("invalid index range [%d:%d]", idx.start, idx.end)}
			}
		}
	}

	result.isMulti = result.isMulti || len(keys) > 1

	if len(result.indexes) == 0 {
		result.isKey = true
		result.addKeys(keys)
		return result, nil
	}

	result.isIndex = true

	if len(result.indexes) != len(keys) {
		return result, &Error{InvalidPath, "cannot specify both array indexes and map keys in a multi-select"}
	}

	return result, err
}

func (s *segment) addKeys(keys []string) {
	s.keys = keys
	for _, k := range keys {
		s.keysRefl = append(s.keysRefl, reflect.ValueOf(k))
	}
}

func parseIndexes(indexes []index, length int, capLength bool) ([]int, *Error) {
	var err *Error
	temp := map[int]struct{}{}
	parsed := []int{}
	for _, idx := range indexes {
		if !idx.hasStart && !idx.hasEnd {
			i, err := wrapIndex(idx.idx, length, capLength)
			if err != nil {
				return nil, err
			}
			temp[i] = struct{}{}
			continue
		}
		var start int
		var end int
		if idx.hasStart {
			start, err = wrapIndex(idx.start, length, capLength)
			if err != nil {
				return nil, err
			}
		}
		if idx.hasEnd {
			end, err = wrapIndex(idx.end-1, length, capLength)
			if err != nil {
				return nil, err
			}
		} else {
			end = length - 1
		}
		if start == end {
			temp[start] = struct{}{}
			continue
		}
		if start > end {
			return parsed, &Error{NotFound, fmt.Sprintf("indexes out of range [%d:%d]", idx.start, idx.end)}
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

func wrapIndex(idx, length int, capLength bool) (int, *Error) {
	tmp := idx
	if tmp < 0 {
		tmp = length + tmp
	}
	if tmp < 0 || (capLength && tmp >= length) {
		return tmp, &Error{NotFound, fmt.Sprintf("index out of range (%d)", idx)}
	}
	return tmp, nil
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func fillSlice(slice reflect.Value, lastIndex int) reflect.Value {
	max := lastIndex + 1
	if max <= slice.Len() {
		return slice
	}
	new := reflect.MakeSlice(slice.Type(), max-slice.Len(), max-slice.Len())
	new = reflect.AppendSlice(slice, new)
	if slice.CanSet() {
		slice.Set(new)
		return slice
	}
	return new
}

func lastChar(val string) string {
	if len(val) == 0 {
		return ""
	}
	return string(val[len(val)-1])
}

func contains(slice []reflect.Value, val reflect.Value) bool {
	for _, v := range slice {
		if v.Comparable() && v.Interface() == val.Interface() {
			return true
		}
	}
	return false
}
