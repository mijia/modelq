package drivers

import (
	"fmt"
	"log"
	"strings"

	"github.com/mijia/modelq/drivers/postgres"
	"github.com/mijia/modelq/gmq"
)

type StringSet map[string]struct{}
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

func (p PostgresDriver) queryPrimaryKeys(db *gmq.Db, dbName string, tables string) (StringSet, error) {
	// FIXME: if we have implemented the JOIN
	pKeys := make(StringSet)

	tcObjs := postgres.TableConstraintsObjs
	kcuObjs := postgres.KeyColumnUsageObjs
	tcFilter := tcObjs.FilterTableSchema("=", dbName).And(tcObjs.FilterConstraintType("=", "PRIMARY KEY"))
	kcuFilter := kcuObjs.FilterTableSchema("=", dbName)
	if len(tables) > 0 {
		tableVs := strings.Split(tables, ",")
		tcFilter = tcFilter.And(tcObjs.FilterTableName("IN", tableVs[0], tableVs[1:]...))
		kcuFilter = kcuFilter.And(kcuObjs.FilterTableName("IN", tableVs[0], tableVs[1:]...))
	}

	tcJoinKeys := make(StringSet)
	err := tcObjs.Select().Where(tcFilter).Iterate(db, func(tc postgres.TableConstraints) bool {
		key := fmt.Sprintf("%s.%s", tc.TableName, tc.ConstraintName)
		tcJoinKeys[key] = struct{}{}
		return true
	})
	if err != nil {
		return pKeys, err
	}

	err = kcuObjs.Select().Where(kcuFilter).Iterate(db, func(kcu postgres.KeyColumnUsage) bool {
		key := fmt.Sprintf("%s.%s", kcu.TableName, kcu.ConstraintName)
		if _, ok := tcJoinKeys[key]; ok {
			pkey := fmt.Sprintf("%s.%s", kcu.TableName, kcu.ColumnName)
			pKeys[pkey] = struct{}{}
		}
		return true
	})
	if err != nil {
		return pKeys, err
	}

	return pKeys, nil
}

func (p PostgresDriver) queryColumns(db *gmq.Db, dbName string, tables string, dbSchema DbSchema) error {
	pKeys, err := p.queryPrimaryKeys(db, dbName, tables)
	if err != nil {
		return err
	}

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
		columnKey := fmt.Sprintf("%s.%s", col.TableName, col.ColumnName)
		if _, ok := pKeys[columnKey]; ok {
			columnKey = "PRI"
		}
		sCol := Column{
			Schema:       col.TableSchema,
			TableName:    col.TableName,
			ColumnName:   col.ColumnName,
			DefaultValue: col.ColumnDefault,
			DataType:     p.dataType(col.DataType),
			ColumnKey:    columnKey,
			Extra:        extra,
		}
		dbSchema[col.TableName] = append(dbSchema[col.TableName], sCol)
		return true
	})
}
