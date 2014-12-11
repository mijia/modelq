package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/mijia/modelq/gmq"
	"time"
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

func (u User) Update(db *sql.DB) (int64, error) {
	fields := []string{"Name", "Password", "IsMarried", "Age", "CreateTime"}
	filter := UserObjs.F_Id("=", u.Id)
	if result, err := UserObjs.Update(u, fields...).Where(filter).Exec(db); err != nil {
		return 0, err
	} else {
		return result.RowsAffected()
	}
}

func (u User) Delete(db *sql.DB) (int64, error) {
	filter := UserObjs.F_Id("=", u.Id)
	if result, err := UserObjs.Delete(u).Where(filter).Exec(db); err != nil {
		return 0, err
	} else {
		return result.RowsAffected()
	}
}

type _UserQuery struct {
	gmq.Query
}

type _UserObjs struct{}

func (o _UserObjs) Names() (string, string) { return "user", "User" }

func (o _UserObjs) Insert(u User) _UserQuery {
	q := _UserQuery{}
	q.Query = gmq.Insert(o, o.collectColumn(u, "Name", "Password", "IsMarried", "Age"))
	return q
}

func (o _UserObjs) Update(u User, fields ...string) _UserQuery {
	q := _UserQuery{}
	q.Query = gmq.Update(o, o.collectColumn(u, fields...))
	return q
}

func (o _UserObjs) Delete(u User) _UserQuery {
	q := _UserQuery{}
	q.Query = gmq.Delete(o)
	return q
}

func (o _UserObjs) F_Id(op string, p int64) gmq.Filter {
	return gmq.UnitFilter("id", op, p)
}

func (o _UserObjs) F_Name(op string, p string) gmq.Filter {
	return gmq.UnitFilter("name", op, p)
}

func (o _UserObjs) F_Password(op string, p string) gmq.Filter {
	return gmq.UnitFilter("password", op, p)
}

func (o _UserObjs) F_IsMarried(op string, p int) gmq.Filter {
	return gmq.UnitFilter("is_married", op, p)
}

func (o _UserObjs) F_Age(op string, p int) gmq.Filter {
	return gmq.UnitFilter("age", op, p)
}

func (o _UserObjs) F_CreateTime(op string, p time.Time) gmq.Filter {
	return gmq.UnitFilter("create_time", op, p)
}

func (o _UserObjs) F_UpdateTime(op string, p time.Time) gmq.Filter {
	return gmq.UnitFilter("update_time", op, p)
}

func (o _UserObjs) C_Id(p ...int64) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"id", value}
}

func (o _UserObjs) C_Name(p ...string) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"name", value}
}

func (o _UserObjs) C_Password(p ...string) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"password", value}
}

func (o _UserObjs) C_IsMarried(p ...int) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"is_married", value}
}

func (o _UserObjs) C_Age(p ...int) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"age", value}
}

func (o _UserObjs) C_CreateTime(p ...time.Time) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"create_time", value}
}

func (o _UserObjs) C_UpdateTime(p ...time.Time) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"update_time", value}
}

func (o _UserObjs) collectColumn(u User, fields ...string) []gmq.Column {
	data := make([]gmq.Column, 0, len(fields))
	for _, f := range fields {
		switch f {
		case "Id":
			data = append(data, o.C_Id(u.Id))
		case "Name":
			data = append(data, o.C_Name(u.Name))
		case "Password":
			data = append(data, o.C_Password(u.Password))
		case "IsMarried":
			data = append(data, o.C_IsMarried(u.IsMarried))
		case "Age":
			data = append(data, o.C_Age(u.Age))
		case "CreateTime":
			data = append(data, o.C_CreateTime(u.CreateTime))
		case "UpdateTime":
			data = append(data, o.C_UpdateTime(u.UpdateTime))
		}
	}
	return data
}

var UserObjs _UserObjs
