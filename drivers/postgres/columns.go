package postgres

import (
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mijia/modelq/gmq"
)

type Columns struct {
	TableSchema     string
	TableName       string
	ColumnName      string
	OrdinalPosition int64
	ColumnDefault   string
	IsNullable      string
	DataType        string
}

func (obj Columns) String() string {
	if data, err := json.Marshal(obj); err != nil {
		return fmt.Sprintf("<Columns>")
	} else {
		return string(data)
	}
}

func (obj Columns) Insert(dbtx gmq.DbTx) (Columns, error) {
	_, err := ColumnsObjs.Insert(obj).Run(dbtx)
	return obj, err
}

func (obj Columns) Update(dbtx gmq.DbTx) (int64, error) {
	return 0, gmq.ErrNoPrimaryKeyDefined
}

func (obj Columns) Delete(dbtx gmq.DbTx) (int64, error) {
	return 0, gmq.ErrNoPrimaryKeyDefined
}

type _ColumnsQuery struct {
	gmq.Query
}

func (q _ColumnsQuery) Where(f gmq.Filter) _ColumnsQuery {
	q.Query = q.Query.Where(f)
	return q
}

func (q _ColumnsQuery) OrderBy(by ...string) _ColumnsQuery {
	tBy := make([]string, 0, len(by))
	for _, b := range by {
		sortDir := ""
		if b[0] == '-' || b[0] == '+' {
			sortDir = string(b[0])
			b = b[1:]
		}
		if col, ok := ColumnsObjs.fcMap[b]; ok {
			tBy = append(tBy, sortDir+col)
		}
	}
	q.Query = q.Query.OrderBy(tBy...)
	return q
}

func (q _ColumnsQuery) GroupBy(by ...string) _ColumnsQuery {
	tBy := make([]string, 0, len(by))
	for _, b := range by {
		if col, ok := ColumnsObjs.fcMap[b]; ok {
			tBy = append(tBy, col)
		}
	}
	q.Query = q.Query.GroupBy(tBy...)
	return q
}

func (q _ColumnsQuery) Limit(offsets ...int64) _ColumnsQuery {
	q.Query = q.Query.Limit(offsets...)
	return q
}

func (q _ColumnsQuery) Page(number, size int) _ColumnsQuery {
	q.Query = q.Query.Page(number, size)
	return q
}

func (q _ColumnsQuery) Run(dbtx gmq.DbTx) (sql.Result, error) {
	return q.Query.Exec(dbtx)
}

type ColumnsRowVisitor func(obj Columns) bool

func (q _ColumnsQuery) Iterate(dbtx gmq.DbTx, functor ColumnsRowVisitor) error {
	return q.Query.SelectList(dbtx, func(columns []gmq.Column, rb []sql.RawBytes) bool {
		obj := ColumnsObjs.toColumns(columns, rb)
		return functor(obj)
	})
}

func (q _ColumnsQuery) One(dbtx gmq.DbTx) (Columns, error) {
	var obj Columns
	err := q.Query.SelectOne(dbtx, func(columns []gmq.Column, rb []sql.RawBytes) bool {
		obj = ColumnsObjs.toColumns(columns, rb)
		return true
	})
	return obj, err
}

func (q _ColumnsQuery) List(dbtx gmq.DbTx) ([]Columns, error) {
	result := make([]Columns, 0, 10)
	err := q.Query.SelectList(dbtx, func(columns []gmq.Column, rb []sql.RawBytes) bool {
		obj := ColumnsObjs.toColumns(columns, rb)
		result = append(result, obj)
		return true
	})
	return result, err
}

type _ColumnsObjs struct {
	fcMap map[string]string
}

func (o _ColumnsObjs) Names() (schema, tbl, alias string) {
	return "information_schema", "columns", "Columns"
}

func (o _ColumnsObjs) Select(fields ...string) _ColumnsQuery {
	q := _ColumnsQuery{}
	if len(fields) == 0 {
		fields = []string{"TableSchema", "TableName", "ColumnName", "OrdinalPosition", "ColumnDefault", "IsNullable", "DataType"}
	}
	q.Query = gmq.Select(o, o.columns(fields...))
	return q
}

func (o _ColumnsObjs) Insert(obj Columns) _ColumnsQuery {
	q := _ColumnsQuery{}
	q.Query = gmq.Insert(o, o.columnsWithData(obj, "TableSchema", "TableName", "ColumnName", "OrdinalPosition", "ColumnDefault", "IsNullable", "DataType"))
	return q
}

func (o _ColumnsObjs) Update(obj Columns, fields ...string) _ColumnsQuery {
	q := _ColumnsQuery{}
	q.Query = gmq.Update(o, o.columnsWithData(obj, fields...))
	return q
}

func (o _ColumnsObjs) Delete() _ColumnsQuery {
	q := _ColumnsQuery{}
	q.Query = gmq.Delete(o)
	return q
}

func (o _ColumnsObjs) FilterTableSchema(op string, p string, ps ...string) gmq.Filter {
	params := make([]interface{}, 1+len(ps))
	params[0] = p
	for i := range ps {
		params[i+1] = ps[i]
	}
	return o.newFilter("table_schema", op, params...)
}

func (o _ColumnsObjs) FilterTableName(op string, p string, ps ...string) gmq.Filter {
	params := make([]interface{}, 1+len(ps))
	params[0] = p
	for i := range ps {
		params[i+1] = ps[i]
	}
	return o.newFilter("table_name", op, params...)
}

func (o _ColumnsObjs) FilterColumnName(op string, p string, ps ...string) gmq.Filter {
	params := make([]interface{}, 1+len(ps))
	params[0] = p
	for i := range ps {
		params[i+1] = ps[i]
	}
	return o.newFilter("column_name", op, params...)
}

func (o _ColumnsObjs) FilterOrdinalPosition(op string, p int64, ps ...int64) gmq.Filter {
	params := make([]interface{}, 1+len(ps))
	params[0] = p
	for i := range ps {
		params[i+1] = ps[i]
	}
	return o.newFilter("ordinal_position", op, params...)
}

func (o _ColumnsObjs) FilterColumnDefault(op string, p string, ps ...string) gmq.Filter {
	params := make([]interface{}, 1+len(ps))
	params[0] = p
	for i := range ps {
		params[i+1] = ps[i]
	}
	return o.newFilter("column_default", op, params...)
}

func (o _ColumnsObjs) FilterIsNullable(op string, p string, ps ...string) gmq.Filter {
	params := make([]interface{}, 1+len(ps))
	params[0] = p
	for i := range ps {
		params[i+1] = ps[i]
	}
	return o.newFilter("is_nullable", op, params...)
}

func (o _ColumnsObjs) FilterDataType(op string, p string, ps ...string) gmq.Filter {
	params := make([]interface{}, 1+len(ps))
	params[0] = p
	for i := range ps {
		params[i+1] = ps[i]
	}
	return o.newFilter("data_type", op, params...)
}

func (o _ColumnsObjs) ColumnTableSchema(p ...string) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"table_schema", value}
}

func (o _ColumnsObjs) ColumnTableName(p ...string) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"table_name", value}
}

func (o _ColumnsObjs) ColumnColumnName(p ...string) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"column_name", value}
}

func (o _ColumnsObjs) ColumnOrdinalPosition(p ...int64) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"ordinal_position", value}
}

func (o _ColumnsObjs) ColumnColumnDefault(p ...string) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"column_default", value}
}

func (o _ColumnsObjs) ColumnIsNullable(p ...string) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"is_nullable", value}
}

func (o _ColumnsObjs) ColumnDataType(p ...string) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"data_type", value}
}

func (o _ColumnsObjs) newFilter(name, op string, params ...interface{}) gmq.Filter {
	if strings.ToUpper(op) == "IN" {
		return gmq.InFilter(name, params)
	}
	return gmq.UnitFilter(name, op, params[0])
}

func (o _ColumnsObjs) toColumns(columns []gmq.Column, rb []sql.RawBytes) Columns {
	obj := Columns{}
	if len(columns) == len(rb) {
		for i := range columns {
			switch columns[i].Name {
			case "table_schema":
				obj.TableSchema = gmq.AsString(rb[i])
			case "table_name":
				obj.TableName = gmq.AsString(rb[i])
			case "column_name":
				obj.ColumnName = gmq.AsString(rb[i])
			case "ordinal_position":
				obj.OrdinalPosition = gmq.AsInt64(rb[i])
			case "column_default":
				obj.ColumnDefault = gmq.AsString(rb[i])
			case "is_nullable":
				obj.IsNullable = gmq.AsString(rb[i])
			case "data_type":
				obj.DataType = gmq.AsString(rb[i])
			}
		}
	}
	return obj
}

func (o _ColumnsObjs) columns(fields ...string) []gmq.Column {
	data := make([]gmq.Column, 0, len(fields))
	for _, f := range fields {
		switch f {
		case "TableSchema":
			data = append(data, o.ColumnTableSchema())
		case "TableName":
			data = append(data, o.ColumnTableName())
		case "ColumnName":
			data = append(data, o.ColumnColumnName())
		case "OrdinalPosition":
			data = append(data, o.ColumnOrdinalPosition())
		case "ColumnDefault":
			data = append(data, o.ColumnColumnDefault())
		case "IsNullable":
			data = append(data, o.ColumnIsNullable())
		case "DataType":
			data = append(data, o.ColumnDataType())
		}
	}
	return data
}

func (o _ColumnsObjs) columnsWithData(obj Columns, fields ...string) []gmq.Column {
	data := make([]gmq.Column, 0, len(fields))
	for _, f := range fields {
		switch f {
		case "TableSchema":
			data = append(data, o.ColumnTableSchema(obj.TableSchema))
		case "TableName":
			data = append(data, o.ColumnTableName(obj.TableName))
		case "ColumnName":
			data = append(data, o.ColumnColumnName(obj.ColumnName))
		case "OrdinalPosition":
			data = append(data, o.ColumnOrdinalPosition(obj.OrdinalPosition))
		case "ColumnDefault":
			data = append(data, o.ColumnColumnDefault(obj.ColumnDefault))
		case "IsNullable":
			data = append(data, o.ColumnIsNullable(obj.IsNullable))
		case "DataType":
			data = append(data, o.ColumnDataType(obj.DataType))
		}
	}
	return data
}

var ColumnsObjs _ColumnsObjs

func init() {
	ColumnsObjs.fcMap = map[string]string{
		"TableSchema":     "table_schema",
		"TableName":       "table_name",
		"ColumnName":      "column_name",
		"OrdinalPosition": "ordinal_position",
		"ColumnDefault":   "column_default",
		"IsNullable":      "is_nullable",
		"DataType":        "data_type",
	}
	gob.Register(Columns{})
}
