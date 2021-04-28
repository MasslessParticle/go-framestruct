# framestruct

framestruct is a simple library for flattening structs, slices, and maps
into pointers to grafana data.Frames.

## Supported Types

framestruct supports conversions of the following types:

- structs
- `map[string]interface{}`
- slices of structs or `map[string]interface{}`

Structs may contain `map[string]interface{}` and maps may contain structs.

### A note on maps

To preserve ordering across runs with maps, framestruct storts fieldnames.
If you want to control the order use a struct or specialy designed map keys.

## Usage

Take a struct with supported types and call `ToDataframe`.

```go
package main

import (
	"fmt"

	"github.com/masslessparticle/go-framestruct"
)

type structWithTags struct {
	Thing1 string  `frame:"first-thing"`
	Thing2 string  `frame:"-"`
	Thing3 nested2 `frame:"third-thing"`
	Thing4 nested2 `frame:",omitparent"`
}

type nested2 struct {
	Thing5 bool
	Thing6 int64
}

func main() {
	strct := structWithTags{
		"foo",
		"bar",
		nested2{
			true,
			100,
		},
		nested2{
			false,
			200,
		},
	}

	frame, err := framestruct.ToDataframe("FrameName", strct)
	if err != nil {
		panic(err)
	}

	fmt.Println(frame.Name)
	fmt.Println(len(frame.Fields))

	fmt.Println(frame.Fields[0].Name)
	fmt.Println(frame.Fields[0].At(0))

	fmt.Println(frame.Fields[2].Name)
	fmt.Println(frame.Fields[2].At(0))

	fmt.Println(frame.Fields[4].Name)
	fmt.Println(frame.Fields[4].At(0))
}
```

Run your code and rejoice!

```
FrameName
3
first-thing
foo
third-thing.Thing6
100
Thing6
200
```

## Struct Tags

- Use the `frame` struct tag to configure conversion behavior. a custom name.
- Use `-` to exclude a field from the output.
- Other flags must be used in the following order
  1. `fieldname`: The first tag present will override the Dataframe column name. By default, framestruct uses the name of the struct field.
  1. `omitparent`: When present, will tell framestruct to use the name of `child` rather than `parent.child` as the Dataframe column name.
  1. `col0`: When present, will make this the 0th column of the Dataframe. Only the first instance of `col0` is respected

### A Note on Maps in struct fields

Maps are treated like a child of their struct field. Maps inherit their parent fields tags. The only supported tag on a map field is `,omitparent`
all other tags result in undefined behavior

e.g.:

```go
type allStructTags struct {
	Foo barBaz
}

type barBaz struct {
	Bar string                 `frame:"zzz,omitparent,col0"`
	Baz map[string]interface{} `frame:",omitparent"`
}

allStructTags{
	Foo: barBaz{
		Bar: "should be first",
		Baz: map[string]interface{}{
			"aaa": "foo",
			"bbb": "foo",
			"ccc": "foo",
		},
	},
}
```

results in the following fields in the following order:

```
zzz
aaa
bbb
ccc
```

## Running Tests

Run tests using go test.

```
$ go test ./...
```

Run benchmarks using go test.

```
$  go test -bench=. -benchmem -run=^$
```
