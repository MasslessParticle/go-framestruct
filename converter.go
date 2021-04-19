package framestruct

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

const frameTag = "frame"

var (
	fieldNames []string
	fields     map[string]*data.Field
)

// ToDataframe flattens an arbitrary struct or slice of sructs into a *data.Frame
func ToDataframe(name string, toConvert interface{}) (*data.Frame, error) {
	fields = make(map[string]*data.Field)
	fieldNames = make([]string, 0)

	v := ensureValue(reflect.ValueOf(toConvert))
	switch v.Kind() {
	case reflect.Slice:
		if err := convertSlice(v); err != nil {
			return nil, err
		}
	case reflect.Struct:
		if err := convertField(v); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported type: can only convert structs or slices of structs")
	}

	//add to frame, iterate to preserve order
	frame := data.NewFrame(name)
	for _, f := range fieldNames {
		frame.Fields = append(frame.Fields, fields[f])
	}

	return frame, nil
}

func convertSlice(s reflect.Value) error {
	for i := 0; i < s.Len(); i++ {
		if err := convertField(s.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func convertField(f reflect.Value) error {
	v := ensureValue(f)
	if err := makeFields(v, ""); err != nil {
		return err
	}

	return nil
}

func makeFields(v reflect.Value, prefix string) error {
	if v.Kind() != reflect.Struct {
		return errors.New("unspported type: cannot convert types witout fields")
	}

	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}

		structField := v.Type().Field(i)
		if structField.Tag.Get(frameTag) == "-" {
			continue
		}

		fieldName := fieldName(structField, prefix)
		switch v.Field(i).Kind() {
		case reflect.Struct:
			makeFields(v.Field(i), fieldName)
		default:
			if err := createField(v.Field(i), fieldName); err != nil {
				return err
			}
			fields[fieldName].Append(v.Field(i).Interface())
		}
	}
	return nil
}

func createField(v reflect.Value, fieldName string) error {
	if _, exists := fields[fieldName]; !exists {
		//keep track of unique fields in the order they appear
		fieldNames = append(fieldNames, fieldName)
		v, err := sliceFor(v.Interface())
		if err != nil {
			return err
		}

		fields[fieldName] = data.NewField(fieldName, nil, v)
	}
	return nil
}

func fieldName(v reflect.StructField, prefix string) string {
	fName := v.Name
	if tag := v.Tag.Get(frameTag); tag != "" {
		fName = tag
	}

	if prefix == "" {
		return fName
	}
	return fmt.Sprintf("%s.%s", prefix, fName)
}

func ensureValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}
