package mock

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/ml444/gkit/dbx"
)

// Driver is an in-memory dbx.Driver for unit tests.
type Driver struct {
	mu     sync.Mutex
	Tables map[string][]map[string]any
	Calls  []string

	TxErr     error
	FindErr   error
	CreateErr error
}

// New creates a mock driver with empty tables.
func New() *Driver {
	return &Driver{Tables: map[string][]map[string]any{}}
}

// Seed inserts rows into a table.
func (m *Driver) Seed(table string, rows ...map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, row := range rows {
		copied := map[string]any{}
		for k, v := range row {
			copied[k] = v
		}
		m.Tables[table] = append(m.Tables[table], copied)
	}
}

func (m *Driver) record(method string) {
	m.Calls = append(m.Calls, method)
}

func (m *Driver) table(b *dbx.QueryBuilder) string {
	if b.Table != "" {
		return b.Table
	}
	return dbxTableName(b.Model)
}

func dbxTableName(model any) string {
	if model == nil {
		return ""
	}
	if t, ok := model.(interface{ TableName() string }); ok {
		return t.TableName()
	}
	return ""
}

func (m *Driver) filterRows(table string, b *dbx.QueryBuilder) []map[string]any {
	rows := append([]map[string]any(nil), m.Tables[table]...)
	for _, w := range b.Wheres {
		rows = filterByWhere(rows, w)
	}
	if b.Limit > 0 && len(rows) > b.Limit {
		rows = rows[:b.Limit]
	}
	return rows
}

func filterByWhere(rows []map[string]any, w dbx.WhereClause) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if matchWhere(row, w) {
			out = append(out, row)
		}
	}
	return out
}

func matchWhere(row map[string]any, w dbx.WhereClause) bool {
	if w.Query == "deleted_at = 0" {
		v, ok := row["deleted_at"]
		if !ok {
			return true
		}
		switch n := v.(type) {
		case int:
			return n == 0
		case int64:
			return n == 0
		case uint32:
			return n == 0
		case uint64:
			return n == 0
		}
	}
	if len(w.Args) == 1 {
		if eq, ok := parseEq(w.Query); ok {
			return reflect.DeepEqual(row[eq], w.Args[0])
		}
	}
	return true
}

func parseEq(q string) (string, bool) {
	const suffix = " = ?"
	if len(q) > len(suffix) && q[len(q)-len(suffix):] == suffix {
		return q[:len(q)-len(suffix)], true
	}
	return "", false
}

func (m *Driver) Find(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("Find")
	if m.FindErr != nil {
		return m.FindErr
	}
	table := m.table(b)
	rows := m.filterRows(table, b)
	return assignRows(dest, rows)
}

func (m *Driver) First(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("First")
	if m.FindErr != nil {
		return m.FindErr
	}
	table := m.table(b)
	rows := m.filterRows(table, b)
	if len(rows) == 0 {
		return dbx.ErrRecordNotFound
	}
	return assignRow(dest, rows[0])
}

func (m *Driver) Count(ctx context.Context, b *dbx.QueryBuilder) (int64, error) {
	m.record("Count")
	table := m.table(b)
	return int64(len(m.filterRows(table, b))), nil
}

func (m *Driver) Create(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("Create")
	if m.CreateErr != nil {
		return 0, m.CreateErr
	}
	table := m.table(b)
	row, err := structToMap(v)
	if err != nil {
		return 0, err
	}
	m.Tables[table] = append(m.Tables[table], row)
	return 1, nil
}

func (m *Driver) CreateInBatches(ctx context.Context, b *dbx.QueryBuilder, values any, batchSize int) (int64, error) {
	v := reflect.ValueOf(values)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	var n int64
	for i := 0; i < v.Len(); i++ {
		rows, err := m.Create(ctx, b, v.Index(i).Interface())
		n += rows
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (m *Driver) Save(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	return m.Create(ctx, b, v)
}

func (m *Driver) Update(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	m.record("Update")
	return 1, nil
}

func (m *Driver) UpdateColumn(ctx context.Context, b *dbx.QueryBuilder, field string, value any) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("UpdateColumn")
	table := m.table(b)
	rows := m.filterRows(table, b)
	if len(rows) == 0 {
		return 0, nil
	}
	for _, idx := range findRowIndexes(m.Tables[table], rows[0]) {
		if b.IncrColumn != "" {
			cur, _ := toInt64(m.Tables[table][idx][b.IncrColumn])
			m.Tables[table][idx][b.IncrColumn] = cur + b.IncrValue
		} else {
			m.Tables[table][idx][field] = value
		}
	}
	return int64(len(rows)), nil
}

func findRowIndexes(all []map[string]any, target map[string]any) []int {
	var idxs []int
	for i, row := range all {
		if reflect.DeepEqual(row, target) {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

func toInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	case uint:
		return int64(n), true
	case uint64:
		return int64(n), true
	case uint32:
		return int64(n), true
	default:
		return 0, false
	}
}

func (m *Driver) Delete(ctx context.Context, b *dbx.QueryBuilder) (int64, error) {
	m.record("Delete")
	return 1, nil
}

func (m *Driver) Scan(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	return m.First(ctx, b, dest)
}

func (m *Driver) Transaction(ctx context.Context, fn func(dbx.Driver) error, opts ...dbx.TxOption) error {
	if m.TxErr != nil {
		return m.TxErr
	}
	return fn(m)
}

func assignRows(dest any, rows []map[string]any) error {
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be pointer")
	}
	sv := dv.Elem()
	if sv.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be pointer to slice")
	}
	elemType := sv.Type().Elem()
	isPtr := elemType.Kind() == reflect.Ptr
	if isPtr {
		elemType = elemType.Elem()
	}
	out := reflect.MakeSlice(sv.Type(), 0, len(rows))
	for _, row := range rows {
		val := reflect.New(elemType)
		if err := mapToStruct(row, val.Interface()); err != nil {
			return err
		}
		if isPtr {
			out = reflect.Append(out, val)
		} else {
			out = reflect.Append(out, val.Elem())
		}
	}
	sv.Set(out)
	return nil
}

func assignRow(dest any, row map[string]any) error {
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be pointer")
	}
	return mapToStruct(row, dest)
}

func structToMap(v any) (map[string]any, error) {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct")
	}
	out := map[string]any{}
	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		col := f.Name
		if j := f.Tag.Get("json"); j != "" {
			col = splitTag(j)
		}
		out[col] = rv.Field(i).Interface()
	}
	return out, nil
}

func mapToStruct(row map[string]any, dest any) error {
	rv := reflect.ValueOf(dest).Elem()
	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		col := f.Name
		if j := f.Tag.Get("json"); j != "" {
			col = splitTag(j)
		}
		val, ok := row[col]
		if !ok {
			continue
		}
		fv := rv.Field(i)
		src := reflect.ValueOf(val)
		if !src.IsValid() || !src.Type().ConvertibleTo(fv.Type()) {
			return fmt.Errorf("cannot convert %T to %s", val, fv.Type())
		}
		fv.Set(src.Convert(fv.Type()))
	}
	return nil
}

func splitTag(tag string) string {
	for i, c := range tag {
		if c == ',' {
			return tag[:i]
		}
	}
	return tag
}

// Conn wraps the mock driver as dbx.Conn.
type Conn struct {
	d *Driver
}

func (m *Driver) Conn() dbx.Conn { return &Conn{d: m} }

func (c *Conn) Driver(ctx context.Context) dbx.Driver { return c.d }

// Begin implements dbx.TxManager.
func (c *Conn) Begin(ctx context.Context, opts ...dbx.TxOption) (dbx.Driver, error) {
	return c.d, nil
}

func (c *Conn) Commit(d dbx.Driver) error { return nil }

func (c *Conn) Rollback(d dbx.Driver) error { return nil }

var _ dbx.TxManager = (*Conn)(nil)

// WithContext satisfies optional context binding.
func (m *Driver) WithContext(ctx context.Context) dbx.Driver { return m }
