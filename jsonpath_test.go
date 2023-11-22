package jsonpath

import (
	"encoding/json"
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
		"  spaces  ": "spaces"
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
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "map-access-dot-notation",
			args: args{
				object: data,
				path:   "key1.key2.key3.key4.key5",
			},
			want:    float64(123),
			wantErr: false,
		},
		{
			name: "map-bracket-notation",
			args: args{
				object: data,
				path:   "[key1][key2][key3][key4][key5]",
			},
			want:    float64(123),
			wantErr: false,
		},
		{
			name: "map-mixed-notation",
			args: args{
				object: data,
				path:   "key1[key2].key3[key4][key5]",
			},
			want:    float64(123),
			wantErr: false,
		},
		{
			name: "map-mixed-notation-quotes",
			args: args{
				object: data,
				path:   "key1['key2'].key3[\"key4\"][key5]",
			},
			want:    float64(123),
			wantErr: false,
		},
		{
			name: "array-access-1",
			args: args{
				object: data,
				path:   "key2.array[0].subkey",
			},
			want:    "val",
			wantErr: false,
		},
		{
			name: "array-access-2",
			args: args{
				object: data,
				path:   "key2.array[0][subkey]",
			},
			want:    "val",
			wantErr: false,
		},
		{
			name: "array-access-3",
			args: args{
				object: data,
				path:   "key2.array[0]['subkey']",
			},
			want:    "val",
			wantErr: false,
		},
		{
			name: "array-access-4",
			args: args{
				object: data,
				path:   "key2.array[1]",
			},
			want:    float64(456),
			wantErr: false,
		},
		{
			name: "array-access-5",
			args: args{
				object: data,
				path:   "key2.array[2]",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "array-access-6",
			args: args{
				object: data,
				path:   "key2.array[-1]",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "array-access-7",
			args: args{
				object: data,
				path:   "key2.array[-2]",
			},
			want:    float64(456),
			wantErr: false,
		},
		{
			name: "array-access-8",
			args: args{
				object: data,
				path:   "key2.array[-3].subkey",
			},
			want:    "val",
			wantErr: false,
		},
		{
			name: "multi-select-array-1",
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
			name: "multi-select-array-2",
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
			name: "multi-select-array-3",
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
			name: "range-select-array-1",
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
			name: "range-select-array-2",
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
			name: "range-select-array-3",
			args: args{
				object: data,
				path:   "key3.array[2:3]",
			},
			want:    "val2",
			wantErr: false,
		},
		{
			name: "range-select-array-4",
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
			name: "range-select-array-5",
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
			name: "range-select-array-6",
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
			name: "range-select-array-7",
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
			name: "range-select-array-8",
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
			name: "range-select-array-9",
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
			name: "range-select-array-10",
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
			name: "multi-select-range-select-array-1",
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
			name: "multi-select-range-select-array-2",
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
		{
			name: "multi-select-map-1",
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
			name: "multi-select-map-2",
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
			name: "multi-select-map-3",
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
			name: "multi-select-map-4",
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
		{
			name: "wildcard-select-array-1",
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
			name: "wildcard-select-array-2",
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
			name: "wildcard-select-array-3",
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
			name: "wildcard-select-map-1",
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
			name: "wildcard-select-map-2",
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
			name: "wildcard-select-map-3",
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
		{
			name: "get-object-1",
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
			name: "get-object-2",
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
			name: "get-object-3",
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
		{
			name: "key-with-double-quotes-1",
			args: args{
				object: data,
				path:   "key5['\"double\"']",
			},
			want:    "double",
			wantErr: false,
		},
		{
			name: "key-with-double-quotes-2",
			args: args{
				object: data,
				path:   "key5[\"\\\"double\\\"\"]",
			},
			want:    "double",
			wantErr: false,
		},
		{
			name: "key-with-single-quotes-1",
			args: args{
				object: data,
				path:   "key5[\"'single'\"]",
			},
			want:    "single",
			wantErr: false,
		},
		{
			name: "key-with-single-quotes-2",
			args: args{
				object: data,
				path:   "key5['\\'single\\'']",
			},
			want:    "single",
			wantErr: false,
		},
		{
			name: "key-with-spaces-1",
			args: args{
				object: data,
				path:   "key5[\"  spaces  \"]",
			},
			want:    "spaces",
			wantErr: false,
		},
		{
			name: "key-with-spaces-2",
			args: args{
				object: data,
				path:   "key5['  spaces  ']",
			},
			want:    "spaces",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.object, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if strings.HasPrefix(tt.name, "wildcard-select-map-") {
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
