package jsonpath

import (
	"encoding/json"
	"reflect"
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
	"key6": {
		"array": [
			{
				"subkey": "val"
			},
			456,
			true
		]
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
				path:   "key6.array[0].subkey",
			},
			want:    "val",
			wantErr: false,
		},
		{
			name: "array-access-2",
			args: args{
				object: data,
				path:   "key6.array[0][subkey]",
			},
			want:    "val",
			wantErr: false,
		},
		{
			name: "array-access-3",
			args: args{
				object: data,
				path:   "key6.array[0]['subkey']",
			},
			want:    "val",
			wantErr: false,
		},
		{
			name: "array-access-4",
			args: args{
				object: data,
				path:   "key6.array[1]",
			},
			want:    float64(456),
			wantErr: false,
		},
		{
			name: "array-access-5",
			args: args{
				object: data,
				path:   "key6.array[2]",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "array-access-6",
			args: args{
				object: data,
				path:   "key6.array[-1]",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "array-access-7",
			args: args{
				object: data,
				path:   "key6.array[-2]",
			},
			want:    float64(456),
			wantErr: false,
		},
		{
			name: "array-access-8",
			args: args{
				object: data,
				path:   "key6.array[-3].subkey",
			},
			want:    "val",
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
				path:   "key6.array",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.object, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet(t *testing.T) {
	udpateMap1Expected := getData()
	udpateMap1Expected.(map[string]interface{})["key1"].(map[string]interface{})["key2"].(map[string]interface{})["key3"].(map[string]interface{})["key4"].(map[string]interface{})["key5"] = "test"
	udpateMap2Expected := getData()
	udpateMap2Expected.(map[string]interface{})["key6"].(map[string]interface{})["array"].([]interface{})[0] = "test"
	udpateMap3Expected := getData()
	udpateMap3Expected.(map[string]interface{})["key6"].(map[string]interface{})["array"].([]interface{})[0].(map[string]interface{})["subkey"] = "test"
	udpateMap4Expected := getData()
	udpateMap4Expected.(map[string]interface{})["key6"].(map[string]interface{})["array"].([]interface{})[2] = "test"

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
			name: "map-set-bracket-notation-2",
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
			want:    udpateMap1Expected,
			wantErr: false,
		},
		{
			name: "update-map-2",
			args: args{
				object: getData(),
				path:   "key6.array[0]",
				value:  "test",
			},
			want:    udpateMap2Expected,
			wantErr: false,
		},
		{
			name: "update-map-3",
			args: args{
				object: getData(),
				path:   "key6.array[0].subkey",
				value:  "test",
			},
			want:    udpateMap3Expected,
			wantErr: false,
		},
		{
			name: "update-map-4",
			args: args{
				object: getData(),
				path:   "key6.array[-1]",
				value:  "test",
			},
			want:    udpateMap4Expected,
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
