package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mijia/modelq/models"
	"log"
	"testing"
)

var db *sql.DB
var _ = log.Println

func TestUserModelCRUD(t *testing.T) {
	user := models.NewUser()
	user.Name = "mijia"
	user.Password = "test12345"

	if err := user.Insert(db, true); err != nil {
		t.Errorf("Insert is not working, %s", err)
	}
	log.Println(user)

	user.Password = "9687"
	if count, err := user.Update(db, true); err != nil || count != 1 {
		t.Errorf("Update is not working, rows affected %d, %s", count, err)
	} else if user.Password != "9687" {
		t.Errorf("Update failed, the field has not changed!, expect 9687, but got %s", user.Password)
	}
	log.Println(user)

	newUser := &models.User{}
	if err := newUser.Get(db, user.Id); err != nil {
		t.Errorf("Get failed, got err, %s", err)
	}
	log.Println(newUser)

	if c, err := newUser.Delete(db); err != nil || c != 1 {
		t.Errorf("Delete failed %s, rows affected %d", err, c)
	}
	if err := newUser.Get(db, newUser.Id); err == nil {
		t.Errorf("The record should be deleted, but we got no error for fetching it")
	}
}

func init() {
	var err error
	db, err = sql.Open("mysql", "root@/blog")
	if err != nil {
		panic(err)
	}
}
