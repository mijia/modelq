package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNotValidWhereClause = errors.New("Some unknown operand or fields found in where clause.")
	ErrNoSuchColumn        = errors.New("Quering on some unknown column.")
	ErrNoSuchMethod        = errors.New("Q method found besides Select, Insert, Delete, Update.")
	ErrMultipleRowReturned = errors.New("Multiple row returned, but suppose there is only one row.")
)

type _TransactionFunc func(tx *sql.Tx) error
type _RowVisitorFunc func([]sql.RawBytes) bool

type _TableModel interface {
	TableName() string
	TableColumns() map[string]string
	ModelFields() []string
}

type _Method int

const (
	Select _Method = iota
	Insert
	Update
	Delete
)

type _QMonad struct {
	method     _Method
	model      _TableModel
	fields     []string
	where      string
	rowVisitor _RowVisitorFunc
}

func (q _QMonad) exec(db *sql.DB, params ...interface{}) (sql.Result, error) {
	if query, err := q.genSql(); err != nil {
		return nil, err
	} else {
		return dbExec(db, query, params)
	}
}

func (q _QMonad) queryOne(db *sql.DB, params ...interface{}) error {
	if query, err := q.genSql(); err != nil {
		return err
	} else {
		rowCount := 0
		var err error
		err = dbQuery(db, query, params, func(r []sql.RawBytes) bool {
			rowCount++
			if rowCount >= 2 {
				err = ErrMultipleRowReturned
				return false
			} else {
				return q.rowVisitor(r)
			}
		})
		if rowCount == 0 {
			return sql.ErrNoRows
		}
		return err
	}
}

func (q _QMonad) query(db *sql.DB, params ...interface{}) error {
	if query, err := q.genSql(); err != nil {
		return err
	} else {
		return dbQuery(db, query, params, q.rowVisitor)
	}
}

func (q _QMonad) genSql() (string, error) {
	mainClause := ""
	fields := make([]string, len(q.fields))
	for i, f := range q.fields {
		if _, ok := q.model.TableColumns()[f]; !ok {
			return "", ErrNoSuchColumn
		}
		fields[i] = quoteName(q.model.TableColumns()[f])
	}
	table := quoteName(q.model.TableName())
	switch q.method {
	case Select:
		fConcat := strings.Join(fields, ", ")
		mainClause = fmt.Sprintf("SELECT %s FROM %s", fConcat, table)
	case Insert:
		fConcat := strings.Join(fields, ", ")
		qMarks := make([]string, len(fields))
		for i := range qMarks {
			qMarks[i] = "?"
		}
		mainClause = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, fConcat, strings.Join(qMarks, ", "))
	case Update:
		for i, f := range fields {
			fields[i] = fmt.Sprintf("%s = ?", f)
		}
		fConcat := strings.Join(fields, ", ")
		mainClause = fmt.Sprintf("UPDATE %s SET %s", table, fConcat)
	case Delete:
		mainClause = fmt.Sprintf("DELETE FROM %s", table)
	default:
		return "", ErrNoSuchMethod
	}

	if whereClause, err := q.parseWhere(); err != nil {
		return "", err
	} else if whereClause != "" {
		mainClause = fmt.Sprintf("%s %s", mainClause, whereClause)
	}

	return mainClause, nil
}

func (q _QMonad) parseWhere() (string, error) {
	if q.where == "" {
		return "", nil
	}
	data := append([]byte{}, []byte("WHERE ")...)
	var symbol []byte
	for i := 0; i < len(q.where); i++ {
		ch := q.where[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			symbol = append(symbol, ch)
		} else {
			if len(symbol) > 0 {
				term := string(symbol)
				if strings.ToUpper(term) == "AND" || strings.ToUpper(term) == "OR" {
					data = append(data, []byte(strings.ToUpper(term))...)
				} else {
					if f, ok := q.model.TableColumns()[term]; !ok {
						return "", ErrNotValidWhereClause
					} else {
						data = append(data, []byte(quoteName(f))...)
					}
				}
				symbol = nil
			}
			data = append(data, ch)
		}
	}
	return string(data), nil
}

func quoteName(n string) string {
	return fmt.Sprintf("`%s`", n)
}

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

func dbExec(db *sql.DB, q string, params []interface{}) (sql.Result, error) {
	var result sql.Result
	err := WithinTx(db, func(tx *sql.Tx) error {
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

func WithinTx(db *sql.DB, fn _TransactionFunc) error {
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

var _ = fmt.Println
