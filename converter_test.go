package framestruct_test

import (
	"testing"
	"time"

	"github.com/masslessparticle/go-framestruct"
	"github.com/stretchr/testify/require"
)

func TestToDataframe(t *testing.T) {
	t.Run("it flattens a struct", func(t *testing.T) {
		strct := simpleStruct{"foo", 36, "baz"}

		frame, err := framestruct.ToDataframe("Results", strct)
		require.Nil(t, err)

		require.Equal(t, "Results", frame.Name)
		require.Len(t, frame.Fields, 3)

		require.Equal(t, "foo", frame.Fields[0].At(0))
		require.Equal(t, int32(36), frame.Fields[1].At(0))
		require.Equal(t, "baz", frame.Fields[2].At(0))
	})

	t.Run("it flattens a pointer to a struct", func(t *testing.T) {
		strct := simpleStruct{"foo", 36, "baz"}

		frame, err := framestruct.ToDataframe("Results", &strct)
		require.Nil(t, err)

		require.Equal(t, "Results", frame.Name)
		require.Len(t, frame.Fields, 3)

		require.Equal(t, "foo", frame.Fields[0].At(0))
		require.Equal(t, int32(36), frame.Fields[1].At(0))
		require.Equal(t, "baz", frame.Fields[2].At(0))
	})

	t.Run("it flattens a slice of structs", func(t *testing.T) {
		strct := []simpleStruct{
			{"foo", 36, "baz"},
			{"foo1", 37, "baz1"},
		}

		frame, err := framestruct.ToDataframe("results", strct)
		require.Nil(t, err)

		require.Len(t, frame.Fields, 3)
		require.Equal(t, 2, frame.Fields[0].Len())
		require.Equal(t, 2, frame.Fields[1].Len())
		require.Equal(t, 2, frame.Fields[2].Len())

		require.Equal(t, "foo", frame.Fields[0].At(0))
		require.Equal(t, "foo1", frame.Fields[0].At(1))

		require.Equal(t, int32(36), frame.Fields[1].At(0))
		require.Equal(t, int32(37), frame.Fields[1].At(1))

		require.Equal(t, "baz", frame.Fields[2].At(0))
		require.Equal(t, "baz1", frame.Fields[2].At(1))
	})

	t.Run("it flattens a pointer to a slice of structs", func(t *testing.T) {
		strct := []simpleStruct{
			{"foo", 36, "baz"},
			{"foo1", 37, "baz1"},
		}

		frame, err := framestruct.ToDataframe("results", &strct)
		require.Nil(t, err)

		require.Len(t, frame.Fields, 3)
		require.Equal(t, 2, frame.Fields[0].Len())
		require.Equal(t, 2, frame.Fields[1].Len())
		require.Equal(t, 2, frame.Fields[2].Len())

		require.Equal(t, "foo", frame.Fields[0].At(0))
		require.Equal(t, "foo1", frame.Fields[0].At(1))

		require.Equal(t, int32(36), frame.Fields[1].At(0))
		require.Equal(t, int32(37), frame.Fields[1].At(1))

		require.Equal(t, "baz", frame.Fields[2].At(0))
		require.Equal(t, "baz1", frame.Fields[2].At(1))
	})

	t.Run("it propertly handles pointers", func(t *testing.T) {
		foo := "foo"
		strct := pointerStruct{&foo}

		frame, err := framestruct.ToDataframe("results", strct)
		require.Nil(t, err)

		require.Len(t, frame.Fields, 1)

		val := frame.Fields[0].At(0).(*string)
		require.Equal(t, "foo", *val)
	})

	t.Run("it ignores unexported fields", func(t *testing.T) {
		strct := noExportedFields{"no!"}

		frame, err := framestruct.ToDataframe("results", strct)
		require.Nil(t, err)

		require.Len(t, frame.Fields, 0)
	})

	t.Run("it ignores fields when the struct tag is a '-'", func(t *testing.T) {
		strct := structWithIgnoredTag{"foo", "bar", "baz"}

		frame, err := framestruct.ToDataframe("results", strct)
		require.Nil(t, err)

		require.Len(t, frame.Fields, 2)
		require.Equal(t, "first-thing", frame.Fields[0].Name)
		require.Equal(t, "foo", frame.Fields[0].At(0))

		require.Equal(t, "third-thing", frame.Fields[1].Name)
		require.Equal(t, "baz", frame.Fields[1].At(0))
	})

	t.Run("it flattens nested structs with dot-names", func(t *testing.T) {
		strct := []nested1{
			{"foo", 36, "baz",
				nested2{true, 100},
			},
			{"foo1", 37, "baz1",
				nested2{false, 101},
			},
		}

		frame, err := framestruct.ToDataframe("results", strct)
		require.Nil(t, err)

		require.Len(t, frame.Fields, 5)
		require.Equal(t, 2, frame.Fields[0].Len())
		require.Equal(t, 2, frame.Fields[1].Len())
		require.Equal(t, 2, frame.Fields[2].Len())
		require.Equal(t, 2, frame.Fields[3].Len())
		require.Equal(t, 2, frame.Fields[4].Len())

		require.Equal(t, "Thing1", frame.Fields[0].Name)
		require.Equal(t, "foo", frame.Fields[0].At(0))
		require.Equal(t, "foo1", frame.Fields[0].At(1))

		require.Equal(t, "Thing2", frame.Fields[1].Name)
		require.Equal(t, int32(36), frame.Fields[1].At(0))
		require.Equal(t, int32(37), frame.Fields[1].At(1))

		require.Equal(t, "Thing3", frame.Fields[2].Name)
		require.Equal(t, "baz", frame.Fields[2].At(0))
		require.Equal(t, "baz1", frame.Fields[2].At(1))

		require.Equal(t, "Thing4.Thing5", frame.Fields[3].Name)
		require.Equal(t, true, frame.Fields[3].At(0))
		require.Equal(t, false, frame.Fields[3].At(1))

		require.Equal(t, "Thing4.Thing6", frame.Fields[4].Name)
		require.Equal(t, int64(100), frame.Fields[4].At(0))
		require.Equal(t, int64(101), frame.Fields[4].At(1))
	})

	t.Run("it uses struct tags if they're present", func(t *testing.T) {
		strct := structWithTags{
			"foo",
			"bar",
			nested2{
				true,
				100,
			},
		}

		frame, err := framestruct.ToDataframe("results", strct)
		require.Nil(t, err)

		require.Len(t, frame.Fields, 4)
		require.Equal(t, "first-thing", frame.Fields[0].Name)
		require.Equal(t, "foo", frame.Fields[0].At(0))

		require.Equal(t, "second-thing", frame.Fields[1].Name)
		require.Equal(t, "bar", frame.Fields[1].At(0))

		require.Equal(t, "third-thing.Thing5", frame.Fields[2].Name)
		require.Equal(t, true, frame.Fields[2].At(0))

		require.Equal(t, "third-thing.Thing6", frame.Fields[3].Name)
		require.Equal(t, int64(100), frame.Fields[3].At(0))
	})

	t.Run("omits the parent struct name if omitparent is present", func(t *testing.T) {
		strct := omitParentStruct{
			"foo",
			"bar",
			nested2{
				true,
				100,
			},
		}

		frame, err := framestruct.ToDataframe("results", strct)
		require.Nil(t, err)

		require.Len(t, frame.Fields, 4)
		require.Equal(t, "first-thing", frame.Fields[0].Name)
		require.Equal(t, "foo", frame.Fields[0].At(0))

		require.Equal(t, "second-thing", frame.Fields[1].Name)
		require.Equal(t, "bar", frame.Fields[1].At(0))

		require.Equal(t, "Thing5", frame.Fields[2].Name)
		require.Equal(t, true, frame.Fields[2].At(0))

		require.Equal(t, "Thing6", frame.Fields[3].Name)
		require.Equal(t, int64(100), frame.Fields[3].At(0))
	})

	t.Run("it returns an error when the struct contains an unsupported type", func(t *testing.T) {
		strct := unsupportedType{32}

		_, err := framestruct.ToDataframe("results", strct)
		require.Error(t, err)
		require.Equal(t, "unsupported type int", err.Error())
	})

	t.Run("it returns an error when a nested struct contains an unsupported type", func(t *testing.T) {
		strct := supportedWithUnsupported{
			"foo",
			unsupportedType{
				100,
			},
		}

		_, err := framestruct.ToDataframe("results", strct)
		require.Error(t, err)
		require.Equal(t, "unsupported type int", err.Error())
	})

	t.Run("it returns an error when any struct contains an unsupported type", func(t *testing.T) {
		strct := unsupportedType{32}

		_, err := framestruct.ToDataframe("results", []unsupportedType{strct})
		require.Error(t, err)
		require.Equal(t, "unsupported type int", err.Error())
	})

	t.Run("it returns an error when non struct types are passed in", func(t *testing.T) {
		_, err := framestruct.ToDataframe("???", []string{"1", "2"})
		require.Error(t, err)

		m := make(map[string]string)
		_, err = framestruct.ToDataframe("???", m)
		require.Error(t, err)

		_, err = framestruct.ToDataframe("???", "can't do a string either")
		require.Error(t, err)
	})

	// This test fails when run with -race when it's not threadsafe
	t.Run("it is threadsafe", func(t *testing.T) {

		start := make(chan struct{})
		end := make(chan struct{})

		go convertStruct(start, end)
		go convertStruct(start, end)

		close(start)
		time.Sleep(20 * time.Millisecond)
		close(end)
	})
}

func convertStruct(start, end chan struct{}) {
	strct := structWithTags{
		"foo",
		"bar",
		nested2{
			true,
			100,
		},
	}

	<-start
	for {
		select {
		case <-end:
			return
		default:
			framestruct.ToDataframe("frame", strct)
		}
	}
}

type noExportedFields struct {
	unexported string
}

type simpleStruct struct {
	Thing1 string
	Thing2 int32
	Thing3 string
}

type nested1 struct {
	Thing1 string
	Thing2 int32
	Thing3 string
	Thing4 nested2
}

type nested2 struct {
	Thing5 bool
	Thing6 int64
}

type structWithTags struct {
	Thing1 string  `frame:"first-thing"`
	Thing2 string  `frame:"second-thing"`
	Thing3 nested2 `frame:"third-thing"`
}

type omitParentStruct struct {
	Thing1 string  `frame:"first-thing"`
	Thing2 string  `frame:"second-thing"`
	Thing3 nested2 `frame:",omitparent"`
}

type structWithIgnoredTag struct {
	Thing1 string `frame:"first-thing"`
	Thing2 string `frame:"-"`
	Thing3 string `frame:"third-thing"`
}

type supportedWithUnsupported struct {
	Foo string
	Bar unsupportedType
}

type unsupportedType struct {
	Foo int
}

type pointerStruct struct {
	Foo *string
}
