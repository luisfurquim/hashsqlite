package hashsqlite

import (
	"reflect"
)

func (hs *HashSqlite) Save(row interface{}) (int64, error) {
	var id int64
	var tabName string
	var refRow reflect.Value
	var err error
	var parms []interface{}

	tabName, refRow, err = hs.getSingleRefs(row)
	if err != nil {
		return 0, err
	}

	Goose.Query.Logf(0, "tabName: %s, refRow: %#v", tabName, refRow)

	id, parms = hs.getParmValues(tabName, refRow, func(r interface{}) {
		// recursively save related tables
		hs.Save(r)
	})

	Goose.Query.Logf(0, "id:%d, parms: %#v", id, parms)

	if id==0 {
		Goose.Query.Logf(0, "hs.insert:%#v, tabName: %s", hs.insert, tabName)
		id, err = hs.insert[tabName].Insert(parms...)
		if err != nil {
			Goose.Query.Logf(1, "Insert error on %s: %s", tabName, err)
			return 0, err
		}
		if hs.tables[tabName].pkName != "" {
			refRow.Field(hs.tables[tabName].pkIndex).SetInt(id)
		}
	} else {
		parms = append(parms, id)
		err = hs.updateBy[tabName]["id"].Exec(parms...)
		if err != nil {
			Goose.Query.Logf(1, "Update error on %s: %s", tabName, err)
			return 0, err
		}
	}

	return id, nil
}
