package hashsqlite

import (
	"reflect"
)

func isNonEmptySlice(v reflect.Value) bool {

	if v.Kind() == reflect.Pointer {
		if !v.IsValid() || v.IsNil() || v.IsZero() {
			return false
		}

		v = v.Elem()
	}

	if !v.IsValid() || v.IsNil() || v.IsZero() {
		return false
	}

	if v.Kind() != reflect.Slice {
		return false
	}

	return v.Len()>0
}
