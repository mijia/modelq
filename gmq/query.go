package gmq

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type _Columns []Column

func (c _Columns) fieldsAndParams(alias string) ([]string, []interface{}) {
	fields := make([]string, len(c))
	params := make([]interface{}, len(c))
	for i, col := range c {
		fields[i] = nameWithAlias(col.Name, alias)
		params[i] = col.Value
	}
	return fields, params
}

type _Query struct {
	model   TableModel
	columns _Columns
	where   Filter
}

func (q _Query) sqlRemains(alias string) (string, []interface{}) {
	statements := make([]string, 0)
	params := make([]interface{}, 0)
	if q.where != nil {
		statements = append(statements, fmt.Sprintf("WHERE %s", q.where.SqlString(alias)))
		params = append(params, q.where.Params()...)
	}
	return strings.Join(statements, " "), params
}

func (q _Query) exec(db *sql.DB, query string, params []interface{}) (sql.Result, error) {
	var result sql.Result
	start := time.Now().UnixNano()
	err := WithinTx(db, func(tx *sql.Tx) error {
		if stmt, txErr := tx.Prepare(query); txErr != nil {
			return txErr
		} else {
			result, txErr = stmt.Exec(params...)
			return txErr
		}
	})
	if Debug {
		log.Printf("Running SQL - [%s], params=%v, duration=%dms", query, params, (time.Now().UnixNano()-start)/1e6)
	}
	return result, err
}

////// Insert Query

type _InsertQuery struct {
	_Query
}

func (q _InsertQuery) Exec(db *sql.DB) (sql.Result, error) {
	if len(q.columns) == 0 {
		return nil, ErrNotEnoughColumns
	}
	table, _ := q.model.Names()
	fields, params := q.columns.fieldsAndParams("")
	qMarks := genQMarks(len(q.columns))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", dbQuote(table), strings.Join(fields, ", "), qMarks)
	return q.exec(db, query, params)
}

func (q _InsertQuery) Where(f Filter) Query { return q }

///// Update Query

type _UpdateQuery struct {
	_Query
}

func (q _UpdateQuery) Exec(db *sql.DB) (sql.Result, error) {
	if len(q.columns) == 0 {
		return nil, ErrNotEnoughColumns
	}
	table, _ := q.model.Names()
	fields, params := q.columns.fieldsAndParams("")
	for i, f := range fields {
		fields[i] = fmt.Sprintf("%s = ?", f)
	}
	query := fmt.Sprintf("UPDATE %s SET %s", dbQuote(table), strings.Join(fields, ", "))
	if remains, extras := q.sqlRemains(""); remains != "" && len(extras) > 0 {
		query = fmt.Sprintf("%s %s", query, remains)
		params = append(params, extras...)
	}
	return q.exec(db, query, params)
}

func (q _UpdateQuery) Where(f Filter) Query {
	q.where = f
	return q
}

///// Delete Query

type _DeleteQuery struct {
	_Query
}

func (q _DeleteQuery) Exec(db *sql.DB) (sql.Result, error) {
	table, _ := q.model.Names()
	query := fmt.Sprintf("DELETE FROM %s", dbQuote(table))
	var params []interface{}
	if remains, extras := q.sqlRemains(""); remains != "" && len(extras) > 0 {
		query = fmt.Sprintf("%s %s", query, remains)
		params = extras
	}
	return q.exec(db, query, params)
}

func (q _DeleteQuery) Where(f Filter) Query {
	q.where = f
	return q
}
