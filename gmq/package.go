package gmq

import (
	"database/sql"
	"errors"
)

var (
	ErrNotSupportedCall = errors.New("Such api cannot be called on this query, e.g. SelectOne on an InsertQuery.")
)

type WithinTxFunctor func(tx *sql.Tx) error

type Column struct {
	Name  string
	Value interface{}
}

type TableModel interface {
	Names() (string, string)
}

type Query interface {
	Exec(db *sql.DB) (sql.Result, error)
	// SelectOne(db *sql.DB) error
	// SelectList(db *sql.DB) error
	// Where(f Filter) Query
	// GroupBy(by string) Query
	// OrderBy(by string) Query
	// OrderByDesc(by string) Query
	// Limit(limit int, offsets ...int) Query
	// Page(number, size int) Query
}

func Insert(model TableModel, columns []Column) Query {
	q := _InsertQuery{}
	q.model = model
	q.columns = columns
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

var Debug bool

func init() {
	Debug = false
}
