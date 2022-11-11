package hashsqlite

import (
	"fmt"
	"strings"
	"reflect"
	"unicode"
	"github.com/gwenn/gosqlite"
)

func New(db *sqlite.Conn, schema Schema) (*HashSqlite, error) {
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
	var j int
	var k int
	var hs HashSqlite
	var fk string
	var xref string
	var xrefs map[string]struct{}
	var refTable string
	var err error
	var opt []string
	var pkName string
	var pkIndex int
	var tmpList map[string]listSpec
	var rule string
	var spec listSpec
	var stmt *sqlite.Stmt
	var cols []int
	var colNames string
	var colDef string
	var sortDef string
//	var filterLen int
	var sortSpec string
	var filterSpec string
	var parts []string
	var allJoins map[string]map[string]map[int]tabRule
	var flds map[int]tabRule
	var index int
	var fRule tabRule

	if len(schema.Tables) == 0 {
		Goose.Init.Logf(1,"Error: %s", ErrNoTablesFound)
		return nil, ErrNoTablesFound
	}

	hs.db		   = db
	hs.tables   = map[string]table{}
	hs.insert   = map[string]*sqlite.Stmt{}
	hs.link     = map[string]*sqlite.Stmt{}
	hs.updateBy = map[string]map[string]*sqlite.Stmt{}
	hs.count    = map[string]*sqlite.Stmt{}
	hs.countBy  = map[string]map[string]*sqlite.Stmt{}
	hs.list     = map[string]map[string]*list{}
	hs.listBy   = map[string]map[string]*sqlite.Stmt{}
	hs.exists   = map[string]map[string]*sqlite.Stmt{}
	hs.delete   = map[string]map[string]*sqlite.Stmt{}

	allJoins    = map[string]map[string]map[int]tabRule{}

	hs.tableType = make(map[string]string, len(schema.Tables))
	for _, tab = range schema.Tables {
		reftab = reflect.TypeOf(tab)
		if reftab.Kind() != reflect.Struct {
			Goose.Init.Logf(1,"Error on %s: %s", reftab.Name(), ErrSpecNotStruct)
			return nil, ErrSpecNotStruct
		}

		tabName = reftab.Name()
tableLoop1:
		for i=0; i<reftab.NumField(); i++ {
			f = reftab.Field(i)
			if len(f.Name)==0 {
				continue
			}

			for _, char = range f.Name {
				if !unicode.IsUpper(char) {
					continue tableLoop1
				}
				break
			}

			if fldName, ok = f.Tag.Lookup("table"); ok && len(fldName)>0 {
				tabName = fldName
				break
			}
		}

		hs.tableType[reftab.Name()] = tabName
		allJoins[tabName] = map[string]map[int]tabRule{}
	}

	Goose.Init.Logf(0,"%#v",hs.tableType)

	for _, tab = range schema.Tables {
		reftab = reflect.TypeOf(tab)
		tabName = reftab.Name()

		fldList = make([]field,0,reftab.NumField())
		xrefs = make(map[string]struct{},8)
		tmpList = map[string]listSpec{}

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
			} else if colNames, ok = f.Tag.Lookup("cols"); ok && len(colNames)>0 {
				sortSpec, _ = f.Tag.Lookup("sort")
				filterSpec, _ = f.Tag.Lookup("by")

				lSpec := listSpec{
					cols: strings.Split(colNames, ","),
					filter: filterSpec,
				}
				if len(sortSpec) > 0 {
					lSpec.sort = strings.Split(sortSpec, ",")
				}

				tmpList[f.Name] = lSpec

			} else {
				fld.joinList = false
				if f.Tag == "pk" {
					if f.Type.Name() != "int64" {
						Goose.Init.Logf(1,"Error for %s: %s", tabName, ErrPKNotI64)
						return nil, ErrPKNotI64
					}
					pkName = f.Name
					pkIndex = i
					continue
				}

				if fldName, ok = f.Tag.Lookup("field"); ok && len(fldName)>0 {
					opt = strings.Split(fldName, ",")
					fld.name = opt[0]
				} else {
					fld.name = f.Name
				}

				fld.fk = ""
				if f.Type.Kind() == reflect.Pointer {
					if fk, ok = hs.tableType[f.Type.Elem().Name()]; ok {
						fld.fk = fk
					} else {
						Goose.Init.Logf(0, "fld.name: %s => %s", fld.name, f.Type.Name())
					}
				} else if f.Type.Kind() == reflect.Slice {
					if xref, ok = hs.tableType[f.Type.Elem().Name()]; ok {
						xrefs[xref] = struct{}{}
					} else {
						Goose.Init.Logf(0, "*fld.name: %s => %s", fld.name, f.Type.Name())
					}
					fld.joinList = true
				} else {
					Goose.Init.Logf(0, "fld.name: %s => %s", fld.name, f.Type.Name())
				}

				fld.index = i

				fldList = append(fldList, fld)
			}
		}

		if len(fldList) > 0 {
			hs.tables[tabName] = table{
				name: tabName,
				fields: fldList,
				xrefs: xrefs,
				pkName: pkName,
				pkIndex: pkIndex,
			}

			hs.list[tabName] = map[string]*list{}

			colNames, cols = fieldJoin(fldList)

			err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s)`, tabName, colNames))
			if err != nil {
				Goose.Init.Logf(1,"Error creating %s table: %s", tabName, err)
				return nil, err
			}

			if len(pkName)>0 {
				colNames = "rowid," + colNames
				cols = append([]int{pkIndex}, cols...)
			}

			stmt, err = db.Prepare(fmt.Sprintf(`SELECT %s FROM %s ORDER BY rowid`, colNames, tabName))
			if err != nil {
				Goose.Init.Logf(1,"Err compiling list * from %s: %s", tabName, err)
				return nil, err
			}

			hs.list[tabName]["*"] = &list{
//				tabName: tabName,
				cols: cols,
				stmt: stmt,
			}

			if len(tmpList) > 0 {
				for rule, spec = range tmpList {
					cols = make([]int, len(spec.cols))
					colDef = ""
					for j=0; j<len(spec.cols); j++ {
						if (spec.cols[j] == pkName) || (spec.cols[j] == "rowid") {
							cols[j] = pkIndex
							if len(colDef) > 0 {
								colDef += ","
							}
							colDef += "rowid"
						} else {
							ok = false
							parts = strings.Split(spec.cols[j], ":")
							for k=0; k<len(fldList); k++ {
								if parts[0] == fldList[k].name {
									cols[j] = fldList[k].index
									if len(colDef) > 0 {
										colDef += ","
									}
									colDef += parts[0]
									ok = true
									break
								}
							}
							if !ok {
								Goose.Init.Logf(1,"Err compiling list %s from %s: %s", rule, tabName, ErrColumnNotFound)
								Goose.Init.Logf(1,"tmpList %#v col %s", tmpList, spec.cols[j])
								Goose.Init.Logf(1,"fldList %#v", fldList)
								return nil, ErrColumnNotFound
							}
							if len(parts) > 1 {
								if _, ok = allJoins[tabName][rule]; !ok {
									allJoins[tabName][rule] = map[int]tabRule{}
								}
								allJoins[tabName][rule][fldList[k].index] = tabRule{
									table: hs.tableType[reftab.Field(fldList[k].index).Type.Elem().Name()],
									rule:  parts[1],
								}
							}
						}
					}

					sortDef=""
					for j=0; j<len(spec.sort); j++ {
						if len(sortDef) > 0 {
							sortDef += ","
						}
						if spec.sort[j][0] == '>' {
							sortDef += spec.sort[j][1:] + " DESC"
						} else {
							sortDef += spec.sort[j]
						}
					}
					if len(sortDef) > 0 {
						sortDef = " ORDER BY " + sortDef
					}

					if len(spec.filter) > 0 {
						spec.filter = " WHERE " + spec.filter
					}
//					filterLen = len(strings.Split(spec.filter,"?")) - 1

					stmt, err = db.Prepare(fmt.Sprintf(`SELECT %s FROM %s%s%s`, colDef, tabName, spec.filter, sortDef))
					if err != nil {
						Goose.Init.Logf(1,"Err compiling list %s from %s: %s", rule, tabName, err)
						Goose.Init.Logf(1,"tmpList %#v", tmpList)
						Goose.Init.Logf(1,"fldList %#v", fldList)
						Goose.Init.Logf(1,"pkName %s, pkIndex: %s", pkName, pkIndex)
						return nil, err
					}

					hs.list[tabName][rule] = &list{
//						tabName: tabName,
						cols: cols,
//						filterLen: filterLen,
						stmt: stmt,
					}
				}
			}

			if len(pkName) > 0 {
				stmt, err = db.Prepare(fmt.Sprintf(`SELECT rowid FROM %s ORDER BY rowid`,  tabName))
				if err != nil {
					Goose.Init.Logf(1,"Err compiling list 0 from %s: %s", tabName, err)
					return nil, err
				}

				hs.list[tabName]["0"] = &list{
//					tabName: tabName,
					cols: []int{pkIndex},
					stmt: stmt,
				}

				colNames, cols = fieldJoin(fldList)
				stmt, err = db.Prepare(fmt.Sprintf(`SELECT rowid,` + colNames + ` FROM %s WHERE rowid=?`,  tabName))
				if err != nil {
					Goose.Init.Logf(1,"Err compiling select pk from %s: %s", tabName, err)
					return nil, err
				}

				hs.list[tabName]["id:*"] = &list{
//					tabName: tabName,
					cols: append([]int{pkIndex},cols...),
					stmt: stmt,
				}
			}

			Goose.Init.Logf(0,`INSERT INTO ` + tabName + ` VALUES (?` + strings.Repeat(",?",fieldLen(fldList)-1) + `)`)
			hs.insert[tabName], err = db.Prepare(`INSERT INTO ` + tabName + ` VALUES (?` + strings.Repeat(",?",fieldLen(fldList)-1) + `)`)
			if err != nil {
				Goose.Init.Logf(1,"Err compiling insert: %s (%#v)", err, fldList)
				return nil, err
			}

			hs.updateBy[tabName] = map[string]*sqlite.Stmt{}
			hs.updateBy[tabName]["id"], err = db.Prepare(`UPDATE ` + tabName + ` SET ` + fieldJoinNameVal(fldList) + ` WHERE rowid=?`)
			if err != nil {
				Goose.Init.Logf(1,"Err compiling updateBy: %s", err)
				return nil, err
			}

			hs.count[tabName], err = db.Prepare(`SELECT count(rowid) FROM ` + tabName)
			if err != nil {
				Goose.Init.Logf(1,"Err compiling count: %s", err)
				return nil, err
			}

			hs.countBy[tabName] = map[string]*sqlite.Stmt{}

/*
			hs.list[tabName], err = db.Prepare(`SELECT . FROM ` + tabName)
			if err != nil {
				Goose.Init.Logf(1,"Err compiling count: %s", err)
				return nil, err
			}

			hs.listBy[tabName] = map[string]*sqlite.Stmt{}
*/


			hs.exists[tabName] = map[string]*sqlite.Stmt{}
			hs.exists[tabName]["id"], err = db.Prepare(`SELECT count(rowid) FROM ` + tabName + ` WHERE rowid=?`)
			if err != nil {
				Goose.Init.Logf(1,"Err compiling exists: %s", err)
				return nil, err
			}

			hs.delete[tabName] = map[string]*sqlite.Stmt{}
			hs.delete[tabName]["id"], err = db.Prepare(`DELETE FROM ` + tabName + ` WHERE rowid=?`)
			if err != nil {
				Goose.Init.Logf(1,"Err compiling delete: %s", err)
				return nil, err
			}

		}
	}

	for tabName, _ = range hs.tables {
		for refTable, _ = range hs.tables[tabName].xrefs {
			if _, ok = hs.tables[refTable].xrefs[tabName]; !ok {
				continue
			}

			if _, ok = hs.link[refTable + "_" + tabName]; ok {
				continue
			}

			err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s_%s (id_%s, id_%s)`, tabName, refTable, tabName, refTable))
			if err != nil {
				Goose.Init.Logf(1,"Error creating %s table: %s", tabName, err)
				return nil, err
			}

			hs.link[tabName + "_" + refTable], err = db.Prepare(fmt.Sprintf(`INSERT INTO %s_%s VALUES (?,?)`, tabName, refTable))
			if err != nil {
				Goose.Init.Logf(1,"Err compiling link: %s", err)
				return nil, err
			}

			hs.unlink[tabName + "_" + refTable], err = db.Prepare(fmt.Sprintf(`DELETE FROM %s_%s WHERE id_%s=? AND id_%s?`, tabName, refTable, tabName, refTable))
			if err != nil {
				Goose.Init.Logf(1,"Err compiling unlink: %s", err)
				return nil, err
			}

			if len(hs.listJoin[tabName]) == 0 {
				hs.listJoin[tabName] = map[string]*sqlite.Stmt{}
			}
			hs.listJoin[tabName][refTable], err = db.Prepare(fmt.Sprintf(`SELECT id_%s FROM %s_%s WHERE id_%s=? `, refTable, tabName, refTable, tabName))
			if err != nil {
				Goose.Init.Logf(1,"Err compiling exists: %s", err)
				return nil, err
			}

			if len(hs.listJoin[refTable]) == 0 {
				hs.listJoin[refTable] = map[string]*sqlite.Stmt{}
			}
			hs.listJoin[refTable][tabName], err = db.Prepare(fmt.Sprintf(`SELECT id_%s FROM %s_%s WHERE id_%s=? `, tabName, tabName, refTable, refTable))
			if err != nil {
				Goose.Init.Logf(1,"Err compiling exists: %s", err)
				return nil, err
			}

		}

		for rule, flds = range allJoins[tabName] {
			if hs.list[tabName][rule].joins == nil {
				hs.list[tabName][rule].joins = map[int]tabRule{}
			}
			for index, fRule = range flds {
				hs.list[tabName][rule].joins[index] = fRule
			}
		}
	}

	return &hs, nil
}

