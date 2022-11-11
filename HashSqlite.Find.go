package hashsqlite

import (
	"strings"
	"reflect"
   "github.com/gwenn/gosqlite"
)

func (hs *HashSqlite) getList(tabName, rule string) (l *list, err error) {
	var ok bool
	var r map[string]*list

	if r, ok = hs.list[tabName]; !ok {
		Goose.Query.Logf(1, "Error listing table %s: %s", tabName, ErrNoTablesFound)
		err = ErrNoTablesFound
		return
	}

	if l, ok = r[rule]; !ok {
		Goose.Query.Logf(1, "Error listing table %s: %s", tabName, ErrRuleNotFound)
		Goose.Query.Logf(1, "rule %s: rules %#v", rule, hs.list)
		err = ErrRuleNotFound
		return
	}

	return
}

func (hs *HashSqlite) nextRow(tabName string, l *list, row reflect.Value, s *sqlite.Stmt, by map[string]interface{}, typ reflect.Type, getField func(reflect.Value, int) reflect.Value) error {
	var parms []interface{}
	var i, c int
	var fld reflect.Value
	var ok bool
	var lst tabRule
	var tName string
	var err error
	var related reflect.Value
	var relatedRow reflect.Value
	var fldName string
	var opt []string
	var fkeys map[int]interface{}
	var fkey interface{}

	parms = make([]interface{}, len(l.cols))
	fkeys = map[int]interface{}{}

	if by == nil {
		by = map[string]interface{}{}
	}

	// allocate the scan parameters
	for i, c = range l.cols {
		fld = getField(row, c)
		if fld.Type().Kind() == reflect.Pointer {
			fld.Set(reflect.New(fld.Type().Elem()))
			tName = hs.tableType[fld.Elem().Type().Name()]
			parms[i] = fld.Elem().Field(hs.tables[tName].pkIndex).Addr().Interface()
			if _, ok = l.joins[c]; ok {
				// let's keep a list of the parameters needed for joins (foreign keys)
				fkeys[c] = parms[i]
			}
		} else {
			parms[i] = fld.Addr().Interface()
		}
	}

	// Now we scan the current table (but not the joined ones...)
	err = s.Scan(parms...)
	if err != nil {
		Goose.Query.Logf(1, "Error scanning on list of table %s: %s", tabName, err)
		return err
	}

	for c, fkey = range fkeys {
		Goose.Query.Logf(0, "c:%d, fkey:%#v", c, *(fkey.(*int64)))
		fld = getField(row, c)
		if *(fkey.(*int64)) == 0 {
			fld.Set(reflect.Zero(fld.Type()))
			continue
		}
		lst = l.joins[c]
		if fldName, ok = typ.Field(c).Tag.Lookup("field"); ok && len(fldName)>0 {
			opt = strings.Split(fldName, ",")
			fldName = opt[0]
		} else {
			fldName = typ.Field(c).Name
		}

//		related = reflect.New(reflect.MakeSlice(reflect.SliceOf(fld.Type().Elem()),0,0).Type())
		related = reflect.New(reflect.SliceOf(fld.Type().Elem()))
		related.Elem().Set(reflect.Append(related.Elem(), fld.Elem()))
		by[fldName] = *(fkey.(*int64))
		Goose.Query.Logf(1, "related.Interface(): %#v", related.Interface())
		Goose.Query.Logf(1, "lst.rule: %#v", lst.rule)
		Goose.Query.Logf(1, "by=%#v", by)
		err = hs.Find(At{
			Table: related.Interface(),
			With: lst.rule,
			By: by,
		})
		if err != nil {
			return err
		}

		Goose.Query.Logf(1, ">>>related.Interface(): %#v", related.Interface())
		if related.Elem().Len() > 0 {
			if !fld.IsValid() || fld.IsNil() || fld.IsZero() {
				fld.Set(reflect.New(fld.Type().Elem()))
			}

			l, err = hs.getList(lst.table, lst.rule)
			if err != nil {
				return err
			}

			relatedRow = related.Elem().Index(0)
			for i, c = range l.cols {
				fld.Elem().Field(c).Set(relatedRow.Field(c))
			}

			Goose.Query.Logf(1, "related.Elem().Index(0): %#v", related.Elem().Index(0).Interface())
			Goose.Query.Logf(1, "fld: %#v", fld.Elem().Interface())
		}
	}

	return nil
}

//func (hs *HashSqlite) Find(tab interface{}, ruleAndFilter ...interface{}) error {
func (hs *HashSqlite) Find(at At) error {
	var tabName string
	var rule string
	var refRow reflect.Value
	var err error
	var isChan bool
	var l *list
	var index int
	var parmName string
	var parm interface{}

	tabName, refRow, isChan, err = hs.getMultiRefs(at.Table)
	if err != nil {
		Goose.Query.Logf(1, "Parameter type error: %s", ErrNoTablesFound)
		return ErrNoTablesFound
	}

	if len(at.With) == 0 {
		rule = "0"
	} else {
		rule = at.With
	}

	l, err = hs.getList(tabName, rule)
	if err != nil {
		return err
	}

	for parmName, parm = range at.By {
		index, err = l.stmt.BindParameterIndex(":" + parmName)
		Goose.Query.Logf(1, "will bind %s with %#v", parmName, parm)
		if err == nil {
			err = l.stmt.BindByIndex(index, parm)
			Goose.Query.Logf(1, "bound %s with %#v: %s", parmName, parm, err)
			if err != nil {
				Goose.Query.Logf(1, "Error binding on list of table %s for %s: %s", tabName, parmName, err)
				return err
			}
		} else {
			Goose.Query.Logf(1, "bind error %s with %#v => %s", parmName, parm, err)
		}
	}

	if isChan {
		go func() {
			err = l.stmt.Select(func(s *sqlite.Stmt) error {
				var row reflect.Value
				var err error

				row = reflect.New(refRow.Type().Elem().Elem())

				err = hs.nextRow(tabName, l, row, s, at.By, row.Elem().Type(), func(r reflect.Value, n int) reflect.Value {
					return r.Elem().Field(n)
				})

				if err != nil {
					return err
				}

				refRow.Send(row)
				
				return nil
			})
			refRow.Close()
			if err != nil {
				Goose.Query.Logf(1, "List error on %s: %s", tabName, err)
				return
			}
		}()
	} else {
		refRow.Set(reflect.MakeSlice(refRow.Type(), 0, 16))

		err = l.stmt.Select(func(s *sqlite.Stmt) error {
			var row reflect.Value
			var err error

			row = reflect.New(refRow.Type().Elem()).Elem() //?
			err = hs.nextRow(tabName, l, row, s, at.By, row.Type(), func(r reflect.Value, n int) reflect.Value {
				return r.Field(n)
			})

			if err != nil {
				return err
			}

			refRow.Set(reflect.Append(refRow, row))

			return nil
		})
		if err != nil {
			Goose.Query.Logf(1, "List error on %s: %s", tabName, err)
			return err
		}
	}

	return nil
}
