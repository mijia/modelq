package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
	"unicode"

	"github.com/mijia/modelq/drivers"
)

type CodeResult struct {
	name string
	err  error
}

type CodeConfig struct {
	packageName    string
	touchTimestamp bool
	template       string
}

func (cc CodeConfig) MustCompileTemplate() *template.Template {
	if cc.template == "" {
		return nil
	}
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"ToLower": strings.ToLower,
		"FirstLower": func(s string) string {
			b := []rune(s)
			b[0] = unicode.ToLower(b[0])
			return string(b)
		},
		"First": func(s string) string {
			if "" == s {
				return ""
			}
			return s[:1]
		},
		"SubStr": func(from int, to int, s string) string {
			l := len(s)
			if from > l {
				return ""
			}
			if to > l {
				return s[from:]
			}
			if 0 == to {
				return s[from:]
			}
			return s[from:to]
		},
	}
	tmpl := template.New("tmp").Funcs(funcMap)

	tmpl, err := tmpl.ParseFiles(cc.template)
	return template.Must(tmpl, err)
}

func generateModels(dbName string, dbSchema drivers.DbSchema, config CodeConfig) {
	customTmpl := config.MustCompileTemplate()

	if fs, err := os.Stat(config.packageName); err != nil || !fs.IsDir() {
		os.Mkdir(config.packageName, os.ModeDir|os.ModePerm)
	}

	jobs := make(chan CodeResult)
	for tbl, cols := range dbSchema {
		go func(tableName string, schema drivers.TableSchema) {
			err := generateModel(dbName, tableName, schema, config, customTmpl)
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

func generateModel(dbName, tName string, schema drivers.TableSchema, config CodeConfig, tmpl *template.Template) error {
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
		Uniques:   make([]ModelField, 0, len(schema)),
		Indexed:   make([]ModelField, 0, len(schema)),
		config:    config,
	}
	needTime := false
	for i, col := range schema {
		field := ModelField{
			Name:            toCapitalCase(col.ColumnName),
			ColumnName:      col.ColumnName,
			Type:            col.DataType,
			IsNullable:      strings.ToUpper(col.IsNullable) == "YES",
			JsonMeta:        fmt.Sprintf("`json:\"%s\"`", col.ColumnName),
			IsPrimaryKey:    strings.ToUpper(col.ColumnKey) == "PRI",
			IsUniqueKey:     strings.ToUpper(col.ColumnKey) == "UNI",
			IsIndexed:       strings.ToUpper(col.ColumnKey) == "MUL",
			IsAutoIncrement: strings.ToUpper(col.Extra) == "AUTO_INCREMENT",
			DefaultValue:    col.DefaultValue,
			Extra:           col.Extra,
			Comment:         col.Comment,
		}
		if field.Type == "time.Time" {
			needTime = true
		}
		if field.IsPrimaryKey {
			model.PrimaryFields = append(model.PrimaryFields, &field)
		}

		if field.IsUniqueKey {
			model.Uniques = append(model.Uniques, field)
		}

		if field.IsIndexed {
			model.Indexed = append(model.Indexed, field)
		}

		model.Fields[i] = field
	}

	if err := model.GenHeader(w, tmpl, needTime); err != nil {
		return fmt.Errorf("[%s] Fail to gen model header, %s", tName, err)
	}
	if err := model.GenStruct(w, tmpl); err != nil {
		return fmt.Errorf("[%s] Fail to gen model struct, %s", tName, err)
	}
	if err := model.GenObjectApi(w, tmpl); err != nil {
		return fmt.Errorf("[%s] Fail to gen model object api, %s", tName, err)
	}
	if err := model.GenQueryApi(w, tmpl); err != nil {
		return fmt.Errorf("[%s] Fail to gen model query api, %s", tName, err)
	}
	if err := model.GenManagedObjApi(w, tmpl); err != nil {
		return fmt.Errorf("[%s] Fail to gen model managed objects api, %s", tName, err)
	}

	return nil
}

type ModelField struct {
	Name            string
	ColumnName      string
	Type            string
	JsonMeta        string
	IsNullable      bool
	IsPrimaryKey    bool
	IsUniqueKey     bool
	IsIndexed       bool
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

type PrimaryFields []*ModelField

func (pf PrimaryFields) FormatObject() func(string) string {
	return func(name string) string {
		// "<Article ArticleId=%v UserId=%v>", obj.ArticleId, obj.UserId
		formats := make([]string, len(pf))
		for i, field := range pf {
			formats[i] = fmt.Sprintf("%s=%%v", field.Name)
		}
		outputs := make([]string, 1+len(pf))
		outputs[0] = fmt.Sprintf("\"<%s %s>\"", name, strings.Join(formats, " "))
		for i, field := range pf {
			outputs[i+1] = fmt.Sprintf("obj.%s", field.Name)
		}
		return strings.Join(outputs, ", ")
	}
}

func (pf PrimaryFields) FormatIncrementId() func() string {
	// obj.Id = {{if eq .PrimaryField.Type "int64"}}id{{else}}{{.PrimaryField.Type}}(id){{end}}
	return func() string {
		for _, field := range pf {
			if field.IsAutoIncrement {
				if field.Type == "int64" {
					return fmt.Sprintf("obj.%s = id", field.Name)
				} else {
					return fmt.Sprintf("obj.%s = %s(id)", field.Name, field.Type)
				}
			}
		}
		return ""
	}
}

func (pf PrimaryFields) FormatFilters() func(string) string {
	// filter := {{.Name}}Objs.Filter{{.PrimaryField.Name}}("=", obj.{{.PrimaryField.Name}})
	return func(name string) string {
		filters := make([]string, len(pf))
		for i, field := range pf {
			if i == 0 {
				filters[i] = fmt.Sprintf("filter := %sObjs.Filter%s(\"=\", obj.%s)", name, field.Name, field.Name)
			} else {
				filters[i] = fmt.Sprintf("filter = filter.And(%sObjs.Filter%s(\"=\", obj.%s))", name, field.Name, field.Name)
			}
		}
		return strings.Join(filters, "\n")
	}
}

type ModelMeta struct {
	Name          string
	DbName        string
	TableName     string
	PrimaryFields PrimaryFields
	Fields        []ModelField
	Uniques       []ModelField
	Indexed       []ModelField
	config        CodeConfig
}

func (m ModelMeta) HasAutoIncrementPrimaryKey() bool {
	for _, pField := range m.PrimaryFields {
		if pField.IsAutoIncrement {
			return true
		}
	}
	return false
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

func (m ModelMeta) GetInsertableFields() []ModelField {
	fields := make([]ModelField, 0, len(m.Fields))
	for _, f := range m.Fields {
		if f.IsPrimaryKey && f.IsAutoIncrement {
			continue
		}
		autoTimestamp := strings.ToUpper(f.DefaultValue) == "CURRENT_TIMESTAMP" ||
			strings.ToUpper(f.DefaultValue) == "NOW()"
		if f.Type == "time.Time" && autoTimestamp && !m.config.touchTimestamp {
			continue
		}
		fields = append(fields, f)
	}
	return fields
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

func (m ModelMeta) GetUpdatableFields() []ModelField {
	fields := make([]ModelField, 0, len(m.Fields))
	for _, f := range m.Fields {
		if f.IsPrimaryKey {
			continue
		}
		autoUpdateTime := strings.ToUpper(f.Extra) == "ON UPDATE CURRENT_TIMESTAMP"
		if autoUpdateTime && !m.config.touchTimestamp {
			continue
		}
		fields = append(fields, f)
	}
	return fields
}

func (m ModelMeta) getTemplate(tmpl *template.Template, name string, defaultTmpl *template.Template) *template.Template {
	if tmpl != nil {
		if definedTmpl := tmpl.Lookup(name); definedTmpl != nil {
			return definedTmpl
		}
	}
	return defaultTmpl
}

func (m ModelMeta) GenHeader(w *bufio.Writer, tmpl *template.Template, importTime bool) error {
	return m.getTemplate(tmpl, "header", tmHeader).Execute(w, map[string]interface{}{
		"DbName":     m.DbName,
		"TableName":  m.TableName,
		"PkgName":    m.config.packageName,
		"ImportTime": importTime,
	})
}

func (m ModelMeta) GenStruct(w *bufio.Writer, tmpl *template.Template) error {
	return m.getTemplate(tmpl, "struct", tmStruct).Execute(w, m)
}

func (m ModelMeta) GenObjectApi(w *bufio.Writer, tmpl *template.Template) error {
	return m.getTemplate(tmpl, "obj_api", tmObjApi).Execute(w, m)
}

func (m ModelMeta) GenQueryApi(w *bufio.Writer, tmpl *template.Template) error {
	return m.getTemplate(tmpl, "query_api", tmQueryApi).Execute(w, m)
}

func (m ModelMeta) GenManagedObjApi(w *bufio.Writer, tmpl *template.Template) error {
	return m.getTemplate(tmpl, "managed_api", tmManagedObjApi).Execute(w, m)
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
