package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/mijia/modelq/drivers"
)

type CodeResult struct {
	name string
	err  error
}

type CodeConfig struct {
	packageName    string
	touchTimestamp bool
}

func generateModels(dbName string, dbSchema drivers.DbSchema, config CodeConfig) {
	if fs, err := os.Stat(config.packageName); err != nil || !fs.IsDir() {
		os.Mkdir(config.packageName, os.ModeDir|os.ModePerm)
	}

	jobs := make(chan CodeResult)
	for tbl, cols := range dbSchema {
		go func(tableName string, schema drivers.TableSchema) {
			err := generateModel(dbName, tableName, schema, config)
			jobs <- CodeResult{tableName, err}
		}(tbl, cols)
	}

	for i := 0; i < len(dbSchema); i++ {
		result := <-jobs
		if result.err != nil {
			log.Printf("Error when generating code for %s, %s", result.name, result.err)
		} else {
			log.Printf("Code generated for table %s, into package %s/%s.go", result.name, config.packageName, result.name)
		}
	}
	close(jobs)
}

func generateModel(dbName, tName string, schema drivers.TableSchema, config CodeConfig) error {
	file, err := os.Create(path.Join(config.packageName, tName+".go"))
	if err != nil {
		return err
	}
	w := bufio.NewWriter(file)

	defer func() {
		w.Flush()
		file.Close()
	}()

	model := ModelMeta{
		Name:      toCapitalCase(tName),
		DbName:    dbName,
		TableName: tName,
		Fields:    make([]ModelField, len(schema)),
		config:    config,
	}
	needTime := false
	for i, col := range schema {
		field := ModelField{
			Name:            toCapitalCase(col.ColumnName),
			ColumnName:      col.ColumnName,
			Type:            col.DataType,
			JsonMeta:        fmt.Sprintf("`json:\"%s\"`", col.ColumnName),
			IsPrimaryKey:    strings.ToUpper(col.ColumnKey) == "PRI",
			IsAutoIncrement: strings.ToUpper(col.Extra) == "AUTO_INCREMENT",
			DefaultValue:    col.DefaultValue,
			Extra:           col.Extra,
			Comment:         col.Comment,
		}
		if field.Type == "time.Time" {
			needTime = true
		}
		if field.IsPrimaryKey {
			model.PrimaryField = &field
		}
		model.Fields[i] = field
	}

	if err := model.GenHeader(w, needTime); err != nil {
		return errors.New(fmt.Sprintf("[%s] Fail to gen model header, %s", tName, err))
	}
	if err := model.GenStruct(w); err != nil {
		return errors.New(fmt.Sprintf("[%s] Fail to gen model struct, %s", tName, err))
	}
	if err := model.GenObjectApi(w); err != nil {
		return errors.New(fmt.Sprintf("[%s] Fail to gen model object api, %s", tName, err))
	}
	if err := model.GenQueryApi(w); err != nil {
		return errors.New(fmt.Sprintf("[%s] Fail to gen model query api, %s", tName, err))
	}
	if err := model.GenManagedObjApi(w); err != nil {
		return errors.New(fmt.Sprintf("[%s] Fail to gen model managed objects api, %s", tName, err))
	}

	return nil
}

type ModelField struct {
	Name            string
	ColumnName      string
	Type            string
	JsonMeta        string
	IsPrimaryKey    bool
	IsAutoIncrement bool
	DefaultValue    string
	Extra           string
	Comment         string
}

func (f ModelField) ConverterFuncName() string {
	convertors := map[string]string{
		"int64":     "AsInt64",
		"int":       "AsInt",
		"string":    "AsString",
		"time.Time": "AsTime",
		"float64":   "AsFloat64",
		"bool":      "AsBool",
	}
	if c, ok := convertors[f.Type]; ok {
		return c
	}
	return "AsString"
}

type ModelMeta struct {
	Name         string
	DbName       string
	TableName    string
	PrimaryField *ModelField
	Fields       []ModelField
	config       CodeConfig
}

func (m ModelMeta) HasAutoIncrementPrimaryKey() bool {
	return m.PrimaryField != nil && m.PrimaryField.IsAutoIncrement
}

func (m ModelMeta) AllFields() string {
	fields := make([]string, len(m.Fields))
	for i, f := range m.Fields {
		fields[i] = fmt.Sprintf("\"%s\"", f.Name)
	}
	return strings.Join(fields, ", ")
}

func (m ModelMeta) InsertableFields() string {
	fields := make([]string, 0, len(m.Fields))
	for _, f := range m.Fields {
		if f.IsPrimaryKey && f.IsAutoIncrement {
			continue
		}
		autoTimestamp := strings.ToUpper(f.DefaultValue) == "CURRENT_TIMESTAMP" ||
			strings.ToUpper(f.DefaultValue) == "NOW()"
		if f.Type == "time.Time" && autoTimestamp && !m.config.touchTimestamp {
			continue
		}
		fields = append(fields, fmt.Sprintf("\"%s\"", f.Name))
	}
	return strings.Join(fields, ", ")
}

func (m ModelMeta) UpdatableFields() string {
	fields := make([]string, 0, len(m.Fields))
	for _, f := range m.Fields {
		if f.IsPrimaryKey {
			continue
		}
		autoUpdateTime := strings.ToUpper(f.Extra) == "ON UPDATE CURRENT_TIMESTAMP"
		if autoUpdateTime && !m.config.touchTimestamp {
			continue
		}
		fields = append(fields, fmt.Sprintf("\"%s\"", f.Name))
	}
	return strings.Join(fields, ", ")
}

func (m ModelMeta) GenHeader(w *bufio.Writer, importTime bool) error {
	return tmHeader.Execute(w, map[string]interface{}{
		"DbName":     m.DbName,
		"TableName":  m.TableName,
		"PkgName":    m.config.packageName,
		"ImportTime": importTime,
	})
}

func (m ModelMeta) GenStruct(w *bufio.Writer) error {
	return tmStruct.Execute(w, m)
}

func (m ModelMeta) GenObjectApi(w *bufio.Writer) error {
	return tmObjApi.Execute(w, m)
}

func (m ModelMeta) GenQueryApi(w *bufio.Writer) error {
	return tmQueryApi.Execute(w, m)
}

func (m ModelMeta) GenManagedObjApi(w *bufio.Writer) error {
	return tmManagedObjApi.Execute(w, m)
}

func toCapitalCase(name string) string {
	// cp___hello_12jiu -> CpHello12Jiu
	data := []byte(name)
	segStart := true
	endPos := 0
	for i := 0; i < len(data); i++ {
		ch := data[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			if segStart {
				if ch >= 'a' && ch <= 'z' {
					ch = ch - 'a' + 'A'
				}
				segStart = false
			} else {
				if ch >= 'A' && ch <= 'Z' {
					ch = ch - 'A' + 'a'
				}
			}
			data[endPos] = ch
			endPos++
		} else if ch >= '0' && ch <= '9' {
			data[endPos] = ch
			endPos++
			segStart = true
		} else {
			segStart = true
		}
	}
	return string(data[:endPos])
}
