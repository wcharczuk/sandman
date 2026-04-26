package db

import "reflect"

// TableNameByType returns the table name for a given reflect.Type by instantiating it and calling o.TableName().
// The type must implement DatabaseMapped or an exception will be returned.
func TableNameByType(t reflect.Type) string {
	instance := reflect.New(t).Interface()
	if typed, isTyped := instance.(TableNameProvider); isTyped {
		return typed.TableName()
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		instance = reflect.New(t).Interface()
		if typed, isTyped := instance.(TableNameProvider); isTyped {
			return typed.TableName()
		}
	}
	return t.Name()
}

// TableName returns the mapped table name for a given instance; it will sniff for the `TableName()` function on the type.
func TableName(obj any) string {
	if typed, isTyped := obj.(TableNameProvider); isTyped {
		return typed.TableName()
	}
	return reflectType(obj).Name()
}
