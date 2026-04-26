package db

import "reflect"

func reflectValue(obj interface{}) reflect.Value {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

func reflectType(obj interface{}) reflect.Type {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}

	return t
}

func reflectSliceType(collection interface{}) reflect.Type {
	v := reflect.ValueOf(collection)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Len() == 0 {
		t := v.Type()
		for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
			t = t.Elem()
		}
		return t
	}
	v = v.Index(0)
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Type()
}

// makeNew creates a new object.
func makeNew(t reflect.Type) interface{} {
	return reflect.New(t).Interface()
}

func makeSliceOfType(t reflect.Type) interface{} {
	return reflect.New(reflect.SliceOf(t)).Interface()
}

// haveSameUnderlyingTypes returns if T and V are such that V is *T or V is **T etc.
//
// It handles the cases where we're assigning T = convert(**T) which can happen when we're setting up
// scan output array.
//
// Convert can smush T and **T together somehow.
func haveSameUnderlyingTypes(t, v reflect.Value) bool {
	tt := t.Type()
	tv := reflectType(v)
	return tv.AssignableTo(tt) || tv.ConvertibleTo(tt)
}
