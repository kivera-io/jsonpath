# JSON Path

Get and set values in a JSON object using jsonpath format.

This library allows you to retrieve and update any value in a JSON string using jsonpath format. The value can be of any type, concrete or not.

## Path Operators
The following operators are available. All operators can be used to both set and get values.

| Syntax | Description |
| :------------: | :------------: |
| `$.` | Root element. Can be ommitted. |
| `.key` | Dot notation. Access a single key with a parent object. |
| `[ key (, key) ]` | Bracket notation. Access one or more keys within a parent object. Single quoted ('key') and double quoted ("key") strings can also be used within square brackets to access keys with special characters. |
| `[ n (, n) ]` | Access one or more indices in a parent array. Negative indices are allowed. |
| `[ start:end ]` | Access a range of indicies in a parent array from the start index, up to but not including the end index. This notation can also be used alongside single index access. |
| `[ n: ]` | Access a range of indicies in a parent array from the start index until the end of the array. |
| `[ :n ]` | Access a range of indicies in a parent array from the start of the array, up to but not including the end index. |
| `.*` *or* `[*]` | Access all elements in the parent object/array. |

## Examples

| Path  | Descripton  |
| :------------ | :------------ |
| `key1.key2.key3`  | Dot notation  |
| `[key1][key2][key3]`  | Bracket notation  |
| `key1[key2].key3`   | Combination of both dot and bracket notation |
| `map['Key with spaces']`   | Access a map key with special characters  |
| `array[0]`  | Access first element of array  |
| `array[-1]`  | Access last element of array  |
| `array[1:3]`  | Access the second and third element of array  |
| `array[:3]`  | Access the first three elements of array  |
| `array[-3:]`  | Access the last three elements of array  |
| `array[ 0, -1, 2:5 ]`  |  Access first, last, and third until fifth element of array   |
| `map[ key1, key2 ]`  | Access key1 and key2 in map  |
| `map[ key1, key2 ].property`   | Access a property from key1 and key2 in map |
| `array[*]`  | Access all elements of array  |
| `map.*`  | Access all items in map  |
| `map[*].property`  | Access a property from all items in map  |

## In Code

First, unmarshal your json string into an interface. Then call jsonpath.Set() and jsonpath.Get() to access and manipulate the data.

```
import "github.com/kivera-io/jsonpath"

example := "{}"
var data interface{}

err := json.Unmarshal([]byte(example), &data)
if err != nil {
    panic(err)
}

// set value at a path
err = jsonpath.Set(data, "test.path", "value")
if err != nil {
    panic(err)
}

// get value at a path
val, err := jsonpath.Get(data, "test.path")
if err != nil {
    panic(err)
}
fmt.Println(val)
```
## Error Handling

There are two types of errors that can be thrown. Ether  `InvalidPath` or `NotFound`.

`InvalidPath` is thrown when a path with invalid syntax has been provided.

`NotFound` indicates that the path has valid syntax, but it does not exist in, or is not valid with, the provided data.

To differentiate between the different errors.

```
if err.(*jsonpath.Error).Code == jsonpath.NotFound {
   do something...
}
```