package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (user *User) updateFieldsByRawBytes(fields []string, rb []sql.RawBytes) {
	for i := 0; i < len(fields) && i < len(rb); i++ {
		switch fields[i] {
		case "Id":
			user.Id = asInt64(rb[i])
		case "Name":
			user.Name = asString(rb[i])
		case "Password":
			user.Password = asString(rb[i])
		case "CreateTime":
			user.CreateTime = asTime(rb[i])
		case "UpdateTime":
			user.UpdateTime = asTime(rb[i])
		}
	}
}

func (user *User) TableName() string {
	return "user"
}

func (user *User) TableColumns() map[string]string {
	return map[string]string{
		"Id":         "id",
		"Name":       "name",
		"Password":   "password",
		"CreateTime": "create_time",
		"UpdateTime": "update_time",
	}
}

func (user *User) ModelFields() []string {
	return []string{"Id", "Name", "Password", "CreateTime", "UpdateTime"}
}

func (user *User) String() string {
	if data, err := json.Marshal(user); err != nil {
		return fmt.Sprintf("<User id=%d>", user.Id)
	} else {
		return string(data)
	}
}

func (user *User) _Q(method _Method, fields []string, where string) _QMonad {
	return _QMonad{
		model:  user,
		method: method,
		fields: fields,
		where:  where,
	}
}

func (user *User) Get(db *sql.DB, id int64) (err error) {
	q := user._Q(Select, user.ModelFields(), "Id = ?")
	q.rowVisitor = func(rb []sql.RawBytes) bool {
		user.updateFieldsByRawBytes(user.ModelFields(), rb)
		return true
	}
	return q.queryOne(db, id)
}

func (user *User) Insert(db *sql.DB, updateSelf bool) error {
	// Omit the `id` since it is auto_increment
	// Omit the create_time, update_time since it is DATETIME field and has a default CURRENT_TIMESTAMP
	q := user._Q(Insert, []string{"Name", "Password"}, "")
	result, err := q.exec(db, user.Name, user.Password)
	if err != nil {
		return err
	}
	if updateSelf {
		if lastId, err := result.LastInsertId(); err != nil {
			return errors.New(fmt.Sprintf("Updated self failed with error, %s", err))
		} else {
			return user.Get(db, lastId)
		}
	}
	return nil
}

func (user *User) Update(db *sql.DB, updateSelf bool) (int64, error) {
	q := user._Q(Update, []string{"Name", "Password", "CreateTime"}, "Id = ?")
	result, err := q.exec(db, user.Name, user.Password, user.CreateTime, user.Id)
	if err != nil {
		return 0, err
	}
	if updateSelf {
		err := user.Get(db, user.Id)
		if err != nil {
			return 0, err
		}
	}
	return result.RowsAffected()
}

func (user *User) Delete(db *sql.DB) (int64, error) {
	q := user._Q(Delete, []string{}, "Id = ?")
	if result, err := q.exec(db, user.Id); err != nil {
		return 0, err
	} else {
		return result.RowsAffected()
	}
}

func NewUser() *User {
	return &User{
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
}
