package hashsqlite

import (
	"reflect"
   "github.com/gwenn/gosqlite"
)

func (hs *HashSqlite) List(tab interface{}, ruleAndFilter ...interface{}) error {
	var tabName string
	var refRow reflect.Value
	var err error
	var ok bool
	var isChan bool
	var rule string
	var r map[string]list
	var l list

	tabName, refRow, isChan, err = hs.getMultiRefs(tab)
	if err != nil {
		Goose.Query.Logf(1, "Parameter type error: %s", ErrNoTablesFound)
		return ErrNoTablesFound
	}

	if len(ruleAndFilter) == 0 {
		rule = "0"
	} else {
		if rule, ok = ruleAndFilter[0].(string); !ok {
			Goose.Query.Logf(1, "Parameter type error: %s", ErrNoRuleFound)
			return ErrNoRuleFound
		}
		ruleAndFilter = ruleAndFilter[1:]
	}

	if r, ok = hs.list[tabName]; !ok {
		Goose.Query.Logf(1, "Error listing table %s: %s", tabName, ErrNoTablesFound)
		return ErrNoTablesFound
	}

	if l, ok = r[rule]; !ok {
		Goose.Query.Logf(1, "Error listing table %s: %s", tabName, ErrRuleNotFound)
		Goose.Query.Logf(1, "rule %s: rules %#v", rule, hs.list)
		return ErrRuleNotFound
	}

	if l.filterLen != len(ruleAndFilter) {
		Goose.Query.Logf(1, "Error listing table %s: %s", tabName, ErrWrongParmCount)
		return ErrWrongParmCount
	}

	err = l.stmt.Bind(ruleAndFilter...)
	if err != nil {
		Goose.Query.Logf(1, "Error binding on list of table %s: %s", tabName, err)
		return err
	}

	if isChan {
		go func() {
			err = l.stmt.Select(func(s *sqlite.Stmt) error {
				var parms []interface{}
				var row reflect.Value
				var i, c int

				row = reflect.New(refRow.Type().Elem().Elem())
				parms = make([]interface{}, len(l.cols))

				for i, c = range l.cols {
					parms[i] = row.Field(c).Addr().Interface()
				}

				err = s.Scan(parms...)
				if err != nil {
					Goose.Query.Logf(1, "Error scanning on list of table %s: %s", tabName, err)
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
			var parms []interface{}
			var row reflect.Value
			var i, c int

			row = reflect.New(refRow.Type().Elem()).Elem()
			parms = make([]interface{}, len(l.cols))

			for i, c = range l.cols {
				parms[i] = row.Field(c).Addr().Interface()
			}

			err = s.Scan(parms...)
			if err != nil {
				Goose.Query.Logf(1, "Error scanning on list of table %s: %s", tabName, err)
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
