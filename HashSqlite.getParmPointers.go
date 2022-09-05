package hashsqlite

import (
	"reflect"
)

func (hs *HashSqlite) getParmPointers(tabName string, refRow reflect.Value) (int64, []interface{}) {
	var fld field
	var parms []interface{}

	parms = make([]interface{}, 0, len(hs.tables[tabName].fields))
	for _, fld = range hs.tables[tabName].fields {
		if !fld.joinList {
			parms = append(parms, refRow.Field(fld.index).Addr().Interface())
		}
	}

	return refRow.Field(hs.tables[tabName].pkIndex).Interface().(int64), parms
}
