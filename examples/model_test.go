package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mijia/modelq/gmq"
	"github.com/mijia/modelq/models"
	"log"
	"testing"
)

var db *sql.DB
var _ = log.Println

/*
func TestModelBatchApi(t *testing.T) {
	var err error
	objs := models.UserObjs
	objs.Select().Filter(objs.FilterId("=", 1).And(objs.FilterName("LIKE", "jia%"))).One(db)
	objs.Select().Filter(objs.FilterName("LIKE", "jia%")).OrderBy("CreateTime").Page(1, 20).List(db)
	objs.Select("Id", "Name").Filter(objs.FilterName("LIKE", "jia%")).OrderByDesc("CreateTime").Page(1, 20).List(db)
	objs.Select("Age").GroupBy("Age").List(db)
	// also we should have an iterate api

	models.WithinTx(func(tx *sql.Tx) error {
		data := models.User{Age: 12, IsMarried: 0}
		objs.Update(data, "Age", "IsMarried").Filter(objs.FilterAge("=", 11)).ExecWithinTx(tx)
		objs.Delete().Filter(objs.FilterAge("=", 12)).ExecWithinTx(tx)
	})
}
*/

func TestModelInstanceApi(t *testing.T) {
	var err error
	user := models.User{}
	user.Name = "mijia"
	user.Password = "test12345"
	user.Age = 15

	if user, err = user.Insert(db); err != nil || user.Id == 0 {
		t.Errorf("Insert is not working, %s", err)
	}

	user.Age = 36
	user.IsMarried = 1
	if affected, err := user.Update(db); err != nil || affected == 0 {
		t.Errorf("Update is not working, %s", err)
	}

	if affected, err := user.Delete(db); err != nil || affected == 0 {
		t.Errorf("Delete is not working, %s", err)
	}
}

func init() {
	var err error
	db, err = sql.Open("mysql", "root@/blog")
	if err != nil {
		panic(err)
	}

	gmq.Debug = true
}
