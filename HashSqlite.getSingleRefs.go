package hashsqlite

import (
	"reflect"
)
 
func (hs *HashSqlite) getSingleRefs(row interface{}) (string, reflect.Value, error) {
	var rowType string
	var tabName string
	var reftab reflect.Type
	var refRow reflect.Value
	var ok bool

	reftab = reflect.TypeOf(row)

	if reftab.Kind() != reflect.Pointer {
		Goose.Query.Logf(1, "Parameter type error: %s", ErrNotStructPointer)
		return "", refRow, ErrNotStructPointer
	}

	reftab = reftab.Elem()
	if reftab.Kind() != reflect.Struct {
		Goose.Query.Logf(1, "Parameter type error: %s", ErrNotStructPointer)
		return "", refRow, ErrNotStructPointer
	}

	rowType = reftab.Name()
	if tabName, ok = hs.tableType[rowType]; !ok {
		Goose.Query.Logf(1, "Parameter type error: %s", ErrNoTablesFound)
		return "", refRow, ErrNoTablesFound
	}

	refRow = reflect.ValueOf(row)
	if !refRow.IsValid() || refRow.IsNil() || refRow.IsZero() {
		Goose.Query.Logf(1, "Parameter error", ErrInvalid)
		return "", refRow, ErrInvalid
	}

	return tabName, refRow.Elem(), nil
}