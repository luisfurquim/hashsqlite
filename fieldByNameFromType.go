package hashsqlite

import (
	"reflect"
)

func fieldByNameFromType(nm string, row reflect.Type) (int, bool) {
	var i int

	for i=0; i<row.NumField(); i++ {
		if nm == row.Field(i).Name {
			return i, true
		}
	}

	return -1, false
}
