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

type Db struct {
	*sql.DB
	driverName string
}

type Tx struct {
	*sql.Tx
	driverName string
}

func (tx *Tx) DriverName() string {
	return tx.driverName
}

func (db *Db) DriverName() string {
	return db.driverName
}

func (db *Db) Beginx() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, driverName: db.driverName}, err
}

func NewDb(db *sql.DB, driverName string) *Db {
	return &Db{
		DB:         db,
		driverName: driverName,
	}
}

func Open(driverName, dataSourceName string) (*Db, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return NewDb(db, driverName), err
}

type WithinTxFunctor func(tx *Tx) error
type QueryRowVisitor func(columns []Column, rb []sql.RawBytes) bool

type Column struct {
	Name  string
	Value interface{}
}

type TableModel interface {
	Names() (schema, tbl, alias string)
}

type DbTx interface {
	DriverName() string
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Query interface {
	String() string
	Exec(dbtx DbTx) (sql.Result, error)
	SelectOne(dbtx DbTx, functor QueryRowVisitor) error
	SelectList(dbtx DbTx, functor QueryRowVisitor) error
	SelectCount(dbtx DbTx, functor QueryRowVisitor) error
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

func WithinTx(db *Db, functor WithinTxFunctor) error {
	if tx, err := db.Beginx(); err != nil {
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

func AsBool(rb sql.RawBytes) bool {
	if len(rb) > 0 {
		if b, err := strconv.ParseBool(string(rb)); err == nil {
			return b
		}
	}
	return false
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
	} else if t, err := time.Parse("2006-01-02", string(rb)); err == nil {
		return t
	} else if t, err := time.Parse("15:04:05", string(rb)); err == nil {
		return t
	}

	return time.Now()
}

func AsByteArray(rb sql.RawBytes) []byte {
       return []byte(rb)
}

var Debug bool

func init() {
	Debug = false
}
