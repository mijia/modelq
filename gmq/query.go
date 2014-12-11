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
		log.Printf("[SQL] %s, duration=%dms", query, (time.Now().UnixNano()-start)/1e6)
	}
	return result, err
}

type _InsertQuery struct {
	_Query
}

func (q _InsertQuery) Exec(db *sql.DB) (sql.Result, error) {
	table, _ := q.model.Names()
	fields, params := q.columns.fieldsAndParams("")
	qMarks := genQMarks(len(q.columns))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(fields, ", "), qMarks)
	return q.exec(db, query, params)
}
