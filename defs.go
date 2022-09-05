package hashsqlite

import (
	"errors"
   "github.com/gwenn/gosqlite"
   "github.com/luisfurquim/goose"
)

type TabCondT struct {
	Tab, Cond string
}

type WhereT struct {
	Where map[string]*sqlite.Stmt
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

type list struct {
	cols []int
	filterLen int
	stmt *sqlite.Stmt
}

type listSpec struct {
	cols []string
	sort []string
	filter string
}

type HashSqlite struct {
	tables map[string]table
	tableType map[string]string

	db *sqlite.Conn

	insert map[string]*sqlite.Stmt
	link map[string]*sqlite.Stmt
	unlink map[string]*sqlite.Stmt
	updateBy map[string]map[string]*sqlite.Stmt
	count map[string]*sqlite.Stmt
	countBy map[string]map[string]*sqlite.Stmt
	list map[string]map[string]list
	listJoin map[string]map[string]*sqlite.Stmt
	listBy map[string]map[string]*sqlite.Stmt
	exists map[string]map[string]*sqlite.Stmt
	delete map[string]map[string]*sqlite.Stmt

//   Find map[string]WhereT `json:"-"` // Find(&T1{}, Join{}, Where{})
//   FindJoined map[string]WhereT `json:"-"`
//   List map[string]*sqlite.Stmt `json:"-"`
//   ListJoined map[string]WhereT `json:"-"`
}

type GooseG struct {
	Init goose.Alert
	Query goose.Alert
}

var Goose GooseG = GooseG{
	Init: goose.Alert(2),
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

