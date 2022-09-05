package hashsqlite

import (
	"reflect"
)

func (hs *HashSqlite) getParmValues(tabName string, refRow reflect.Value, recursive func(row interface{})) (int64, []interface{}) {
	var fld field
	var parms []interface{}
	var fk reflect.Value
	var id int64

	parms = make([]interface{}, 0, len(hs.tables[tabName].fields))
	for _, fld = range hs.tables[tabName].fields {
		if fld.joinList {
			continue
		}

		if fld.fk != "" {
			fk = refRow.Field(fld.index)
			if !fk.IsValid() || fk.IsNil() || fk.IsZero() {
				parms = append(parms, 0)
			} else {
				if recursive != nil {
					recursive(fk.Interface())
				}
				fk = fk.Elem().Field(hs.tables[fld.fk].pkIndex)
				parms = append(parms, fk.Interface())
			}
		} else {
			parms = append(parms, refRow.Field(fld.index).Interface())
		}
	}

	if hs.tables[tabName].pkName != "" {
		id = refRow.Field(hs.tables[tabName].pkIndex).Interface().(int64)
	}

	return id, parms
}
