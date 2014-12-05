package models

import (
	"database/sql"
	"log"
	"testing"
)

func TestQMethod(t *testing.T) {
	q := _QMonad{
		method: Select,
		model:  &User{},
		fields: []string{"Name", "Password"},
		where:  "(Id = ? AND Name > ?) OR CreateTime = ?",
	}
	q.rowVisitor = func(r []sql.RawBytes) bool {
		log.Println(r)
		return true
	}
	if s, err := q.genSql(); err != nil {
		t.Errorf("genSql failed with error, %s", err)
	} else {
		log.Println(s)
	}

	q.method = Update
	if s, err := q.genSql(); err != nil {
		t.Errorf("genSql failed with error, %s", err)
	} else {
		log.Println(s)
	}

	q.method = Insert
	if s, err := q.genSql(); err != nil {
		t.Errorf("genSql failed with error, %s", err)
	} else {
		log.Println(s)
	}

	q.method = Delete
	if s, err := q.genSql(); err != nil {
		t.Errorf("genSql failed with error, %s", err)
	} else {
		log.Println(s)
	}
}

var _ = log.Println
