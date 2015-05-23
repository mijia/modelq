package examples

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	models "github.com/mijia/modelq/examples/mysql"
	"github.com/mijia/modelq/gmq"
)

var db *gmq.Db

func TestModelIterator(t *testing.T) {
	objs := models.UserObjs
	err := objs.Select().Where(objs.FilterAge(">=", 36)).OrderBy("-Age").
		Iterate(db, func(u models.User) bool {
		if u.Age < 40 {
			fmt.Println(u.Name, u.Age)
		}
		return true
	})
	if err != nil {
		t.Errorf("Iterateor is not working, %s", err)
	}
}

func TestModelBatchApi(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	err := gmq.WithinTx(db, func(tx *gmq.Tx) error {
		for i := 0; i < 5; i++ {
			user := models.User{}
			user.Name = fmt.Sprintf("mijia_%d_%d", time.Now().UnixNano(), rand.Int63())
			user.Password = "123456789"
			user.Age = rand.Intn(120) + 1
			if _, err := user.Insert(tx); err != nil {
				t.Errorf("Failed to insert test data for batch query, %s", user)
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Errorf("Failed to insert test data in transaction, %s", err)
	}

	objs := models.UserObjs
	query := objs.Select().Where(objs.FilterName("LIKE", "mijia%"))
	if _, err := query.List(db); err != nil {
		t.Errorf("Failed to list query, %s", err)
	}

	query = objs.Select("Id", "Age").Where(objs.FilterAge(">=", 10)).
		OrderBy("Age", "+Id").Page(9, 5)
	if _, err := query.List(db); err != nil {
		t.Errorf("Failed to list query, %s", err)
	}

	query = objs.Select("Age", "Id").Where(objs.FilterAge("IN", 12, 13, 14)).
		GroupBy("Age")
	if _, err := query.List(db); err != nil {
		t.Errorf("Failed to list query, %s", err)
	}

	data := models.User{Age: 19, IsMarried: 1}
	query = objs.Update(data, "Age", "IsMarried").Where(objs.FilterAge("=", data.Age-1))
	if _, err := query.Run(db); err != nil {
		t.Errorf("Failed to do batch update, %s", err)
	}

	query = objs.Delete().Where(objs.FilterAge(">", 70))
	if _, err := query.Run(db); err != nil {
		t.Errorf("Failed to do batch delete, %s", err)
	}
}

func TestModelInstanceApi(t *testing.T) {
	var err error
	user := models.User{}
	user.Name = "mijia"
	user.Password = "test12345"
	user.Age = 15

	if user, err = user.Insert(db); err != nil || user.Id == 0 {
		t.Errorf("Insert is not working, %v", err)
	}

	userId := user.Id
	objs := models.UserObjs
	query := objs.Select().Where(objs.FilterId("=", userId))

	if user, err = query.One(db); err != nil {
		t.Errorf("Select one is not working, %v", err)
	}

	user.Age = 36
	user.IsMarried = 1
	if affected, err := user.Update(db); err != nil || affected == 0 {
		t.Errorf("Update is not working, %v", err)
	}
	if user, err = query.One(db); err != nil {
		t.Errorf("Select one is not working, %v", err)
	}

	article := models.Article{
		UserId: user.Id,
		Title:  "Hello World",
	}
	if article, err = article.Insert(db); err != nil || article.Id == 0 {
		t.Errorf("Insert is not working for article, %v", err)
	}

	comment := models.Comment{
		UserId:    user.Id,
		ArticleId: article.Id,
	}
	if comment, err = comment.Insert(db); err != nil {
		t.Errorf("Fail to insert a comment, %v", err)
	}
	comment.Content = "Woow"
	if affected, err := comment.Update(db); err != nil || affected == 0 {
		t.Errorf("Fail to update a comment, %s", err)
	}

	if affected, err := user.Delete(db); err != nil || affected == 0 {
		t.Errorf("Delete is not working, %v", err)
	}
}

func init() {
	var err error
	db, err = gmq.Open("mysql", "root@/blog")
	if err != nil {
		panic(err)
	}
	gmq.Debug = true
}
