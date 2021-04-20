package framestruct

import (
	"errors"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"reflect"
)

const frameTag = "frame"

type converter struct {
	fieldNames []string
	fields     map[string]*data.Field
}

// ToDataframe flattens an arbitrary struct or slice of sructs into a *data.Frame
func ToDataframe(name string, toConvert interface{}) (*data.Frame, error) {
	cr := &converter{
		fields: make(map[string]*data.Field),
	}

	return cr.toDataframe(name, toConvert)
}

func (c *converter) toDataframe(name string, toConvert interface{}) (*data.Frame, error) {
	v := c.ensureValue(reflect.ValueOf(toConvert))
	switch v.Kind() {
	case reflect.Slice:
		if err := c.convertSlice(v); err != nil {
			return nil, err
		}
	case reflect.Struct:
		if err := c.convertField(v); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported type: can only convert structs or slices of structs")
	}

	//add to frame, iterate to preserve order
	frame := data.NewFrame(name)
	for _, f := range c.fieldNames {
		frame.Fields = append(frame.Fields, c.fields[f])
	}

	return frame, nil
}

func (c *converter) convertSlice(s reflect.Value) error {
	for i := 0; i < s.Len(); i++ {
		if err := c.convertField(s.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func (c *converter) convertField(f reflect.Value) error {
	v := c.ensureValue(f)
	if err := c.makeFields(v, ""); err != nil {
		return err
	}

	return nil
}

func (c *converter) makeFields(v reflect.Value, prefix string) error {
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

		fieldName := c.fieldName(structField, prefix)
		switch v.Field(i).Kind() {
		case reflect.Struct:
			c.makeFields(v.Field(i), fieldName)
		default:
			if err := c.createField(v.Field(i), fieldName); err != nil {
				return err
			}
			c.fields[fieldName].Append(v.Field(i).Interface())
		}
	}
	return nil
}

func (c *converter) createField(v reflect.Value, fieldName string) error {
	if _, exists := c.fields[fieldName]; !exists {
		//keep track of unique fields in the order they appear
		c.fieldNames = append(c.fieldNames, fieldName)
		v, err := sliceFor(v.Interface())
		if err != nil {
			return err
		}

		c.fields[fieldName] = data.NewField(fieldName, nil, v)
	}
	return nil
}

func (c *converter) fieldName(v reflect.StructField, prefix string) string {
	fName := v.Name
	if tag := v.Tag.Get(frameTag); tag != "" {
		fName = tag
	}

	if prefix == "" {
		return fName
	}

	return prefix + "." + fName
}

func (c *converter) ensureValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}
