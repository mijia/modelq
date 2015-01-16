package examples

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	_ "github.com/lib/pq"
	models "github.com/mijia/modelq/examples/postgres"
	"github.com/mijia/modelq/gmq"
)

var pqdb *gmq.Db

func TestPqModelIterator(t *testing.T) {
	objs := models.UserObjs
	err := objs.Select().Where(objs.FilterAge(">=", 36)).OrderBy("-Age").
		Iterate(pqdb, func(u models.User) bool {
		if u.Age < 50 {
			fmt.Println(u.Name, u.Age)
		}
		return true
	})
	if err != nil {
		t.Errorf("Iterateor is not working, %s", err)
	}
}

func TestPqModelBatchApi(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	err := gmq.WithinTx(pqdb, func(tx *gmq.Tx) error {
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
	query := objs.Select().Where(objs.FilterAge(">=", 5))
	if _, err := query.List(pqdb); err != nil {
		t.Errorf("Failed to list query, %s", err)
	}

	query = objs.Select("Id", "Age").Where(objs.FilterAge(">=", 10)).
		OrderBy("Age", "+Id").Page(9, 5)
	if _, err := query.List(pqdb); err != nil {
		t.Errorf("Failed to list query, %s", err)
	}

	query = objs.Select("Age").Where(objs.FilterAge("IN", 12, 13, 14)).
		GroupBy("Age")
	if _, err := query.List(pqdb); err != nil {
		t.Errorf("Failed to list query, %s", err)
	}

	data := models.User{Age: 19, IsMarried: true}
	query = objs.Update(data, "Age", "IsMarried").Where(objs.FilterAge("=", data.Age-1))
	if _, err := query.Run(pqdb); err != nil {
		t.Errorf("Failed to do batch update, %s", err)
	}

	query = objs.Delete().Where(objs.FilterAge(">", 70))
	if _, err := query.Run(pqdb); err != nil {
		t.Errorf("Failed to do batch delete, %s", err)
	}
}

func TestPqModelInstanceApi(t *testing.T) {
	var err error
	user := models.User{}
	user.Name = "mijia"
	user.Password = "test12345"
	user.Age = 15

	if user, err = user.Insert(pqdb); err != nil {
		t.Errorf("Insert is not working, %v", err)
	} else {
		log.Println(user)
	}

	objs := models.UserObjs
	query := objs.Select().Where(objs.FilterName("=", user.Name))

	if user, err = query.One(pqdb); err != nil {
		t.Errorf("Select one is not working, %v", err)
	}

	user.Age = 36
	user.IsMarried = true
	if _, err := user.Update(pqdb); err != nil {
		t.Errorf("Update is not working, %v", err)
	}
	if user, err = query.One(pqdb); err != nil {
		t.Errorf("Select one is not working, %v", err)
	}

	if _, err := user.Delete(pqdb); err != nil {
		t.Errorf("Delete is not working, %v", err)
	}
}

func init() {
	var err error
	pqdb, err = gmq.Open("postgres", "dbname=blog sslmode=disable")
	if err != nil {
		panic(err)
	}
	gmq.Debug = true
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
