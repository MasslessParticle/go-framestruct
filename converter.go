package framestruct

import (
	"errors"
	"reflect"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

const frameTag = "frame"

type converter struct {
	fieldNames []string
	fields     map[string]*data.Field
	tags       []string
}

// ToDataframe flattens an arbitrary struct or slice of structs into a *data.Frame
func ToDataframe(name string, toConvert interface{}) (*data.Frame, error) {
	cr := &converter{
		fields: make(map[string]*data.Field),
		tags:   make([]string, 2),
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
		return errors.New("unsupported type: cannot convert types without fields")
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanInterface() {
			continue
		}

		structField := v.Type().Field(i)
		tags := structField.Tag.Get(frameTag)

		if tags == "-" {
			continue
		}

		fieldName := c.fieldName(structField.Name, tags, prefix)
		switch field.Kind() {
		case reflect.Struct:
			if err := c.makeFields(field, fieldName); err != nil {
				return err
			}
		default:
			if err := c.createField(field, fieldName); err != nil {
				return err
			}
			c.fields[fieldName].Append(field.Interface())
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

func (c *converter) fieldName(fieldName, tags, prefix string) string {
	c.parseTags(tags, ",")
	if c.tags[0] == "omitparent" || c.tags[1] == "omitparent" {
		return ""
	}

	if c.tags[0] != "" {
		fieldName = c.tags[0]
	}

	if prefix == "" {
		return fieldName
	}

	return prefix + "." + fieldName
}

func (c *converter) ensureValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

func (c *converter) parseTags(s, sep string) {
	// if we do it this way, we avoid all the allocs
	// of strings.Split
	c.tags[0] = ""
	c.tags[1] = ""

	i := 0
	for i < 2 {
		m := strings.Index(s, sep)
		if m < 0 {
			break
		}
		c.tags[i] = s[:m]
		s = s[m+len(sep):]
		i++
	}
	c.tags[i] = s
}
