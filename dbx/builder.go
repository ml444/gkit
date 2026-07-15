package dbx

import (
	"fmt"
	"reflect"
	"strings"
)

type WhereClause struct {
	Query string
	Args  []any
}

type joinClause struct {
	Query string
	Args  []any
}

type incrField struct {
	Field string
	Value int64
}

// QueryBuilder accumulates SQL fragments in a driver-agnostic way.
type QueryBuilder struct {
	Table    string
	Model    any
	Selects  []string
	Omits    []string
	Wheres   []WhereClause
	OrWheres []WhereClause
	Orders   []OrderColumn
	OrderRaw []string
	Groups   []string
	Having   *WhereClause
	Joins    []joinClause
	Limit    int
	Offset   int
	Unscoped bool
	Distinct bool

	IncrColumn       string
	IncrValue        int64
	ReturningColumns []string
	ForUpdate        bool
}

func newQueryBuilder(modelOrTable any) *QueryBuilder {
	b := &QueryBuilder{}
	switch v := modelOrTable.(type) {
	case string:
		b.Table = v
	case ITable:
		b.Table = v.TableName()
	default:
		b.Model = modelOrTable
	}
	return b
}

func tableNameFromModel(model any) string {
	if model == nil {
		return ""
	}
	if t, ok := model.(ITable); ok {
		return t.TableName()
	}
	t := reflect.TypeOf(model)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ""
	}
	name := t.Name()
	var b strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

func (b *QueryBuilder) Clone() *QueryBuilder {
	if b == nil {
		return &QueryBuilder{}
	}
	c := *b
	c.Selects = append([]string(nil), b.Selects...)
	c.Omits = append([]string(nil), b.Omits...)
	c.Wheres = append([]WhereClause(nil), b.Wheres...)
	c.OrWheres = append([]WhereClause(nil), b.OrWheres...)
	c.Orders = append([]OrderColumn(nil), b.Orders...)
	c.OrderRaw = append([]string(nil), b.OrderRaw...)
	c.Groups = append([]string(nil), b.Groups...)
	c.Joins = append([]joinClause(nil), b.Joins...)
	c.ReturningColumns = append([]string(nil), b.ReturningColumns...)
	if b.Having != nil {
		h := *b.Having
		h.Args = append([]any(nil), b.Having.Args...)
		c.Having = &h
	}
	return &c
}

func (b *QueryBuilder) addWhere(query string, args ...any) {
	b.Wheres = append(b.Wheres, WhereClause{Query: query, Args: args})
}

func (b *QueryBuilder) addOrWhere(query string, args ...any) {
	b.OrWheres = append(b.OrWheres, WhereClause{Query: query, Args: args})
}

func (b *QueryBuilder) addMapWhere(m map[string]any) {
	for k, v := range m {
		b.addWhere(fmt.Sprintf("%s = ?", k), v)
	}
}

func (b *QueryBuilder) removeSoftDeleteFilter() {
	out := b.Wheres[:0]
	for _, w := range b.Wheres {
		if w.Query == "deleted_at = 0" && len(w.Args) == 0 {
			continue
		}
		out = append(out, w)
	}
	b.Wheres = out
}
