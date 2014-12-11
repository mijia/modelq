package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/mijia/modelq/gmq"
)

func (u User) String() string {
	if data, err := json.Marshal(u); err != nil {
		return fmt.Sprintf("<User id=%d>", u.Id)
	} else {
		return string(data)
	}
}

func (u User) Insert(db *sql.DB) (User, error) {
	if result, err := UserObjs.Insert(u).Exec(db); err != nil {
		return u, err
	} else {
		if id, err := result.LastInsertId(); err != nil {
			return u, err
		} else {
			u.Id = id
			return u, nil
		} 
	}
}

type _UserQuery struct {
	gmq.Query
}

type _UserObjs struct {}

func (o _UserObjs) Names() (string, string) { return "user", "User" }

func (o _UserObjs) Insert(u User) _UserQuery {
	q := _UserQuery{}
	q.Query = gmq.Insert(o, o.collectColumn(u, "Name", "Password", "IsMarried", "Age"))
	return q
}

func (o _UserObjs) collectColumn(u User, fields ...string) []gmq.Column {
	data := make([]gmq.Column, 0, len(fields))
	for _, f := range fields {
		switch f {
		case "Id":
			data = append(data, gmq.Column{"id", u.Id})
		case "Name":
			data = append(data, gmq.Column{"name", u.Name})
		case "Password":
			data = append(data, gmq.Column{"password", u.Password})
		case "IsMarried":
			data = append(data, gmq.Column{"is_married", u.IsMarried})
		case "Age":
			data = append(data, gmq.Column{"age", u.Age})
		case "CreateTime":
			data = append(data, gmq.Column{"create_time", u.CreateTime})
		case "UpdateTime":
			data = append(data, gmq.Column{"update_time", u.UpdateTime})
		}
	}
	return data
}

var UserObjs _UserObjs

func init() {
	UserObjs = _UserObjs{}
}