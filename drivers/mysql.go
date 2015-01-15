package drivers

import (
	"log"
	"strings"

	"github.com/mijia/modelq/drivers/mysql"
	"github.com/mijia/modelq/gmq"
)

type MysqlDriver struct{}

func (m MysqlDriver) LoadDatabaseSchema(dsnString, schema, tableNames string) (DbSchema, error) {
	log.Printf("[MySQL Driver] Start to load tables schema from db, %s, tables=%s", schema, tableNames)
	db, err := gmq.Open("mysql", m.useInformationSchema(dsnString, schema))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return nil, err
	}

	dbSchema := make(DbSchema)
	if err = m.queryColumns(db, schema, tableNames, dbSchema); err != nil {
		return nil, err
	}
	log.Printf("[MySQL Driver] Loaded schema data of %d tables from db[%s]", len(dbSchema), schema)
	return dbSchema, nil
}

func (m MysqlDriver) dataType(colDataType string) string {
	kFieldTypes := map[string]string{
		"bigint":   "int64",
		"int":      "int",
		"tinyint":  "int",
		"char":     "string",
		"varchar":  "string",
		"datetime": "time.Time",
		"decimal":  "float64",
	}
	if fieldType, ok := kFieldTypes[strings.ToLower(colDataType)]; !ok {
		return "string"
	} else {
		return fieldType
	}
}

func (m MysqlDriver) queryColumns(db *gmq.Db, dbName string, tables string, dbSchema DbSchema) error {
	objs := mysql.ColumnsObjs
	filter := objs.FilterTableSchema("=", dbName)
	if len(tables) > 0 {
		tableVs := strings.Split(tables, ",")
		filter = filter.And(objs.FilterTableName("IN", tableVs[0], tableVs[1:]...))
	}

	query := objs.Select().Where(filter).OrderBy("TableName", "OrdinalPosition")
	return query.Iterate(db, func(col mysql.Columns) bool {
		if _, ok := dbSchema[col.TableName]; !ok {
			dbSchema[col.TableName] = make(TableSchema, 0, 5)
		}
		sCol := Column{
			Schema:       col.TableSchema,
			TableName:    col.TableName,
			ColumnName:   col.ColumnName,
			DefaultValue: col.ColumnDefault,
			DataType:     m.dataType(col.DataType),
			ColumnType:   col.ColumnType,
			ColumnKey:    col.ColumnKey,
			Extra:        col.Extra,
			Comment:      col.ColumnComment,
		}
		dbSchema[col.TableName] = append(dbSchema[col.TableName], sCol)
		return true
	})
}

func (m MysqlDriver) useInformationSchema(dsn string, schema string) string {
	return strings.Replace(dsn, "/"+schema, "/information_schema", 1)
}
