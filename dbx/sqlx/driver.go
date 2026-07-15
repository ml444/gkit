package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	jmo "github.com/jmoiron/sqlx"

	"github.com/ml444/gkit/dbx"
)

// Conn wraps *sqlx.DB as dbx.Conn.
type Conn struct {
	db *jmo.DB
}

// NewConn returns dbx.Conn backed by sqlx.
func NewConn(db *jmo.DB) dbx.Conn {
	return &Conn{db: db}
}

func (c *Conn) Driver(ctx context.Context) dbx.Driver {
	return &Driver{db: c.db, ctx: ctx}
}

// DB returns the underlying *sqlx.DB.
func (c *Conn) DB() *jmo.DB {
	return c.db
}

// Begin implements dbx.TxManager.
func (c *Conn) Begin(ctx context.Context, opts ...dbx.TxOption) (dbx.Driver, error) {
	var o sql.TxOptions
	for _, opt := range opts {
		opt(&o)
	}
	tx, err := c.db.BeginTxx(ctx, &o)
	if err != nil {
		return nil, err
	}
	return &txDriver{Driver: &Driver{db: c.db, tx: tx, ctx: ctx}, tx: tx}, nil
}

func (c *Conn) Commit(d dbx.Driver) error {
	td, ok := d.(*txDriver)
	if !ok {
		return sql.ErrTxDone
	}
	return td.tx.Commit()
}

func (c *Conn) Rollback(d dbx.Driver) error {
	td, ok := d.(*txDriver)
	if !ok {
		return sql.ErrTxDone
	}
	return td.tx.Rollback()
}

type txDriver struct {
	*Driver
	tx *jmo.Tx
}

// Driver implements dbx.Driver using sqlx.
type Driver struct {
	db  *jmo.DB
	tx  *jmo.Tx
	ctx context.Context
}

func (d *Driver) WithContext(ctx context.Context) dbx.Driver {
	return &Driver{db: d.db, tx: d.tx, ctx: ctx}
}

func (d *Driver) execContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	if d.tx != nil {
		return d.tx.ExecContext(ctx, q, args...)
	}
	return d.db.ExecContext(ctx, q, args...)
}

func (d *Driver) selectContext(ctx context.Context, dest any, q string, args ...any) error {
	if d.tx != nil {
		return d.tx.SelectContext(ctx, dest, q, args...)
	}
	return d.db.SelectContext(ctx, dest, q, args...)
}

func (d *Driver) getContext(ctx context.Context, dest any, q string, args ...any) error {
	if d.tx != nil {
		return d.tx.GetContext(ctx, dest, q, args...)
	}
	return d.db.GetContext(ctx, dest, q, args...)
}

func (d *Driver) context() context.Context {
	if d.ctx != nil {
		return d.ctx
	}
	return context.Background()
}

func (d *Driver) Find(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	q, args, err := compileSelect(b)
	if err != nil {
		return err
	}
	return mapErr(d.selectContext(ctx, dest, q, args...))
}

func (d *Driver) First(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	nb := b.Clone()
	nb.Limit = 1
	q, args, err := compileSelect(nb)
	if err != nil {
		return err
	}
	return mapErr(d.getContext(ctx, dest, q, args...))
}

func (d *Driver) Count(ctx context.Context, b *dbx.QueryBuilder) (int64, error) {
	q, args, err := compileCount(b)
	if err != nil {
		return 0, err
	}
	var total int64
	err = d.getContext(ctx, &total, q, args...)
	return total, err
}

func (d *Driver) Create(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	q, args, err := compileInsert(b, v)
	if err != nil {
		return 0, err
	}
	res, err := d.execContext(d.context(), q, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (d *Driver) CreateInBatches(ctx context.Context, b *dbx.QueryBuilder, values any, batchSize int) (int64, error) {
	v := reflect.ValueOf(values)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	var total int64
	for i := 0; i < v.Len(); i++ {
		n, err := d.Create(ctx, b, v.Index(i).Interface())
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func (d *Driver) Save(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	return d.Create(ctx, b, v)
}

func (d *Driver) Update(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	q, args, err := compileUpdate(b, v)
	if err != nil {
		return 0, err
	}
	res, err := d.execContext(d.context(), q, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (d *Driver) UpdateColumn(ctx context.Context, b *dbx.QueryBuilder, field string, value any) (int64, error) {
	q, args, err := compileUpdateColumn(b, field, value)
	if err != nil {
		return 0, err
	}
	res, err := d.execContext(d.context(), q, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (d *Driver) Delete(ctx context.Context, b *dbx.QueryBuilder) (int64, error) {
	q, args, err := compileDelete(b)
	if err != nil {
		return 0, err
	}
	res, err := d.execContext(d.context(), q, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (d *Driver) Scan(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	return d.First(ctx, b, dest)
}

func (d *Driver) Transaction(ctx context.Context, fn func(dbx.Driver) error, opts ...dbx.TxOption) error {
	var o sql.TxOptions
	for _, opt := range opts {
		opt(&o)
	}
	tx, err := d.db.BeginTxx(ctx, &o)
	if err != nil {
		return err
	}
	td := &Driver{db: d.db, tx: tx, ctx: ctx}
	if err := fn(td); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return dbx.ErrRecordNotFound
	}
	return err
}

func tableName(b *dbx.QueryBuilder) string {
	if b.Table != "" {
		return b.Table
	}
	if b.Model != nil {
		if t, ok := b.Model.(interface{ TableName() string }); ok {
			return t.TableName()
		}
		return tableNameFromModel(b.Model)
	}
	return ""
}

func tableNameFromModel(model any) string {
	if model == nil {
		return ""
	}
	t := reflect.TypeOf(model)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ""
	}
	return camelToSnake(t.Name())
}

func compileSelect(b *dbx.QueryBuilder) (string, []any, error) {
	table := tableName(b)
	if table == "" {
		return "", nil, fmt.Errorf("sqlx: missing table")
	}
	cols := "*"
	if len(b.Selects) > 0 {
		cols = strings.Join(b.Selects, ", ")
	}
	var args []any
	where, wargs := compileWhere(b)
	args = append(args, wargs...)
	q := fmt.Sprintf("SELECT %s FROM %s%s", cols, table, where)
	q += compileOrder(b)
	if b.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, b.Limit)
	}
	if b.Offset > 0 {
		q += " OFFSET ?"
		args = append(args, b.Offset)
	}
	if b.ForUpdate {
		q += " FOR UPDATE"
	}
	return q, args, nil
}

func compileCount(b *dbx.QueryBuilder) (string, []any, error) {
	table := tableName(b)
	if table == "" {
		return "", nil, fmt.Errorf("sqlx: missing table")
	}
	where, args := compileWhere(b)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s%s", table, where), args, nil
}

func compileWhere(b *dbx.QueryBuilder) (string, []any) {
	var parts []string
	var args []any
	for _, w := range b.Wheres {
		parts = append(parts, w.Query)
		args = append(args, w.Args...)
	}
	for _, w := range b.OrWheres {
		parts = append(parts, w.Query)
		args = append(args, w.Args...)
	}
	if len(parts) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

func compileOrder(b *dbx.QueryBuilder) string {
	var parts []string
	for _, o := range b.OrderRaw {
		parts = append(parts, o)
	}
	for _, o := range b.Orders {
		dir := "ASC"
		if o.Desc {
			dir = "DESC"
		}
		parts = append(parts, fmt.Sprintf("%s %s", o.Field, dir))
	}
	if len(parts) == 0 {
		return ""
	}
	return " ORDER BY " + strings.Join(parts, ", ")
}

func compileInsert(b *dbx.QueryBuilder, v any) (string, []any, error) {
	table := tableName(b)
	if table == "" {
		return "", nil, fmt.Errorf("sqlx: missing table")
	}
	cols, vals, err := structColumns(v)
	if err != nil {
		return "", nil, err
	}
	ph := strings.TrimRight(strings.Repeat("?,", len(cols)), ",")
	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ","), ph)
	return q, vals, nil
}

func compileUpdate(b *dbx.QueryBuilder, v any) (string, []any, error) {
	table := tableName(b)
	if table == "" {
		return "", nil, fmt.Errorf("sqlx: missing table")
	}
	cols, vals, err := structColumns(v)
	if err != nil {
		return "", nil, err
	}
	var sets []string
	for _, c := range cols {
		sets = append(sets, c+" = ?")
	}
	where, wargs := compileWhere(b)
	args := append(vals, wargs...)
	return fmt.Sprintf("UPDATE %s SET %s%s", table, strings.Join(sets, ", "), where), args, nil
}

func compileUpdateColumn(b *dbx.QueryBuilder, field string, value any) (string, []any, error) {
	table := tableName(b)
	if table == "" {
		return "", nil, fmt.Errorf("sqlx: missing table")
	}
	where, args := compileWhere(b)
	if b.IncrColumn != "" && b.IncrValue != 0 {
		col := b.IncrColumn
		v := b.IncrValue
		expr := fmt.Sprintf("COALESCE(%s, 0) + ?", col)
		if v < 0 {
			expr = fmt.Sprintf("COALESCE(%s, 0) - ?", col)
			v = -v
		}
		args = append([]any{v}, args...)
		return fmt.Sprintf("UPDATE %s SET %s = %s%s", table, col, expr, where), args, nil
	}
	args = append([]any{value}, args...)
	return fmt.Sprintf("UPDATE %s SET %s = ?%s", table, field, where), args, nil
}

func compileDelete(b *dbx.QueryBuilder) (string, []any, error) {
	table := tableName(b)
	if table == "" {
		return "", nil, fmt.Errorf("sqlx: missing table")
	}
	where, args := compileWhere(b)
	return fmt.Sprintf("DELETE FROM %s%s", table, where), args, nil
}

func structColumns(v any) ([]string, []any, error) {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Map {
		return mapColumns(rv)
	}
	if rv.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("sqlx: expected struct or map")
	}
	var cols []string
	var vals []any
	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		col := columnName(f)
		cols = append(cols, col)
		vals = append(vals, rv.Field(i).Interface())
	}
	return cols, vals, nil
}

func mapColumns(rv reflect.Value) ([]string, []any, error) {
	var cols []string
	var vals []any
	for _, k := range rv.MapKeys() {
		cols = append(cols, fmt.Sprint(k.Interface()))
		vals = append(vals, rv.MapIndex(k).Interface())
	}
	return cols, vals, nil
}

func columnName(f reflect.StructField) string {
	if j := f.Tag.Get("json"); j != "" {
		return strings.Split(j, ",")[0]
	}
	return camelToSnake(f.Name)
}

func camelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}
