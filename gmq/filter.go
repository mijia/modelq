package gmq

import (
	"fmt"
	"strings"
)

type Filter interface {
	SqlString(alias string) string
	Params() []interface{}
	And(Filter) Filter
	Or(Filter) Filter
}

func UnitFilter(name, op string, param interface{}) Filter {
	return _UnitFilter{name: name, op: op, param: param}
}

func InFilter(name string, params []interface{}) Filter {
	return _InFilter{name: name, params: params}
}

func AndFilter(left, right Filter, others ...Filter) Filter {
	fs := make([]Filter, 2+len(others))
	fs[0] = left
	fs[1] = right
	for i := range others {
		fs[i+2] = others[i]
	}
	return _AndFilter{fs: fs}
}

func OrFilter(left, right Filter, others ...Filter) Filter {
	fs := make([]Filter, 2+len(others))
	fs[0] = left
	fs[1] = right
	for i := range others {
		fs[i+2] = others[i]
	}
	return _OrFilter{fs: fs}
}

type _UnitFilter struct {
	name  string
	op    string
	param interface{}
}

func (f _UnitFilter) SqlString(alias string) string {
	return fmt.Sprintf("%s %s ?", nameWithAlias(f.name, alias), f.op)
}

func (f _UnitFilter) Params() []interface{} {
	return []interface{}{f.param}
}

func (f _UnitFilter) And(o Filter) Filter { return AndFilter(f, o) }
func (f _UnitFilter) Or(o Filter) Filter  { return OrFilter(f, o) }
func (f _UnitFilter) String() string      { return f.SqlString("") }

type _InFilter struct {
	name   string
	params []interface{}
}

func (f _InFilter) SqlString(alias string) string {
	qMarks := make([]string, len(f.params))
	for i := range f.params {
		qMarks[i] = "?"
	}
	return fmt.Sprintf("%s IN (%s)", nameWithAlias(f.name, alias), strings.Join(qMarks, ", "))
}

func (f _InFilter) Params() []interface{} {
	return f.params
}

func (f _InFilter) And(o Filter) Filter { return AndFilter(f, o) }
func (f _InFilter) Or(o Filter) Filter  { return OrFilter(f, o) }
func (f _InFilter) String() string      { return f.SqlString("") }

type _AndFilter struct {
	fs []Filter
}

func (f _AndFilter) SqlString(alias string) string {
	subs := make([]string, len(f.fs))
	for i, sf := range f.fs {
		subs[i] = sf.SqlString(alias)
	}
	return fmt.Sprintf("(%s)", strings.Join(subs, " AND "))
}

func (f _AndFilter) Params() []interface{} {
	params := make([]interface{}, 0, len(f.fs))
	for _, sf := range f.fs {
		params = append(params, sf.Params()...)
	}
	return params
}

func (f _AndFilter) And(o Filter) Filter { return AndFilter(f, o) }
func (f _AndFilter) Or(o Filter) Filter  { return OrFilter(f, o) }
func (f _AndFilter) String() string      { return f.SqlString("") }

type _OrFilter struct {
	fs []Filter
}

func (f _OrFilter) SqlString(alias string) string {
	subs := make([]string, len(f.fs))
	for i, sf := range f.fs {
		subs[i] = sf.SqlString(alias)
	}
	return fmt.Sprintf("(%s)", strings.Join(subs, " OR "))
}

func (f _OrFilter) Params() []interface{} {
	params := make([]interface{}, 0, len(f.fs))
	for _, sf := range f.fs {
		params = append(params, sf.Params()...)
	}
	return params
}

func (f _OrFilter) And(o Filter) Filter { return AndFilter(f, o) }
func (f _OrFilter) Or(o Filter) Filter  { return OrFilter(f, o) }
func (f _OrFilter) String() string      { return f.SqlString("") }
