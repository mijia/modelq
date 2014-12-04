package main

import (
	"database/sql"
	"log"
	"sort"
	"strconv"
	"strings"
	"fmt"
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

	err := query(db,
		func(r []sql.RawBytes) {
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
		},
		q, params...)
	if err != nil {
		return err
	}

	return nil
}

type rowVisitor func([]sql.RawBytes)

func query(db *sql.DB, visitor rowVisitor, q string, params ...interface{}) error {
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
			visitor(vals)
		}
	}
	return nil
}

func asString(rb sql.RawBytes) string {
	if len(rb) > 0 {
		return string(rb)
	}
	return ""
}

func asInt(rb sql.RawBytes) int {
	if len(rb) > 0 {
		ans, _ := strconv.Atoi(string(rb))
		return ans
	}
	return 0
}
