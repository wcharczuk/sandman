package viewutil

import (
	"fmt"
	"html/template"
	"reflect"
	"strings"
	"time"

	"go.charczuk.com/sdk/stringutil"
)

// FormFor yields an html form input for a given value.
//
// The value is reflected and the `form:"..."` struct tags
// are used to set options for individual fields.
func FormFor(obj any, action string) template.HTML {
	if typed, ok := obj.(FormProvider); ok {
		return typed.Form(action)
	}

	rt := reflect.TypeOf(obj)
	rv := reflect.ValueOf(obj)
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintf(`<form method="POST" data-form-for="%s" action="%s">`, rt.Name(), action))
	sb.WriteString("\n")
	for x := 0; x < rt.NumField(); x++ {
		f := rt.Field(x)
		v := rv.Field(x)
		input := controlForField(f, v)
		if input != "" {
			sb.WriteString(input)
			sb.WriteString("\n")
		}
	}
	sb.WriteString(`<input type="submit">Submit</input>`)
	sb.WriteString("\n")
	sb.WriteString("</form>")
	return template.HTML(sb.String())
}

// FormProvider will shortcut reflection in the `FormFor` call.
type FormProvider interface {
	Form(string) template.HTML
}

// InputProvider will shortcut reflection in the input generation step
// of a `FormFor` or `ControlFor` call.
type InputProvider interface {
	Input() template.HTML
}

// ControlFor returns just the input for a given struct field by name.
func ControlFor(obj any, fieldName string) template.HTML {
	rt := reflect.TypeOf(obj)
	fieldType, ok := rt.FieldByName(fieldName)
	if !ok {
		return ""
	}
	rv := reflect.ValueOf(obj)
	fieldValue := rv.FieldByName(fieldName)
	return template.HTML(controlForField(fieldType, fieldValue))
}

func controlForField(field reflect.StructField, fieldValue reflect.Value) string {
	tag := parseFormStructTag(field.Tag.Get("form"))
	if tag.Skip {
		return ""
	}

	// drill through the value itself
	// and if it's not nil, make a string version of it
	for fieldValue.Kind() == reflect.Ptr {
		fieldValue = fieldValue.Elem()
	}
	fieldValueElem := fieldValue.Interface()

	inputName := field.Name
	var label string
	if tag.Label != "" {
		label = fmt.Sprintf(`<label for="%s">%s</label>`, inputName, tag.Label)
	}

	var attrs []string
	attrs = append(attrs, fmt.Sprintf(`name="%s"`, inputName))
	attrs = append(attrs, inputTypeForControl(field, fieldValue, fieldValueElem, tag))
	attrs = append(attrs, tag.Attrs...)

	switch t := fieldValueElem.(type) {
	case bool:
		if t {
			attrs = append(attrs, "checked")
		}
	case *bool:
		if t != nil && *t {
			attrs = append(attrs, "checked")
		}
	default:
		var value string
		if fieldValueElem != nil {
			value = fmt.Sprint(fieldValueElem)
		}
		if value != "" {
			attrs = append(attrs, fmt.Sprintf(`value="%s"`, value))
		}
	}
	return label + "<input " + strings.Join(attrs, " ") + "/>"
}

func inputTypeForControl(fieldType reflect.StructField, fieldTypeValue reflect.Value, fieldValue any, tag formStructTag) string {
	if tag.InputType != "" {
		return fmt.Sprintf(`type="%s"`, tag.InputType)
	}
	switch fieldValue.(type) {
	case string, *string:
		return `type="text"`
	case time.Time, *time.Time:
		return `type="datetime-local"`
	case bool, *bool:
		return `type="checkbox"`
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return `type="number"`
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64:
		return `type="number"`
	default:
		return `type="text"`
	}
}

type formStructTag struct {
	Skip      bool
	Label     string
	InputType string
	Attrs     []string
}

func parseFormStructTag(tagValue string) (output formStructTag) {
	if tagValue == "-" {
		output.Skip = true
		return
	}

	fields := stringutil.SplitQuoted(tagValue, ",")
	for _, field := range fields {
		key, value, _ := strings.Cut(field, "=")
		switch key {
		case "label":
			output.Label = stringutil.TrimQuotes(value)
		case "type":
			output.InputType = stringutil.TrimQuotes(value)
		default:
			output.Attrs = append(output.Attrs, field)
		}
	}
	return
}
