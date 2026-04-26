package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// NewColumnFromFieldTag reads the contents of a field tag, ex: `json:"foo" db:"bar,isprimarykey,isserial"
func NewColumnFromFieldTag(field reflect.StructField) *Column {
	db := field.Tag.Get("db")
	if db != "-" {
		col := Column{}
		col.FieldName = field.Name
		col.ColumnName = field.Name
		col.FieldType = field.Type
		if db != "" {
			pieces := strings.Split(db, ",")
			if !strings.HasPrefix(db, ",") {
				// note: split will return the original string
				// if the string _does not contain_ the split string
				// so this index is safe to do, but we also
				// generally want the first token in the csv
				// regardless (why we use pieces at all)
				col.ColumnName = pieces[0]
			}
			if len(pieces) > 1 {
				for _, p := range pieces[1:] {
					switch strings.TrimSpace(strings.ToLower(p)) {
					case "pk":
						col.IsPrimaryKey = true
					case "uk":
						col.IsUniqueKey = true
					case "auto", "serial":
						col.IsAuto = true
					case "readonly":
						col.IsReadOnly = true
					case "inline":
						col.Inline = true
					case "json":
						col.IsJSON = true
					default:
						panic("invalid struct tag key; " + p)
					}
				}
			}
		}
		return &col
	}

	return nil
}

// Column represents a single field on a struct that is mapped to the database.
type Column struct {
	Parent       *Column
	TableName    string
	FieldName    string
	FieldType    reflect.Type
	ColumnName   string
	Index        int
	IsPrimaryKey bool
	IsUniqueKey  bool
	IsAuto       bool
	IsReadOnly   bool
	IsJSON       bool
	Inline       bool
}

// SetValue sets the field on a database mapped object to the instance of `value`.
func (c Column) SetValue(reference, value any) error {
	return c.SetValueReflected(reflectValue(reference), value)
}

// SetValueReflected sets the field on a reflect value object to the instance of `value`.
func (c Column) SetValueReflected(reference reflect.Value, value any) error {
	objectField := reference.FieldByName(c.FieldName)

	// check if we've been passed a reference for the target object
	if !objectField.CanSet() {
		return fmt.Errorf("hit a field we can't set; did you forget to pass the object as a reference? field: %s", c.FieldName)
	}

	// special case for `db:"...,json"` fields.
	if c.IsJSON {
		var deserialized interface{}
		if objectField.Kind() == reflect.Ptr {
			deserialized = reflect.New(objectField.Type().Elem()).Interface()
		} else {
			deserialized = objectField.Addr().Interface()
		}

		switch valueContents := value.(type) {
		case *sql.NullString:
			if !valueContents.Valid {
				objectField.Set(reflect.Zero(objectField.Type()))
				return nil
			}
			if err := json.Unmarshal([]byte(valueContents.String), deserialized); err != nil {
				return err
			}
		case sql.NullString:
			if !valueContents.Valid {
				objectField.Set(reflect.Zero(objectField.Type()))
				return nil
			}
			if err := json.Unmarshal([]byte(valueContents.String), deserialized); err != nil {
				return err
			}
		case *string:
			if err := json.Unmarshal([]byte(*valueContents), deserialized); err != nil {
				return err
			}
		case string:
			if err := json.Unmarshal([]byte(valueContents), deserialized); err != nil {
				return err
			}
		case *[]byte:
			if err := json.Unmarshal(*valueContents, deserialized); err != nil {
				return err
			}
		case []byte:
			if err := json.Unmarshal(valueContents, deserialized); err != nil {
				return err
			}
		default:
			return fmt.Errorf("set value; invalid type for assignment to json field; field %s", c.FieldName)
		}

		if rv := reflect.ValueOf(deserialized); !rv.IsValid() {
			objectField.Set(reflect.Zero(objectField.Type()))
		} else {
			if objectField.Kind() == reflect.Ptr {
				objectField.Set(rv)
			} else {
				objectField.Set(rv.Elem())
			}
		}
		return nil
	}

	valueReflected := reflectValue(value)
	if !valueReflected.IsValid() { // if the value is nil
		objectField.Set(reflect.Zero(objectField.Type())) // zero the field
		return nil
	}

	// if we can direct assign the value to the field
	if valueReflected.Type().AssignableTo(objectField.Type()) {
		objectField.Set(valueReflected)
		return nil
	}

	// convert and assign
	if valueReflected.Type().ConvertibleTo(objectField.Type()) ||
		haveSameUnderlyingTypes(objectField, valueReflected) {
		objectField.Set(valueReflected.Convert(objectField.Type()))
		return nil
	}

	if objectField.Kind() == reflect.Ptr && valueReflected.CanAddr() {
		if valueReflected.Addr().Type().AssignableTo(objectField.Type()) {
			objectField.Set(valueReflected.Addr())
			return nil
		}
		if valueReflected.Addr().Type().ConvertibleTo(objectField.Type()) {
			objectField.Set(valueReflected.Convert(objectField.Elem().Type()).Addr())
			return nil
		}
		return fmt.Errorf("set value; can addr value but can't figure out how to assign or convert; field: %s", c.FieldName)
	}

	return fmt.Errorf("set value; ran out of ways to set the field; field: %s", c.FieldName)
}

// GetValue returns the value for a column on a given database mapped object.
func (c Column) GetValue(object any) interface{} {
	value := reflectValue(object)
	if c.Parent != nil {
		embedded := value.Field(c.Parent.Index)
		valueField := embedded.Field(c.Index)
		return valueField.Interface()
	}
	valueField := value.Field(c.Index)
	return valueField.Interface()
}
