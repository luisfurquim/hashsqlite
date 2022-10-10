package hashsqlite

import (
	"errors"
   "github.com/gwenn/gosqlite"
   "github.com/luisfurquim/goose"
)

type At struct {
	Table interface{}
	With string
	By map[string]interface{}
}

type Schema struct {
	Tables []interface{}
	MxN map[string]string
}

type table struct {
	name string
	fields []field
	xrefs map[string]struct{}
	pkName string
	pkIndex int
}

type field struct {
	name string
	fk string
	joinList bool
	index int
}

type tabRule struct {
	table string
	rule string
}

type list struct {
//	tabName string
	cols []int
	joins map[int]tabRule
//	filterLen int
	stmt *sqlite.Stmt
}

type listSpec struct {
	cols           []string
	colTypes map[int]string
	sort           []string
	filter           string
}

type HashSqlite struct {
	tables map[string]table
	tableType map[string]string

	db *sqlite.Conn

	list map[string]map[string]*list

	insert map[string]*sqlite.Stmt
	link map[string]*sqlite.Stmt
	unlink map[string]*sqlite.Stmt
	updateBy map[string]map[string]*sqlite.Stmt
	count map[string]*sqlite.Stmt
	countBy map[string]map[string]*sqlite.Stmt
	listJoin map[string]map[string]*sqlite.Stmt
	listBy map[string]map[string]*sqlite.Stmt
	exists map[string]map[string]*sqlite.Stmt
	delete map[string]map[string]*sqlite.Stmt

}

type GooseG struct {
	Init goose.Alert
	Query goose.Alert
}

var Goose GooseG = GooseG{
	Init: goose.Alert(2),
	Query: goose.Alert(2),
}

var ErrSpecNotStruct         error = errors.New("Specification is not of struct type")
var ErrChanNotAllowed        error = errors.New("Channel not allowed")
var ErrNoTablesFound         error = errors.New("No tables found")
var ErrColumnNotFound        error = errors.New("Column not found")
var ErrRuleNotFound	        error = errors.New("Rule not found")
var ErrNotStructPointer      error = errors.New("Parameter must be of pointer to struct type")
var ErrNotStructSlicePointer error = errors.New("Parameter must be of pointer to slice of struct type")
var ErrNotStructPointerChan  error = errors.New("Parameter must be of channel of pointer to struct type")
var ErrNoRuleFound           error = errors.New("No rule found")
var ErrInvalid               error = errors.New("Invalid")
var ErrPKNotI64              error = errors.New("Primary key is not int64")
var ErrWrongParmCount        error = errors.New("Wrong parameter count")

