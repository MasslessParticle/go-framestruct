# framestruct

framestruct is a simple library for flattening structs and slices of structs
into grafana data.Frame pointers.

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
```

## Field Names

By default, framestruct will use the name of the struct field as the Dataframe column name.
To change this behavior, use the `frame` struct tag. To ignore a field, use `-`.

For nested structs, framestruct with name the Dataframe column in the form `parent.child`

## Running Tests

Run tests using go test.

```
$ go test ./...
```

Run benchmarks using go test.

```
$  go test -bench=. -benchmem -run=^$
```
