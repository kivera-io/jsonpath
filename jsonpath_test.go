package jsonpath

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

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
		"[.]": "specials"
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
		wantErr     bool
		wantErrCode string
		wantErrMsg  string
	}{
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
				name: "negative-index-1",
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
				want:    "val2",
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
				wantErr: false,
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
				wantErr: false,
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
				wantErr: false,
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
				wantErr: false,
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
				wantErr: false,
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
				wantErr: false,
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
					path:   "key5['[.]']",
				},
				want:    "specials",
				wantErr: false,
			},
			{
				name: "quoted-special-characters-2",
				args: args{
					object: data,
					path:   "key5[\"[.]\"]",
				},
				want:    "specials",
				wantErr: false,
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
				name: "path-not-found",
				args: args{
					object: data,
					path:   "key1.key2.key3.key4.key5.key6",
				},
				wantErr:     true,
				wantErrCode: NotFound,
				wantErrMsg:  "path not found",
			},
			{
				name: "invalid-path-1",
				args: args{
					object: data,
					path:   "key1..key2",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "empty path segment",
			},
			{
				name: "invalid-path-2",
				args: args{
					object: data,
					path:   "key1.key2[]",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "empty path segment",
			},
			{
				name: "invalid-path-3",
				args: args{
					object: data,
					path:   "key1.key2['test'",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "missing closing bracket",
			},
			{
				name: "invalid-path-4",
				args: args{
					object: data,
					path:   "key1.key2[0]]",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "missing opening bracket",
			},
			{
				name: "invalid-path-5",
				args: args{
					object: data,
					path:   "key1.key2['test]",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "missing closing quote",
			},
			{
				name: "invalid-path-6",
				args: args{
					object: data,
					path:   "key1.key2['test'][']",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "missing closing quote",
			},
			{
				name: "invalid-path-7",
				args: args{
					object: data,
					path:   "key1.key2['test\"][\\'\"]",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "missing closing quote",
			},
			{
				name: "invalid-path-8",
				args: args{
					object: data,
					path:   "key1.key2.'test'",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "cannot use quotes outside of brackets",
			},
			{
				name: "invalid-path-9",
				args: args{
					object: data,
					path:   "key1.key2.key3.",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "path cannot end with '.' separator",
			},
			{
				name: "invalid-path-10",
				args: args{
					object: data,
					path:   "key1.key2.key3[0].",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "path cannot end with '.' separator",
			},
			{
				name: "invalid-path-11",
				args: args{
					object: data,
					path:   "key1.  key2",
				},
				wantErr:     true,
				wantErrCode: InvalidPath,
				wantErrMsg:  "cannot use whitespace characters outside quotes and brackets",
			},
		},
	}
	for groupName, group := range tests {
		for _, tt := range group {
			testName := fmt.Sprintf("%s-%s", groupName, tt.name)
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
				if strings.HasPrefix(testName, "wildcard-map-") {
					sort.Slice(got, func(i, j int) bool {
						return got.([]interface{})[i].(string) < got.([]interface{})[j].(string)
					})
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Get() = %v, want %v", got, tt.want)
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
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "map-set-dot-notation-1",
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
			name: "map-set-dot-notation-2",
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
			name: "map-set-bracket-notation",
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
			name: "map-set-mixed-notation",
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
			name: "map-set-mixed-notation-quotes",
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
		{
			name: "array-set-1",
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
			name: "array-set-2",
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
			name: "array-set-3",
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
			name: "set-object-1",
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
			name: "set-object-2",
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
			name: "set-object-3",
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
		{
			name: "update-map-1",
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
			name: "update-array-1",
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
			name: "update-array-2",
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
			name: "update-array-3",
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
			name: "multi-set-map-1",
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
			name: "multi-set-map-2",
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
			name: "multi-set-map-3",
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
			name: "multi-set-map-4",
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
			name: "multi-set-array-1",
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
			name: "multi-set-array-2",
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
		{
			name: "range-set-array-1",
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
			name: "range-set-array-2",
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
		{
			name: "update-wildcard-1",
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
			name: "update-wildcard-2",
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
			name: "update-wildcard-3",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Set(tt.args.object, tt.args.path, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.object, tt.want) {
				t.Errorf("data = %v, want %v", tt.args.object, tt.want)
			}
		})
	}
}
