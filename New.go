package hashsqlite

import (
	"fmt"
	"strings"
	"reflect"
	"unicode"
	"github.com/gwenn/gosqlite"
)

func New(db *sqlite.Conn, tabs ...interface{}) (*HashSqlite, error) {
	var tab interface{}
	var reftab reflect.Type
	var tabName string
	var field reflect.StructField
	var fldName string
	var fldList []string
	var ok bool
	var char rune
	var i int
	var hs HashSqlite

	for _, tab = range tabs {
		reftab = reflect.TypeOf(tab)
		if reftab.Kind() != reflect.Struct {
			Goose.Spec.Logf(1,"Error on %s: %s", reftab.Name(), ErrSpecNotStruct)
			return nil, ErrSpecNotStruct
		}

		tabName = reftab.Name()

		fldList = make([]string,0,reftab.NumField())
tableLoop:
		for i=0; i<reftab.NumField(); i++ {
			field = reftab.Field(i)
			if len(field.Name)==0 {
				continue
			}

			for _, char = range field.Name {
				if !unicode.IsUpper(char) {
					continue tableLoop
				}
				break
			}

			if fldName, ok = field.Tag.Lookup("table"); ok && len(fldName)>0 {
				tabName = fldName
			} else if fldName, ok = field.Tag.Lookup("name"); ok && len(fldName)>0 {
				fldList = append(fldList, "`" + fldName + "`")
			} else {
				fldList = append(fldList, field.Name)
			}
		}

		if len(fldList) > 0 {
			fmt.Printf(`CREATE TABLE IF NOT EXISTS %s (%s)`, tabName, strings.Join(fldList, ","))
/*
			err = ds.db.Exec(`CREATE TABLE IF NOT EXISTS dataset (id_user, locator, Name, Organizer, Comments, DateCreated, LastUpdate, Stage)`)
			if err != nil {
				Goose.Logf(0,"Err creating research table: %s", err)
				return
			}
*/
		}
	}

	return &hs, nil
}

