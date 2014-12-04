package main

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
)

var _ = fmt.Println

type _Column struct {
	tblName      string
	colName      string
	position     int
	defaultValue string
	isNullable   bool
	dataType     string
	keyType      string
	extra        string
	comment      string
}

type _TableSchema []_Column

func (ts _TableSchema) Len() int { return len(ts) }
func (ts _TableSchema) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}
func (ts _TableSchema) Less(i, j int) bool {
	return ts[i].position < ts[j].position
}

type _DbSchema map[string]_TableSchema

func loadTablesMeta(cfg *_DsnConfig, tableNames string) (_DbSchema, error) {
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

	dbSchema := make(_DbSchema)
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

func queryColumns(db *sql.DB, dbName string, table string, dbSchema _DbSchema) error {
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
			col := _Column{
				tblName:      asString(r[0]),
				colName:      asString(r[1]),
				position:     asInt(r[2]),
				defaultValue: asString(r[3]),
				isNullable:   asString(r[4]) == "YES",
				dataType:     asString(r[5]),
				keyType:      asString(r[6]),
				extra:        asString(r[7]),
				comment:      asString(r[8]),
			}
			if _, ok := dbSchema[col.tblName]; !ok {
				dbSchema[col.tblName] = make(_TableSchema, 0)
			}
			dbSchema[col.tblName] = append(dbSchema[col.tblName], col)
		}
		return true
	})
	if err != nil {
		return err
	}

	return nil
}
