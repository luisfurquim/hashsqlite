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

type table struct {
	name string
	fields []field
}

type field struct {
	name string
	foreign string
}

type HashSqlite struct {
	tables map[string]table
   Find map[string]WhereT `json:"-"`
   FindJoined map[string]WhereT `json:"-"`
   List map[string]*sqlite.Stmt `json:"-"`
   ListJoined map[string]WhereT `json:"-"`
   Add map[string]*sqlite.Stmt `json:"-"`
//   Check map[string]WhereT `json:"-"`
}

type GooseG struct {
	Init goose.Alert
}

var Goose GooseG = GooseG{
	Init: goose.Alert(2),
}

var ErrSpecNotStruct error = errors.New("Specification is not of struct type")


