# JSON Path

Get and set values in an unmarshalled json using json path format.

## Examples

First, unmarshal your json string into an interface.

```
var data interface{}

err := json.Unmarshal([]byte(example), &data)
if err != nil {
    panic(err)
}
```

## Set Values

Set values in your data using the following.

```
// set a path
err = SetJsonPath(data, "test.path", "value")
if err != nil {
    fmt.Println(err)
}

// set an array index
err = SetJsonPath(data, "test.array[0]", "value")
if err != nil {
    fmt.Println(err)
}

// set a complex path
err = SetJsonPath(data, "test.array[1].map[0][0].key", "value")
if err != nil {
    fmt.Println(err)
}

// set an object as a value
err = SetJsonPath(data, "test.object", map[string]interface{}{"test":123})
if err != nil {
    fmt.Println(err)
}
```

## Get Values

Retrieve values in your data using the following.

```
// get a value at a path
val, err := GetJsonPath(data, "test.path")
if err != nil {
    panic(err)
}
fmt.Println(val)

// get a value at an array index
val, err := GetJsonPath(data, "test.array[0]")
if err != nil {
    panic(err)
}
fmt.Println(val)
```
