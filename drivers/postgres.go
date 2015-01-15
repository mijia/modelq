package drivers

import (
	"fmt"
	"log"
	"strings"

	"github.com/mijia/modelq/drivers/postgres"
	"github.com/mijia/modelq/gmq"
)

type PostgresDriver struct{}

func (p PostgresDriver) LoadDatabaseSchema(dsnString, schema, tableNames string) (DbSchema, error) {
	log.Printf("[Postgres Driver] Start to load tables schema from database, %s, tables=%s", schema, tableNames)
	db, err := gmq.Open("postgres", dsnString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return nil, err
	}

	dbSchema := make(DbSchema)
	if err = p.queryColumns(db, schema, tableNames, dbSchema); err != nil {
		return nil, err
	}
	log.Printf("[Postgres Driver] Loaded schema data of %d tables from database schema[%s]", len(dbSchema), schema)
	return dbSchema, nil
}

func (p PostgresDriver) dataType(dt string) string {
	kFieldTypes := map[string]string{
		"bigint":    "int64",
		"int":       "int",
		"integer":   "int",
		"smallint":  "int",
		"character": "string",
		"text":      "string",
		"timestamp": "time.Time",
		"numeric":   "float64",
		"boolean":   "bool",
	}
	dt = strings.Split(dt, " ")[0]
	if fieldType, ok := kFieldTypes[strings.ToLower(dt)]; !ok {
		return "string"
	} else {
		return fieldType
	}
}

func (p PostgresDriver) queryColumns(db *gmq.Db, dbName string, tables string, dbSchema DbSchema) error {
	// 1. need to extract PRIMARY KEYS from information_schema.TABLE_CONSTRAINTS
	// 2. find the columns/tables from information_schema.KEY_COLUMN_USAGE
	objs := postgres.ColumnsObjs
	filter := objs.FilterTableSchema("=", dbName)
	if len(tables) > 0 {
		tableVs := strings.Split(tables, ",")
		filter = filter.And(objs.FilterTableName("IN", tableVs[0], tableVs[1:]...))
	}

	query := objs.Select().Where(filter).OrderBy("TableName", "OrdinalPosition")
	return query.Iterate(db, func(col postgres.Columns) bool {
		if _, ok := dbSchema[col.TableName]; !ok {
			dbSchema[col.TableName] = make(TableSchema, 0, 5)
		}
		extra := ""
		if strings.HasPrefix(col.ColumnDefault, "nextval(") {
			extra = "AUTO_INCREMENT"
		}
		sCol := Column{
			Schema:       col.TableSchema,
			TableName:    col.TableName,
			ColumnName:   col.ColumnName,
			DefaultValue: col.ColumnDefault,
			DataType:     p.dataType(col.DataType),
			Extra:        extra,
		}
		dbSchema[col.TableName] = append(dbSchema[col.TableName], sCol)
		fmt.Println(col)
		return true
	})
}
