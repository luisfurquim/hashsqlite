package hashsqlite

import (
	"reflect"
)
 
func (hs *HashSqlite) getMultiRefs(row interface{}) (string, reflect.Value, bool, error) {
	var rowType string
	var tabName string
	var reftab reflect.Type
	var refRow reflect.Value
	var ok bool
	var isChan bool

	reftab = reflect.TypeOf(row)

	if reftab.Kind() == reflect.Chan {
		isChan = true
		reftab = reftab.Elem()

		if reftab.Kind() != reflect.Pointer {
			Goose.Query.Logf(1, "Parameter type error: %s", ErrNotStructPointerChan)
			return "", refRow, isChan, ErrNotStructPointerChan
		}

		reftab = reftab.Elem()
		if reftab.Kind() != reflect.Struct {
			Goose.Query.Logf(1, "Parameter type error: %s", ErrNotStructPointerChan)
			return "", refRow, isChan, ErrNotStructPointerChan
		}
	} else {
		if reftab.Kind() != reflect.Pointer {
			Goose.Query.Logf(1, "Parameter type error: %s", ErrNotStructSlicePointer)
			return "", refRow, isChan, ErrNotStructSlicePointer
		}

		reftab = reftab.Elem()
		if reftab.Kind() != reflect.Slice {
			Goose.Query.Logf(1, "Parameter type error: %s", ErrNotStructSlicePointer)
			return "", refRow, isChan, ErrNotStructSlicePointer
		}

		reftab = reftab.Elem()
		if reftab.Kind() != reflect.Struct {
			Goose.Query.Logf(1, "Parameter type error: %s", ErrNotStructSlicePointer)
			return "", refRow, isChan, ErrNotStructSlicePointer
		}
	}

	rowType = reftab.Name()
	if tabName, ok = hs.tableType[rowType]; !ok {
		Goose.Query.Logf(1, "Parameter type error: %s", ErrNoTablesFound)
		return "", refRow, isChan, ErrNoTablesFound
	}

	refRow = reflect.ValueOf(row)
	if !refRow.IsValid() || refRow.IsNil() || refRow.IsZero() {
		Goose.Query.Logf(1, "Parameter error", ErrInvalid)
		return "", refRow, isChan, ErrInvalid
	}

	if isChan {
		return tabName, refRow, isChan, nil
	}

	return tabName, refRow.Elem(), isChan, nil
}