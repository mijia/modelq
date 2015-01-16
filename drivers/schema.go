package drivers

import (
	"fmt"
	"strings"
)

type Column struct {
	Schema       string
	TableName    string
	ColumnName   string
	DefaultValue string
	DataType     string
	ColumnType   string
	ColumnKey    string
	Extra        string
	Comment      string
}

type TableSchema []Column
type DbSchema map[string]TableSchema

type Driver interface {
	LoadDatabaseSchema(dsnString string, schema string, tableNames string) (DbSchema, error)
}

func LoadDatabaseSchema(driver, dsnString, schema, tableNames string) (dbSchema DbSchema, err error) {
	if driver, ok := drivers[strings.ToLower(driver)]; !ok {
		return nil, fmt.Errorf("Not supported driver %s", driver)
	} else {
		return driver.LoadDatabaseSchema(dsnString, schema, tableNames)
	}
}

var drivers map[string]Driver

func init() {
	drivers = map[string]Driver{
		"mysql":    MysqlDriver{},
		"postgres": PostgresDriver{},
	}
}
