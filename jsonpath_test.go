package jsonpath

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

var runTest = os.Getenv("TEST_NAME")

var example = `
{
	"key1": {
		"key2": {
			"key3": {
				"key4": {
					"key5": 123
				}
			}
		}
	},
	"key2": {
		"array": [
			{
				"subkey": "val"
			},
			456,
			true
		]
	},
	"key3": {
		"array": [
			"val0",
			"val1",
			"val2",
			"val3",
			"val4",
			"val5"
		],
		"map": {
			"key1": "val1",
			"key2": "val2",
			"key3": "val3"
		}
	},
	"key4": [
		{
			"key1": "val1"
		},
		{
			"key1": "val2"
		},
		{
			"key1": "val3"
		}
	],
	"key5": {
		"'single'": "single",
		"\"double\"": "double",
		"  spaces  ": "spaces",
		"][.,": "specials",
		"null_value": null,
		"empty_slice": [],
		"empty_map": {},
		"int": 123,
		"float": 1.23
	},
	"key6": {
		"recursive": "val1",
		"key7": {
			"recursive": "val2",
			"key8": {
				"recursive": "val3"
			},
			"key9": [
				{
					"recursive": "val4"
				},
				{
					"recursive": "val5"
				}
			]
		}
	},
	"key7": {
		"recursive": [
			{
				"recursive": {
					"recursive": true
				}
			}
		],
		"arrays": {
			"a": [
				"val1",
				"val2"
			],
			"b": [
				"val3",
				"val4"
			],
			"c": [
				"val5",
				"val6"
			]
		}
	}
}`

func getData() interface{} {
	var data interface{}
	err := json.Unmarshal([]byte(example), &data)
	if err != nil {
		panic(err)
	}
	return data
}

func TestCompile(t *testing.T) {
	type args struct {
		path string
	}
	tests := map[string][]struct {
		name         string
		args         args
		wantSegments int
		wantErr      bool
		wantErrCode  string
		wantErrMsg   string
	}{
		"success": {
			{
				name: "base-1",
				args: args{
					path: "$",
				},
				wantSegments: 0,
			},
			{
				name: "base-2",
				args: args{
					path: "$.",
				},
				wantSegments: 0,
			},
			{
				name: "base-3",
				args: args{
					path: ".",
				},
				wantSegments: 0,
			},
			{
				name: "dot-notation",
				args: args{
					path: ".key1.key2",
				},
				wantSegments: 2,
			},
			{
				name: "bracket-notation",
				args: args{
					path: "$['key1']['key2']",
				},
				wantSegments: 2,
			},
			{
				name: "mixed-notation-1",
				args: args{
					path: "$.key1['key2']",
				},
				wantSegments: 2,
			},
			{
				name: "mixed-notation-2",
				args: args{
					path: "['key1'].key2",
				},
				wantSegments: 2,
			},
			{
				name: "mixed-notation-3",
				args: args{
					path: "$.key1.['key2'].key3[\"key4\"]",
				},
				wantSegments: 4,
			},
			{
				name: "mulit-keys",
				args: args{
					path: ".key1['key2','key3']",
				},
				wantSegments: 2,
			},
			{
				name: "array-1",
				args: args{
					path: "[0,1,2]",
				},
				wantSegments: 1,
			},
			{
				name: "array-2",
				args: args{
					path: "[0,1,2]",
				},
				wantSegments: 1,
			},
			{
				name: "index-range-1",
				args: args{
					path: "[0:1]",
				},
				wantSegments: 1,
			},
			{
				name: "index-range-2",
				args: args{
					path: "[0:-1]",
				},
				wantSegments: 1,
			},
			{
				name: "index-range-3",
				args: args{
					path: "[0:]",
				},
				wantSegments: 1,
			},
			{
				name: "index-range-4",
				args: args{
					path: "[:-2]",
				},
				wantSegments: 1,
			},
			{
				name: "multi-index",
				args: args{
					path: "[0, 1:3, -5:, :10]",
				},
				wantSegments: 1,
			},
			{
				name: "wildcard-1",
				args: args{
					path: "$.*",
				},
				wantSegments: 1,
			},
			{
				name: "wildcard-2",
				args: args{
					path: "$.[*]",
				},
				wantSegments: 1,
			},
			{
				name: "wildcard-3",
				args: args{
					path: "$.key1.*.key2",
				},
				wantSegments: 3,
			},
			{
				name: "wildcard-4",
				args: args{
					path: "$.key1.*[0]",
				},
				wantSegments: 3,
			},
			{
				name: "recursive-1",
				args: args{
					path: "$..key1",
				},
				wantSegments: 1,
			},
			{
				name: "recursive-2",
				args: args{
					path: "key1.key2..key3",
				},
				wantSegments: 3,
			},
			{
				name: "recursive-3",
				args: args{
					path: "..key1[0:3]",
				},
				wantSegments: 2,
			},
			{
				name: "complex-1",
				args: args{
					path: "$.key1[0, 1:5]..key2.*.[ 'key3' , \"key4\", '*'][*]",
				},
				wantSegments: 6,
			},
			{
				name: "complex-2",
				args: args{
					path: "$..key1.*.*[-1]..key2[ 'key3', 'key4' ]..[0:10]",
				},
				wantSegments: 7,
			},
		},
		"errors": {
			{
				name: "empty-path",
				args: args{
					path: "",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "empty path",
			},
			{
				name: "invalid-whitespace-1",
				args: args{
					path: "key1. .key2",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "cannot use whitespace characters outside quotes and brackets",
			},
			{
				name: "invalid-whitespace-2",
				args: args{
					path: "key1.   key2",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "cannot use whitespace characters outside quotes and brackets",
			},
			{
				name: "invalid-whitespace-3",
				args: args{
					path: "$. []",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "cannot use whitespace characters outside quotes and brackets",
			},
			{
				name: "invalid-whitespace-4",
				args: args{
					path: "$.key1.key2.   ",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "cannot use whitespace characters outside quotes and brackets",
			},
			{
				name: "empty-bracket",
				args: args{
					path: "key1.key2[]",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "empty path segment",
			},
			{
				name: "missing-closing-bracket",
				args: args{
					path: "key1.key2['test'",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "missing closing bracket",
			},
			{
				name: "missing-opening-bracket",
				args: args{
					path: "key1.key2[0]]",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "missing opening bracket",
			},
			{
				name: "missing-closing-quote-1",
				args: args{
					path: "key1.key2['test]",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "missing closing quote",
			},
			{
				name: "missing-closing-quote-2",
				args: args{
					path: "key1.key2['test'][']",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "missing closing quote",
			},
			{
				name: "missing-closing-quote-3",
				args: args{
					path: "key1.key2['test\"][\\'\"]",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "missing closing quote",
			},
			{
				name: "quotes-outside-brackets",
				args: args{
					path: "key1.key2.'test'",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "cannot use quotes outside of brackets",
			},
			{
				name: "end-with-separator-1",
				args: args{
					path: "key1.key2.key3.",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "path cannot end with '.' separator",
			},
			{
				name: "end-with-separator-2",
				args: args{
					path: "key1.key2.key3[0].",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "path cannot end with '.' separator",
			},
			{
				name: "invalid-recursive-1",
				args: args{
					path: "key1...key2",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "invalid recursive path",
			},
			{
				name: "invalid-recursive-2",
				args: args{
					path: "... ..key2",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "invalid recursive path",
			},
			{
				name: "invalid-recursive-3",
				args: args{
					path: "key1.key2..",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "invalid recursive path",
			},
			{
				name: "invalid-recursive-4",
				args: args{
					path: "..",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "invalid recursive path",
			},
			{
				name: "invalid-recursive-5",
				args: args{
					path: "$..",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "invalid recursive path",
			},
			{
				name: "invalid-index-range-1",
				args: args{
					path: "$.test[1:1]",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "invalid index range",
			},
			{
				name: "invalid-index-range-2",
				args: args{
					path: "$.test[ 0, 1, 2:2]",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "invalid index range",
			},
		},
	}
	for groupName, group := range tests {
		for _, tt := range group {
			testName := fmt.Sprintf("%s-%s", groupName, tt.name)
			if runTest != "" && testName != runTest {
				continue
			}
			t.Run(testName, func(t *testing.T) {
				got, err := Compile(tt.args.path)
				if (err != nil) != tt.wantErr {
					t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErr {
					if err.(*Error).Code != tt.wantErrCode {
						t.Errorf("Get() errCode = %v, wantCode %v", err.(*Error).Code, tt.wantErrCode)
					}
					if !strings.Contains(err.Error(), tt.wantErrMsg) {
						t.Errorf("Get() errMsg = %v, wantMsg %v", err.(*Error).Msg, tt.wantErrMsg)
					}
					return
				}

				if len(got.segments) != tt.wantSegments {
					t.Errorf("Segments = %v, want %v", len(got.segments), tt.wantSegments)
				}
			})
		}
	}
}

func TestGet(t *testing.T) {
	data := getData()

	type args struct {
		object interface{}
		path   string
	}
	tests := map[string][]struct {
		name        string
		args        args
		want        interface{}
		wantJson    string
		sortResult  bool
		wantErr     bool
		wantErrCode string
		wantErrMsg  string
	}{
		"base": {
			{
				name: "get-whole-1",
				args: args{
					object: data,
					path:   "$",
				},
				want:    data,
				wantErr: false,
			},
			{
				name: "get-whole-2",
				args: args{
					object: data,
					path:   "$.",
				},
				want:    data,
				wantErr: false,
			},
			{
				name: "get-whole-3",
				args: args{
					object: data,
					path:   ".",
				},
				want:    data,
				wantErr: false,
			},
		},
		"map-access": {
			{
				name: "dot-notation",
				args: args{
					object: data,
					path:   "key1.key2.key3.key4.key5",
				},
				want:    float64(123),
				wantErr: false,
			},
			{
				name: "bracket-notation",
				args: args{
					object: data,
					path:   "[key1][key2][key3][key4][key5]",
				},
				want:    float64(123),
				wantErr: false,
			},
			{
				name: "mixed-notation",
				args: args{
					object: data,
					path:   "key1[key2].key3[key4][key5]",
				},
				want:    float64(123),
				wantErr: false,
			},
			{
				name: "mixed-notation-quotes",
				args: args{
					object: data,
					path:   "key1['key2'].key3[\"key4\"][key5]",
				},
				want:    float64(123),
				wantErr: false,
			},
			{
				name: "root-notation",
				args: args{
					object: data,
					path:   "$.key1.key2.key3.key4.key5",
				},
				want:    float64(123),
				wantErr: false,
			},
		},
		"array-access": {
			{
				name: "subkey-1",
				args: args{
					object: data,
					path:   "key2.array[0].subkey",
				},
				want:    "val",
				wantErr: false,
			},
			{
				name: "subkey-2",
				args: args{
					object: data,
					path:   "key2.array[0][subkey]",
				},
				want:    "val",
				wantErr: false,
			},
			{
				name: "subkey-3",
				args: args{
					object: data,
					path:   "key2.array[0]['subkey']",
				},
				want:    "val",
				wantErr: false,
			},
			{
				name: "index-1",
				args: args{
					object: data,
					path:   "key2.array[1]",
				},
				want:    float64(456),
				wantErr: false,
			},
			{
				name: "index-2",
				args: args{
					object: data,
					path:   "key2.array[2]",
				},
				want:    true,
				wantErr: false,
			},
			{
				name: "negative-index-1",
				args: args{
					object: data,
					path:   "key2.array[-1]",
				},
				want:    true,
				wantErr: false,
			},
			{
				name: "negative-index-2",
				args: args{
					object: data,
					path:   "key2.array[-2]",
				},
				want:    float64(456),
				wantErr: false,
			},
			{
				name: "negative-index-subkey",
				args: args{
					object: data,
					path:   "key2.array[-3].subkey",
				},
				want:    "val",
				wantErr: false,
			},
		},
		"multi-select": {
			{
				name: "array-1",
				args: args{
					object: data,
					path:   "key3.array[0,1,2]",
				},
				want: []interface{}{
					"val0",
					"val1",
					"val2",
				},
				wantErr: false,
			},
			{
				name: "array-2",
				args: args{
					object: data,
					path:   "key3.array[0,2]",
				},
				want: []interface{}{
					"val0",
					"val2",
				},
				wantErr: false,
			},
			{
				name: "array-3",
				args: args{
					object: data,
					path:   "key3.array[1,2]",
				},
				want: []interface{}{
					"val1",
					"val2",
				},
				wantErr: false,
			},
			{
				name: "map-1",
				args: args{
					object: data,
					path:   "key3.map['key1','key2','key3']",
				},
				want: []interface{}{
					"val1",
					"val2",
					"val3",
				},
				wantErr: false,
			},
			{
				name: "map-2",
				args: args{
					object: data,
					path:   "key3.map['key1','key3']",
				},
				want: []interface{}{
					"val1",
					"val3",
				},
				wantErr: false,
			},
			{
				name: "map-3",
				args: args{
					object: data,
					path:   "key3.map['key2','key3']",
				},
				want: []interface{}{
					"val2",
					"val3",
				},
				wantErr: false,
			},
			{
				name: "map-4",
				args: args{
					object: data,
					path:   "key3.map[ key1, 'key2', \"key3\" ]",
				},
				want: []interface{}{
					"val1",
					"val2",
					"val3",
				},
				wantErr: false,
			},
		},
		"index-range": {
			{
				name: "array-1",
				args: args{
					object: data,
					path:   "key3.array[0:5]",
				},
				want: []interface{}{
					"val0",
					"val1",
					"val2",
					"val3",
					"val4",
				},
				wantErr: false,
			},
			{
				name: "array-2",
				args: args{
					object: data,
					path:   "key3.array[1:4]",
				},
				want: []interface{}{
					"val1",
					"val2",
					"val3",
				},
				wantErr: false,
			},
			{
				name: "array-3",
				args: args{
					object: data,
					path:   "key3.array[2:3]",
				},
				want:    []interface{}{"val2"},
				wantErr: false,
			},
			{
				name: "array-4",
				args: args{
					object: data,
					path:   "key3.array[3:]",
				},
				want: []interface{}{
					"val3",
					"val4",
					"val5",
				},
				wantErr: false,
			},
			{
				name: "array-5",
				args: args{
					object: data,
					path:   "key3.array[:4]",
				},
				want: []interface{}{
					"val0",
					"val1",
					"val2",
					"val3",
				},
				wantErr: false,
			},
			{
				name: "array-6",
				args: args{
					object: data,
					path:   "key3.array[-2:]",
				},
				want: []interface{}{
					"val4",
					"val5",
				},
				wantErr: false,
			},
			{
				name: "array-7",
				args: args{
					object: data,
					path:   "key3.array[-6:5]",
				},
				want: []interface{}{
					"val0",
					"val1",
					"val2",
					"val3",
					"val4",
				},
				wantErr: false,
			},
			{
				name: "array-8",
				args: args{
					object: data,
					path:   "key3.array[1:-1]",
				},
				want: []interface{}{
					"val1",
					"val2",
					"val3",
					"val4",
				},
				wantErr: false,
			},
			{
				name: "array-9",
				args: args{
					object: data,
					path:   "key4[0:2].key1",
				},
				want: []interface{}{
					"val1",
					"val2",
				},
				wantErr: false,
			},
			{
				name: "array-10",
				args: args{
					object: data,
					path:   "key4[1:].key1",
				},
				want: []interface{}{
					"val2",
					"val3",
				},
				wantErr: false,
			},
			{
				name: "multi-select-1",
				args: args{
					object: data,
					path:   "key3.array[ 0, 1:4, 4:5 ]",
				},
				want: []interface{}{
					"val0",
					"val1",
					"val2",
					"val3",
					"val4",
				},
				wantErr: false,
			},
			{
				name: "multi-select-2",
				args: args{
					object: data,
					path:   "key3.array[ 1:3, 3:5 ]",
				},
				want: []interface{}{
					"val1",
					"val2",
					"val3",
					"val4",
				},
				wantErr: false,
			},
		},
		"wildcard": {
			{
				name: "array-1",
				args: args{
					object: data,
					path:   "key3.array.*",
				},
				want: []interface{}{
					"val0",
					"val1",
					"val2",
					"val3",
					"val4",
					"val5",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "array-2",
				args: args{
					object: data,
					path:   "key3.array[*]",
				},
				want: []interface{}{
					"val0",
					"val1",
					"val2",
					"val3",
					"val4",
					"val5",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "array-3",
				args: args{
					object: data,
					path:   "key3.array[ * ]",
				},
				want: []interface{}{
					"val0",
					"val1",
					"val2",
					"val3",
					"val4",
					"val5",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "array-4",
				args: args{
					object: data,
					path:   "key4.*.key1",
				},
				want: []interface{}{
					"val1",
					"val2",
					"val3",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "map-1",
				args: args{
					object: data,
					path:   "key3.map.*",
				},
				want: []interface{}{
					"val1",
					"val2",
					"val3",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "map-2",
				args: args{
					object: data,
					path:   "key3.map[*]",
				},
				want: []interface{}{
					"val1",
					"val2",
					"val3",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "map-3",
				args: args{
					object: data,
					path:   "key3.map[ * ]",
				},
				want: []interface{}{
					"val1",
					"val2",
					"val3",
				},
				wantErr:    false,
				sortResult: true,
			},
		},
		"get-object": {
			{
				name: "single-key",
				args: args{
					object: data,
					path:   "key1",
				},
				want: map[string]interface{}{
					"key2": map[string]interface{}{
						"key3": map[string]interface{}{
							"key4": map[string]interface{}{
								"key5": float64(123),
							},
						},
					},
				},
				wantErr: false,
			},
			{
				name: "nested-key",
				args: args{
					object: data,
					path:   "key1.key2.key3",
				},
				want: map[string]interface{}{
					"key4": map[string]interface{}{
						"key5": float64(123),
					},
				},
				wantErr: false,
			},
			{
				name: "array",
				args: args{
					object: data,
					path:   "key2.array",
				},
				want: []interface{}{
					map[string]interface{}{
						"subkey": "val",
					},
					float64(456),
					true,
				},
				wantErr: false,
			},
		},
		"key-formatting": {
			{
				name: "double-quotes-1",
				args: args{
					object: data,
					path:   "key5['\"double\"']",
				},
				want:    "double",
				wantErr: false,
			},
			{
				name: "double-quotes-2",
				args: args{
					object: data,
					path:   "key5[\"\\\"double\\\"\"]",
				},
				want:    "double",
				wantErr: false,
			},
			{
				name: "single-quotes-1",
				args: args{
					object: data,
					path:   "key5[\"'single'\"]",
				},
				want:    "single",
				wantErr: false,
			},
			{
				name: "single-quotes-2",
				args: args{
					object: data,
					path:   "key5['\\'single\\'']",
				},
				want:    "single",
				wantErr: false,
			},
			{
				name: "spaces-1",
				args: args{
					object: data,
					path:   "key5[\"  spaces  \"]",
				},
				want:    "spaces",
				wantErr: false,
			},
			{
				name: "spaces-2",
				args: args{
					object: data,
					path:   "key5['  spaces  ']",
				},
				want:    "spaces",
				wantErr: false,
			},
			{
				name: "quoted-special-characters-1",
				args: args{
					object: data,
					path:   "key5['][.,']",
				},
				want:    "specials",
				wantErr: false,
			},
			{
				name: "quoted-special-characters-2",
				args: args{
					object: data,
					path:   "key5[\"][.,\"]",
				},
				want:    "specials",
				wantErr: false,
			},
		},
		"recursive": {
			{
				name: "map-access-1",
				args: args{
					object: data,
					path:   "key6..recursive",
				},
				want: []interface{}{
					"val1",
					"val2",
					"val3",
					"val4",
					"val5",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "map-access-2",
				args: args{
					object: data,
					path:   "key6['key7'].key9..recursive",
				},
				want: []interface{}{
					"val4",
					"val5",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "map-access-3",
				args: args{
					object: data,
					path:   "key6..key9[0,1].recursive",
				},
				want: []interface{}{
					"val4",
					"val5",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "map-access-4",
				args: args{
					object: data,
					path:   "key2..subkey",
				},
				want: []interface{}{
					"val",
				},
				wantErr: false,
			},
			{
				name: "map-access-5",
				args: args{
					object: data,
					path:   "key7..recursive",
				},
				wantJson: "[true,{\"recursive\":true},[{\"recursive\":{\"recursive\":true}}]]",
				wantErr:  false,
			},
			{
				name: "slice-access-5",
				args: args{
					object: data,
					path:   "key7.arrays..[0]",
				},
				want: []interface{}{
					"val1",
					"val3",
					"val5",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "slice-access-2",
				args: args{
					object: data,
					path:   "key7.arrays..[1]",
				},
				want: []interface{}{
					"val2",
					"val4",
					"val6",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "slice-access-3",
				args: args{
					object: data,
					path:   "key7.arrays..[0,1]",
				},
				want: []interface{}{
					"val1",
					"val2",
					"val3",
					"val4",
					"val5",
					"val6",
				},
				wantErr:    false,
				sortResult: true,
			},
		},
		"errors": {
			{
				name: "missing-key-1",
				args: args{
					object: data,
					path:   "none",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "key does not exist",
			},
			{
				name: "missing-key-2",
				args: args{
					object: data,
					path:   "key1.key2.key3.key99",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "key does not exist",
			},
			{
				name: "missing-key-3",
				args: args{
					object: data,
					path:   "key2.array[0].missing",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "key does not exist",
			},
			{
				name: "missing-index-1",
				args: args{
					object: data,
					path:   "key2.array[3]",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "index out of range",
			},
			{
				name: "missing-index-2",
				args: args{
					object: data,
					path:   "key2.array[3].missing",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "index out of range",
			},
			{
				name: "missing-index-3",
				args: args{
					object: data,
					path:   "key3.array[-10]",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "index out of range",
			},
			{
				name: "missing-index-4",
				args: args{
					object: data,
					path:   "key3.array[0:10]",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "index out of range",
			},
			{
				name: "incorrect-access-type-1",
				args: args{
					object: data,
					path:   "key3.map[0]",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "cannot access map with an index",
			},
			{
				name: "incorrect-access-type-2",
				args: args{
					object: data,
					path:   "key3.array.key",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "cannot access array with a key",
			},
			{
				name: "path-not-found-1",
				args: args{
					object: data,
					path:   "key1.key2.key3.key4.key5.key6",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "path not found",
			},
			{
				name: "path-not-found-2",
				args: args{
					object: data,
					path:   "key1..missing",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "path not found",
			},
			{
				name: "path-not-found-3",
				args: args{
					object: data,
					path:   "key6..recursive.missing",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "path not found",
			},
			{
				name: "path-not-found-4",
				args: args{
					object: data,
					path:   "..missing",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "path not found",
			},
		},
		"multi-method": {
			{
				name: "mapp-and-index",
				args: args{
					object: data,
					path:   "$.key3['array'][2]",
				},
				want:    "val2",
				wantErr: false,
			},
			{
				name: "wildcard-and-recursive",
				args: args{
					object: data,
					path:   "key6[key7].*..recursive",
				},
				want: []interface{}{
					"val3",
					"val4",
					"val5",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "recursive-and-index",
				args: args{
					object: data,
					path:   "key6..key9[0].recursive",
				},
				want: []interface{}{
					"val4",
				},
				wantErr: false,
			},
			{
				name: "recursive-and-index-range",
				args: args{
					object: data,
					path:   "key6..key9[0:2].recursive",
				},
				want: []interface{}{
					"val4",
					"val5",
				},
				wantErr:    false,
				sortResult: true,
			},
			{
				name: "recursive-and-mulit-access",
				args: args{
					object: data,
					path:   "key6.key7['key8','key9']..recursive",
				},
				want: []interface{}{
					"val3",
					"val4",
					"val5",
				},
				wantErr:    false,
				sortResult: true,
			},
		},
		"other": {
			{
				name: "null-value",
				args: args{
					object: data,
					path:   "$.key5.null_value",
				},
				want:    nil,
				wantErr: false,
			},
			{
				name: "empty-slice",
				args: args{
					object: data,
					path:   "$.key5.empty_slice",
				},
				want:    []interface{}{},
				wantErr: false,
			},
			{
				name: "empty-map",
				args: args{
					object: data,
					path:   "$.key5.empty_map",
				},
				want:    map[string]interface{}{},
				wantErr: false,
			},
			{
				name: "int",
				args: args{
					object: data,
					path:   "$.key5.int",
				},
				want:    float64(123),
				wantErr: false,
			},
			{
				name: "float",
				args: args{
					object: data,
					path:   "$.key5.float",
				},
				want:    float64(1.23),
				wantErr: false,
			},
		},
	}
	for groupName, group := range tests {
		for _, tt := range group {
			testName := fmt.Sprintf("%s-%s", groupName, tt.name)
			if runTest != "" && testName != runTest {
				continue
			}
			t.Run(testName, func(t *testing.T) {
				got, err := Get(tt.args.object, tt.args.path)
				if (err != nil) != tt.wantErr {
					t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErr {
					if err.(*Error).Code != tt.wantErrCode {
						t.Errorf("Get() errCode = %v, wantCode %v", err.(*Error).Code, tt.wantErrCode)
					}
					if !strings.Contains(err.Error(), tt.wantErrMsg) {
						t.Errorf("Get() errMsg = %v, wantMsg %v", err.(*Error).Msg, tt.wantErrMsg)
					}
					return
				}
				if tt.wantJson != "" {
					resp, err := json.Marshal(got)
					if err != nil {
						t.Errorf("Get() error = %v", err)
					}
					if string(resp) != tt.wantJson {
						t.Errorf("Get() = %v, want %v", string(resp), tt.wantJson)
					}
				}
				if tt.want != nil {
					if tt.sortResult {
						sort.Slice(got, func(i, j int) bool {
							return got.([]interface{})[i].(string) < got.([]interface{})[j].(string)
						})
					}
					if !reflect.DeepEqual(got, tt.want) {
						t.Errorf("Get() = %v, want %v", got, tt.want)
					}
				}
			})
		}
	}
}

func TestSet(t *testing.T) {
	type args struct {
		object interface{}
		path   string
		value  interface{}
	}

	tests := map[string][]struct {
		name        string
		args        args
		want        interface{}
		wantErr     bool
		wantErrCode string
		wantErrMsg  string
		strictMode  bool
	}{
		"map-set": {
			{
				name: "dot-notation-1",
				args: args{
					object: map[string]interface{}{},
					path:   "key1",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": "val",
				},
				wantErr: false,
			},
			{
				name: "dot-notation-2",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2.key3",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": map[string]interface{}{
							"key3": "val",
						},
					},
				},
				wantErr: false,
			},
			{
				name: "bracket-notation",
				args: args{
					object: map[string]interface{}{},
					path:   "[key1][key2][key3]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": map[string]interface{}{
							"key3": "val",
						},
					},
				},
				wantErr: false,
			},
			{
				name: "mixed-notation",
				args: args{
					object: map[string]interface{}{},
					path:   "key1[key2].key3",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": map[string]interface{}{
							"key3": "val",
						},
					},
				},
				wantErr: false,
			},
			{
				name: "mixed-notation-quotes",
				args: args{
					object: map[string]interface{}{},
					path:   "key1['key2'][\"key3\"]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": map[string]interface{}{
							"key3": "val",
						},
					},
				},
				wantErr: false,
			},
		},
		"array-set": {
			{
				name: "single-index-1",
				args: args{
					object: map[string]interface{}{},
					path:   "key1[0]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": []interface{}{
						"val",
					},
				},
				wantErr: false,
			},
			{
				name: "single-index-2",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2[0]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": []interface{}{
							"val",
						},
					},
				},
				wantErr: false,
			},
			{
				name: "single-index-3",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2[0].key3",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": []interface{}{
							map[string]interface{}{
								"key3": "val",
							},
						},
					},
				},
				wantErr: false,
			},
			{
				name: "add-slice-range-1",
				args: args{
					object: getData(),
					path:   "key3.array[6]",
					value:  "val",
				},
				want: func() interface{} {
					expected := getData()
					slice := expected.(map[string]interface{})["key3"].(map[string]interface{})["array"].([]interface{})
					slice = fillSlice(slice, 6)
					slice[6] = "val"
					expected.(map[string]interface{})["key3"].(map[string]interface{})["array"] = slice
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "add-slice-range-2",
				args: args{
					object: getData(),
					path:   "key3.array[10]",
					value:  "val",
				},
				want: func() interface{} {
					expected := getData()
					slice := expected.(map[string]interface{})["key3"].(map[string]interface{})["array"].([]interface{})
					slice = fillSlice(slice, 10)
					slice[10] = "val"
					expected.(map[string]interface{})["key3"].(map[string]interface{})["array"] = slice
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "add-slice-range-3",
				args: args{
					object: getData(),
					path:   "key3.array[0:8]",
					value:  "val",
				},
				want: func() interface{} {
					expected := getData()
					slice := expected.(map[string]interface{})["key3"].(map[string]interface{})["array"].([]interface{})
					slice = fillSlice(slice, 7)
					for i := range slice {
						slice[i] = "val"
					}
					expected.(map[string]interface{})["key3"].(map[string]interface{})["array"] = slice
					return expected
				}(),
				wantErr: false,
			},
		},
		"object-set": {
			{
				name: "dot-notation-1",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2.key3",
					value: map[string]interface{}{
						"key4": "val",
					},
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": map[string]interface{}{
							"key3": map[string]interface{}{
								"key4": "val",
							},
						},
					},
				},
				wantErr: false,
			},
			{
				name: "dot-notation-2",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2",
					value: map[string]interface{}{
						"key3": map[string]interface{}{
							"key4": "val",
						},
					},
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": map[string]interface{}{
							"key3": map[string]interface{}{
								"key4": "val",
							},
						},
					},
				},
				wantErr: false,
			},
			{
				name: "dot-notation-3",
				args: args{
					object: map[string]interface{}{},
					path:   "key1[0]",
					value: map[string]interface{}{
						"key3": map[string]interface{}{
							"key4": "val",
						},
					},
				},
				want: map[string]interface{}{
					"key1": []interface{}{
						map[string]interface{}{
							"key3": map[string]interface{}{
								"key4": "val",
							},
						},
					},
				},
				wantErr: false,
			},
		},
		"update": {
			{
				name: "map-1",
				args: args{
					object: getData(),
					path:   "key1.key2.key3.key4.key5",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key1"].(map[string]interface{})["key2"].(map[string]interface{})["key3"].(map[string]interface{})["key4"].(map[string]interface{})["key5"] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "array-1",
				args: args{
					object: getData(),
					path:   "key2.array[0]",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key2"].(map[string]interface{})["array"].([]interface{})[0] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "array-2",
				args: args{
					object: getData(),
					path:   "key2.array[0].subkey",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key2"].(map[string]interface{})["array"].([]interface{})[0].(map[string]interface{})["subkey"] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "array-3",
				args: args{
					object: getData(),
					path:   "key2.array[-1]",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key2"].(map[string]interface{})["array"].([]interface{})[2] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "strict-map-1",
				args: args{
					object: getData(),
					path:   "key1.key2.key3.key4.key5",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key1"].(map[string]interface{})["key2"].(map[string]interface{})["key3"].(map[string]interface{})["key4"].(map[string]interface{})["key5"] = "test"
					return expected
				}(),
				wantErr:    false,
				strictMode: true,
			},
			{
				name: "strict-array-1",
				args: args{
					object: getData(),
					path:   "key2.array[0]",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key2"].(map[string]interface{})["array"].([]interface{})[0] = "test"
					return expected
				}(),
				wantErr:    false,
				strictMode: true,
			},
			{
				name: "strict-array-2",
				args: args{
					object: getData(),
					path:   "key2.array[0].subkey",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key2"].(map[string]interface{})["array"].([]interface{})[0].(map[string]interface{})["subkey"] = "test"
					return expected
				}(),
				wantErr:    false,
				strictMode: true,
			},
			{
				name: "strict-array-3",
				args: args{
					object: getData(),
					path:   "key2.array[-1]",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key2"].(map[string]interface{})["array"].([]interface{})[2] = "test"
					return expected
				}(),
				wantErr:    false,
				strictMode: true,
			},
		},
		"multi-set": {
			{
				name: "map-1",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2[key3,key4,key5]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": map[string]interface{}{
							"key3": "val",
							"key4": "val",
							"key5": "val",
						},
					},
				},
				wantErr: false,
			},
			{
				name: "map-2",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2[ key3, key4, key5 ]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": map[string]interface{}{
							"key3": "val",
							"key4": "val",
							"key5": "val",
						},
					},
				},
				wantErr: false,
			},
			{
				name: "map-3",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2[key3,'key4',\"key5\"]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": map[string]interface{}{
							"key3": "val",
							"key4": "val",
							"key5": "val",
						},
					},
				},
				wantErr: false,
			},
			{
				name: "map-4",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2[ key3, 'key4', \"key5\" ]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": map[string]interface{}{
							"key3": "val",
							"key4": "val",
							"key5": "val",
						},
					},
				},
				wantErr: false,
			},
			{
				name: "array-1",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2[0,1,2]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": []interface{}{
							"val",
							"val",
							"val",
						},
					},
				},
				wantErr: false,
			},
			{
				name: "array-2",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2[ 0, 1, 2 ]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": []interface{}{
							"val",
							"val",
							"val",
						},
					},
				},
				wantErr: false,
			},
		},
		"range-set": {
			{
				name: "array-1",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2[0:2]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": []interface{}{
							"val",
							"val",
						},
					},
				},
				wantErr: false,
			},
			{
				name: "array-2",
				args: args{
					object: map[string]interface{}{},
					path:   "key1.key2[0:4]",
					value:  "val",
				},
				want: map[string]interface{}{
					"key1": map[string]interface{}{
						"key2": []interface{}{
							"val",
							"val",
							"val",
							"val",
						},
					},
				},
				wantErr: false,
			},
		},
		"wildcard-update": {
			{
				name: "array",
				args: args{
					object: getData(),
					path:   "key3.array.*",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key3"].(map[string]interface{})["array"] = []interface{}{
						"test", "test", "test", "test", "test", "test",
					}
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "map",
				args: args{
					object: getData(),
					path:   "key3.map.*",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key3"].(map[string]interface{})["map"] = map[string]interface{}{
						"key1": "test",
						"key2": "test",
						"key3": "test",
					}
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "nested-objects",
				args: args{
					object: getData(),
					path:   "key4.*.key1",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key4"].([]interface{})[0].(map[string]interface{})["key1"] = "test"
					expected.(map[string]interface{})["key4"].([]interface{})[1].(map[string]interface{})["key1"] = "test"
					expected.(map[string]interface{})["key4"].([]interface{})[2].(map[string]interface{})["key1"] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "strict-array",
				args: args{
					object: getData(),
					path:   "key3.array.*",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key3"].(map[string]interface{})["array"] = []interface{}{
						"test", "test", "test", "test", "test", "test",
					}
					return expected
				}(),
				wantErr:    false,
				strictMode: true,
			},
			{
				name: "strict-map",
				args: args{
					object: getData(),
					path:   "key3.map.*",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key3"].(map[string]interface{})["map"] = map[string]interface{}{
						"key1": "test",
						"key2": "test",
						"key3": "test",
					}
					return expected
				}(),
				wantErr:    false,
				strictMode: true,
			},
			{
				name: "strict-nested-objects",
				args: args{
					object: getData(),
					path:   "key4.*.key1",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key4"].([]interface{})[0].(map[string]interface{})["key1"] = "test"
					expected.(map[string]interface{})["key4"].([]interface{})[1].(map[string]interface{})["key1"] = "test"
					expected.(map[string]interface{})["key4"].([]interface{})[2].(map[string]interface{})["key1"] = "test"
					return expected
				}(),
				wantErr:    false,
				strictMode: true,
			},
		},
		"errors": {
			{
				name: "incorrect-access-type-1",
				args: args{
					object: getData(),
					path:   "key3.map[0]",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "cannot set map with an index",
			},
			{
				name: "incorrect-access-type-2",
				args: args{
					object: getData(),
					path:   "key3.map[0,1,2]",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "cannot set map with an index",
			},
			{
				name: "incorrect-access-type-3",
				args: args{
					object: getData(),
					path:   "key3.map[0:2]",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "cannot set map with an index",
			},
			{
				name: "incorrect-access-type-4",
				args: args{
					object: getData(),
					path:   "key3.array.key",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "cannot set array with a key",
			},
			{
				name: "incorrect-access-type-5",
				args: args{
					object: getData(),
					path:   "key3.array['key']",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "cannot set array with a key",
			},
			{
				name: "incorrect-access-type-6",
				args: args{
					object: getData(),
					path:   "key3.array['key1', 'key2']",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "cannot set array with a key",
			},
			{
				name: "incorrect-access-type-7",
				args: args{
					object: getData(),
					path:   "key1.key2.key3.key4.key5.*",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "path not found",
			},
			{
				name: "incorrect-access-type-8",
				args: args{
					object: getData(),
					path:   "key1.key2.key3.key4.key5.*",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "path not found",
			},
			{
				name: "incorrect-access-type-9",
				args: args{
					object: getData(),
					path:   "key1.key2.key3.key4.key5.*",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "path not found",
			},
			{
				name: "incorrect-access-type-10",
				args: args{
					object: getData(),
					path:   "key1.key2.key3.key4.key5..recursive",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "path not found",
			},
			{
				name: "strict-index-out-of-range-1",
				args: args{
					object: getData(),
					path:   "key3.array[6]",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "index out of range",
				strictMode:  true,
			},
			{
				name: "strict-index-out-of-range-2",
				args: args{
					object: getData(),
					path:   "key3.array[0,1,2,3,4,5,6]",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "index out of range",
				strictMode:  true,
			},
			{
				name: "strict-index-out-of-range-3",
				args: args{
					object: getData(),
					path:   "key3.array[0:10]",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "index out of range",
				strictMode:  true,
			},
			{
				name: "strict-key-not-found-1",
				args: args{
					object: getData(),
					path:   ".missing",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "key does not exist",
				strictMode:  true,
			},
			{
				name: "strict-key-not-found-2",
				args: args{
					object: getData(),
					path:   "key1.key2.missing",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "key does not exist",
				strictMode:  true,
			},
			{
				name: "strict-key-not-found-3",
				args: args{
					object: getData(),
					path:   "key1.key2.key3.key4.key6",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "key does not exist",
				strictMode:  true,
			},
			{
				name: "strict-key-not-found-4",
				args: args{
					object: getData(),
					path:   "key1.key2.key3.key4.key5.key6",
					value:  "test",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "path not found",
				strictMode:  true,
			},
		},
		"recursive-set": {
			{
				name: "map-1",
				args: args{
					object: getData(),
					path:   "key6..recursive",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key6"].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key8"].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[0].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[1].(map[string]interface{})["recursive"] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "map-2",
				args: args{
					object: getData(),
					path:   "key6['key7'].key9..recursive",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[0].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[1].(map[string]interface{})["recursive"] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "strict-map-1",
				args: args{
					object: getData(),
					path:   "key6..recursive",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key6"].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key8"].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[0].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[1].(map[string]interface{})["recursive"] = "test"
					return expected
				}(),
				wantErr:    false,
				strictMode: true,
			},
			{
				name: "strict-map-2",
				args: args{
					object: getData(),
					path:   "key6['key7'].key9..recursive",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[0].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[1].(map[string]interface{})["recursive"] = "test"
					return expected
				}(),
				wantErr:    false,
				strictMode: true,
			},
			{
				name: "map-3",
				args: args{
					object: getData(),
					path:   "key6..key9[0,1].recursive",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[0].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[1].(map[string]interface{})["recursive"] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "map-4",
				args: args{
					object: getData(),
					path:   "key7..recursive",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key7"].(map[string]interface{})["recursive"] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "map-5",
				args: args{
					object: getData(),
					path:   "key7..recursive[0].recursive",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key7"].(map[string]interface{})["recursive"].([]interface{})[0].(map[string]interface{})["recursive"] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "slice-1",
				args: args{
					object: getData(),
					path:   "key7.arrays..[0]",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["a"].([]interface{})[0] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["b"].([]interface{})[0] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["c"].([]interface{})[0] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "slice-2",
				args: args{
					object: getData(),
					path:   "key7.arrays..[1]",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["a"].([]interface{})[1] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["b"].([]interface{})[1] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["c"].([]interface{})[1] = "test"
					return expected
				}(),
				wantErr: false,
			},
			{
				name: "strict-slice-1",
				args: args{
					object: getData(),
					path:   "key7.arrays..[0]",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["a"].([]interface{})[0] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["b"].([]interface{})[0] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["c"].([]interface{})[0] = "test"
					return expected
				}(),
				wantErr:    false,
				strictMode: true,
			},
			{
				name: "strict-slice-2",
				args: args{
					object: getData(),
					path:   "key7.arrays..[1]",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["a"].([]interface{})[1] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["b"].([]interface{})[1] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["c"].([]interface{})[1] = "test"
					return expected
				}(),
				wantErr:    false,
				strictMode: true,
			},
			{
				name: "slice-3",
				args: args{
					object: getData(),
					path:   "key7.arrays..[0,1]",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["a"].([]interface{})[0] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["a"].([]interface{})[1] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["b"].([]interface{})[0] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["b"].([]interface{})[1] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["c"].([]interface{})[0] = "test"
					expected.(map[string]interface{})["key7"].(map[string]interface{})["arrays"].(map[string]interface{})["c"].([]interface{})[1] = "test"
					return expected
				}(),
				wantErr: false,
			},
		},
		"multi-method": {
			{
				name: "wildcard-and-recursive",
				args: args{
					object: getData(),
					path:   "key6[key7].*..recursive",
					value:  "test",
				},
				want: func() interface{} {
					expected := getData()
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key8"].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[0].(map[string]interface{})["recursive"] = "test"
					expected.(map[string]interface{})["key6"].(map[string]interface{})["key7"].(map[string]interface{})["key9"].([]interface{})[1].(map[string]interface{})["recursive"] = "test"
					return expected
				}(),
				wantErr: false,
			},
		},
	}

	for groupName, group := range tests {
		for _, tt := range group {
			testName := fmt.Sprintf("%s-%s", groupName, tt.name)
			if runTest != "" && testName != runTest {
				continue
			}
			t.Run(testName, func(t *testing.T) {
				c, err := Compile(tt.args.path)
				if err != nil {
					t.Errorf("Compile() error = %v", err)
					return
				}
				if tt.strictMode {
					c.SetStrict()
				}
				err = c.Set(tt.args.object, tt.args.value)
				if (err != nil) != tt.wantErr {
					t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErr {
					if err.(*Error).Code != tt.wantErrCode {
						t.Errorf("Get() errCode = %v, wantCode %v", err.(*Error).Code, tt.wantErrCode)
					}
					if !strings.Contains(err.Error(), tt.wantErrMsg) {
						t.Errorf("Get() errMsg = %v, wantMsg %v", err.(*Error).Msg, tt.wantErrMsg)
					}
					return
				}
				if !reflect.DeepEqual(tt.args.object, tt.want) {
					t.Errorf("data = %v, want %v", tt.args.object, tt.want)
				}
			})
		}
	}
}
