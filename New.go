package hashsqlite

import (
	"fmt"
	"reflect"
	"unicode"
	"github.com/gwenn/gosqlite"
)

func New(db *sqlite.Conn, tabs ...interface{}) (*HashSqlite, error) {
	var tab interface{}
	var reftab reflect.Type
	var tabName string
	var f reflect.StructField
	var fld field
	var fldName string
	var fldList []field
	var ok bool
	var char rune
	var i int
	var hs HashSqlite
	var foreign string
	var err error

	hs.tables = map[string]table{}

	for _, tab = range tabs {
		reftab = reflect.TypeOf(tab)
		if reftab.Kind() != reflect.Struct {
			Goose.Init.Logf(1,"Error on %s: %s", reftab.Name(), ErrSpecNotStruct)
			return nil, ErrSpecNotStruct
		}

		tabName = reftab.Name()

		fldList = make([]field,0,reftab.NumField())
tableLoop:
		for i=0; i<reftab.NumField(); i++ {
			f = reftab.Field(i)
			if len(f.Name)==0 {
				continue
			}

			for _, char = range f.Name {
				if !unicode.IsUpper(char) {
					continue tableLoop
				}
				break
			}

			if fldName, ok = f.Tag.Lookup("table"); ok && len(fldName)>0 {
				tabName = fldName
			} else {
				if fldName, ok = f.Tag.Lookup("name"); ok && len(fldName)>0 {
					fld.name = "`" + fldName + "`"
				} else {
					fld.name = f.Name
				}

				if foreign, ok = f.Tag.Lookup("foreign"); ok && len(foreign)>0 {
					fld.foreign = "`" + foreign + "`"
				} else {
					fld.foreign = ""
				}

				fldList = append(fldList, fld)
			}
		}

		if len(fldList) > 0 {
			hs.tables[tabName] = table{
				name: tabName,
				fields: fldList,
			}
			err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s)`, tabName, fieldJoin(fldList)))
			if err != nil {
				Goose.Init.Logf(1,"Error creating %s table: %s", tabName, err)
				return nil, err
			}
		}
	}

	return &hs, nil
}

