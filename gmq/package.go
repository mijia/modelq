package gmq

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

var (
	ErrNoPrimaryKeyDefined = errors.New("Cannot call this, because there is no primary key defined for the model.")
	ErrNotSupportedCall    = errors.New("Such api cannot be called on this query, e.g. SelectOne on an InsertQuery.")
	ErrNotEnoughColumns    = errors.New("Not enough columns data for Insert/Update.")
	ErrMultipleRowReturned = errors.New("Multiple row returned, but suppose there is only one row.")
	ErrNotDbTxObject       = errors.New("This is not a valid database/sql.Db or sql.Tx")
)

type WithinTxFunctor func(tx *sql.Tx) error
type QueryRowVisitor func(columns []Column, rb []sql.RawBytes) bool

type Column struct {
	Name  string
	Value interface{}
}

type TableModel interface {
	Names() (string, string)
}

type DbTx interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Query interface {
	String() string
	Exec(dbtx DbTx) (sql.Result, error)
	SelectOne(dbtx DbTx, functor QueryRowVisitor) error
	SelectList(dbtx DbTx, functor QueryRowVisitor) error
	Where(f Filter) Query
	OrderBy(by ...string) Query
	Limit(offsets ...int64) Query
	Page(number, size int) Query
	GroupBy(by ...string) Query
}

func Select(model TableModel, columns []Column) Query {
	q := _SelectQuery{}
	q.model = model
	q.columns = columns
	return q
}

func Insert(model TableModel, columns []Column) Query {
	q := _InsertQuery{}
	q.model = model
	q.columns = columns
	return q
}

func Update(model TableModel, columns []Column) Query {
	q := _UpdateQuery{}
	q.model = model
	q.columns = columns
	return q
}

func Delete(model TableModel) Query {
	q := _DeleteQuery{}
	q.model = model
	return q
}

func WithinTx(db *sql.DB, functor WithinTxFunctor) error {
	if tx, err := db.Begin(); err != nil {
		return err
	} else {
		err := functor(tx)
		if err != nil {
			tx.Rollback()
			return err
		} else {
			return tx.Commit()
		}
	}
}

func AsString(rb sql.RawBytes) string {
	if len(rb) > 0 {
		return string(rb)
	}
	return ""
}

func AsInt(rb sql.RawBytes) int {
	return int(AsInt64(rb))
}

func AsInt64(rb sql.RawBytes) int64 {
	if len(rb) > 0 {
		if n, err := strconv.ParseInt(string(rb), 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func AsFloat64(rb sql.RawBytes) float64 {
	if len(rb) > 0 {
		if n, err := strconv.ParseFloat(string(rb), 64); err == nil {
			return n
		}
	}
	return 0
}

func AsTime(rb sql.RawBytes) time.Time {
	if t, err := time.Parse("2006-01-02 15:04:05", string(rb)); err == nil {
		return t
	}
	return time.Now()
}

var Debug bool

func init() {
	Debug = false
}
