package main

var tmplDbutils string = `package %s

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var _ = fmt.Println

var (
	ErrMultipleRowReturned = errors.New("Multiple row returned, but suppose there is only one row.")
)

type _TransactionFunc func(tx *sql.Tx) error
type _RowVisitorFunc func([]sql.RawBytes) bool

func dbQuery(db *sql.DB, q string, params []interface{}, visitor _RowVisitorFunc) error {
	if rows, err := db.Query(q, params...); err != nil {
		return err
	} else {
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return err
		}
		vals := make([]sql.RawBytes, len(cols))
		ints := make([]interface{}, len(cols))
		for i := range ints {
			ints[i] = &vals[i]
		}
		for rows.Next() {
			if err := rows.Scan(ints...); err != nil {
				return err
			}
			if continued := visitor(vals); !continued {
				break
			}
		}
	}
	return nil
}

func dbExec(db *sql.DB, q string, params ...interface{}) (sql.Result, error) {
	var result sql.Result
	err := WithinSession(db, func(tx *sql.Tx) error {
		if stmt, err := tx.Prepare(q); err != nil {
			return err
		} else {
			result, err = stmt.Exec(params...)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return result, err
}

func WithinSession(db *sql.DB, fn _TransactionFunc) error {
	if tx, err := db.Begin(); err != nil {
		return err
	} else {
		err := fn(tx)
		if err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				return rErr
			} else {
				return err
			}
		} else {
			return tx.Commit()
		}
	}
}

func asString(rb sql.RawBytes) string {
	if len(rb) > 0 {
		return string(rb)
	}
	return ""
}

func asInt(rb sql.RawBytes) int {
	return int(asInt64(rb))
}

func asInt64(rb sql.RawBytes) int64 {
	if len(rb) > 0 {
		if n, err := strconv.ParseInt(string(rb), 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func asTime(rb sql.RawBytes) time.Time {
	if t, err := time.Parse("2006-01-02 15:04:05", string(rb)); err == nil {
		return t
	}
	return time.Now()
}
`
