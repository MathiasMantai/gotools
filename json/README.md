This package provides functions to handle json data and files


## IntoStruct

### Example
We have the following json file named test.json
```json
{
    "name": "test",
    "author": "testauthor"
}
```

Then we can use the IntoStruct function to get the data like this:

```go
type JsonTest struct {
	Name string `json:"name"`
	Author string `json:"author"`
}

func main() {
    var test JsonTest
	json.IntoStruct("./test.json", &test)
}
```