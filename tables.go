package main

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
)

var _ = fmt.Println

type Column struct {
	TableName    string
	ColumnName   string
	Position     int
	DefaultValue string
	IsNullable   bool
	DataType     string
	KeyType      string
	Extra        string
	Comment      string
}

type TableSchema []Column

func (ts TableSchema) Len() int { return len(ts) }
func (ts TableSchema) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}
func (ts TableSchema) Less(i, j int) bool {
	return ts[i].Position < ts[j].Position
}

type DbSchema map[string]TableSchema

func loadTablesMeta(cfg *_DsnConfig, tableNames string) (DbSchema, error) {
	tables := strings.Split(tableNames, ",")
	log.Printf("Start to load tables schema from db, %s, tables=%s", cfg.dbname, tables)
	db, err := sql.Open("mysql", cfg.dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	dbSchema := make(DbSchema)
	if len(tables) == 0 {
		err = queryColumns(db, cfg.dbname, "", dbSchema)
		if err != nil {
			return nil, err
		}
	} else {
		for _, table := range tables {
			err = queryColumns(db, cfg.dbname, table, dbSchema)
			if err != nil {
				return nil, err
			}
		}
	}

	for tbl, cols := range dbSchema {
		sort.Sort(cols)
		dbSchema[tbl] = cols
	}
	log.Printf("Loaded schema data of %d table(s) from db[%s]", len(dbSchema), cfg.dbname)
	return dbSchema, nil
}

func queryColumns(db *sql.DB, dbName string, table string, dbSchema DbSchema) error {
	kCols := 9
	q := `
		SELECT 
			TABLE_NAME, 
			COLUMN_NAME, 
			ORDINAL_POSITION, 
			COLUMN_DEFAULT, 
			IS_NULLABLE,
			DATA_TYPE,
			COLUMN_KEY,
			EXTRA,
			COLUMN_COMMENT
		FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=?`
	params := []interface{}{dbName}
	if table != "" {
		q = q + " AND TABLE_NAME=?"
		params = []interface{}{dbName, table}
	}

	err := dbQuery(db, q, params, func(r []sql.RawBytes) bool {
		if len(r) == kCols {
			col := Column{
				TableName:    asString(r[0]),
				ColumnName:   asString(r[1]),
				Position:     asInt(r[2]),
				DefaultValue: asString(r[3]),
				IsNullable:   asString(r[4]) == "YES",
				DataType:     asString(r[5]),
				KeyType:      asString(r[6]),
				Extra:        asString(r[7]),
				Comment:      asString(r[8]),
			}
			if _, ok := dbSchema[col.TableName]; !ok {
				dbSchema[col.TableName] = make(TableSchema, 0)
			}
			dbSchema[col.TableName] = append(dbSchema[col.TableName], col)
		}
		return true
	})
	if err != nil {
		return err
	}

	return nil
}
