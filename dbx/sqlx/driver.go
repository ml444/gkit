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
		return d.tx.ExecContext(ctx, d.tx.Rebind(q), args...)
	}
	return d.db.ExecContext(ctx, d.db.Rebind(q), args...)
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
	if ctx == nil {
		ctx = d.context()
	}
	res, err := d.execContext(ctx, q, args...)
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
	q, args, err := compileUpsert(b, v, d.db.DriverName()) // 新增：INSERT ... ON CONFLICT/DUPLICATE
	if err != nil {
		return 0, err
	}
	res, err := d.execContext(ctx, q, args...) // 注意：用参数 ctx（见 B-sqlx-ctx）
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
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
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r) // 保留 panic 语义
		}
	}()
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
	var args []any
	var andParts, orParts []string
	for _, w := range b.Wheres {
		andParts = append(andParts, w.Query)
		args = append(args, w.Args...)
	}
	for _, w := range b.OrWheres {
		orParts = append(orParts, w.Query)
		args = append(args, w.Args...)
	}
	var groups []string
	if len(andParts) > 0 {
		groups = append(groups, strings.Join(andParts, " AND "))
	}
	if len(orParts) > 0 {
		groups = append(groups, "("+strings.Join(orParts, " OR ")+")")
	}
	if len(groups) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(groups, " AND "), args
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
		// TODO: zero value Omit
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

// camelToSnake
// Note: DisplayID → display_i_d
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

// compileUpsert 生成基于不同方言的 Upsert SQL 语句及参数
func compileUpsert(b *dbx.QueryBuilder, v any, dialect string) (string, []any, error) {
	// 1. 获取模型元数据 (你需要桥接 gkit/dbx 内部的反射/解析逻辑)
	tableName := tableName(b) // 替换为实际获取表名的方法

	// 伪方法：解析模型 v，返回所有列名、对应的值、以及主键列名
	// columns = []string{"id", "name", "age", "created_at"}
	// values = []any{1, "Alice", 25, "2023-10-01"}
	// pks = []string{"id"}
	columns, values, pks, err := parseModelForUpsert(v)
	if err != nil {
		return "", nil, err
	}
	if len(columns) == 0 {
		return "", nil, errors.New("sqlx: no columns to upsert")
	}

	// 2. 构建基础 INSERT 部分
	// 结果如: INSERT INTO users (id, name, age) VALUES (?, ?, ?)
	var query strings.Builder
	query.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES (", tableName, strings.Join(columns, ", ")))

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?" // sqlx 后续如果需要 PG 的 $1，通常由 sqlx.Rebind 处理
	}
	query.WriteString(strings.Join(placeholders, ", "))
	query.WriteString(")")

	// 3. 筛选需要 Update 的字段 (排除主键，防止更新时修改主键)
	var updateCols []string
	for _, col := range columns {
		isPk := false
		for _, pk := range pks {
			if col == pk {
				isPk = true
				break
			}
		}
		if !isPk {
			updateCols = append(updateCols, col)
		}
	}

	// 边缘情况：如果除了主键没有其他字段，执行 DO NOTHING
	if len(updateCols) == 0 {
		return compileDoNothing(query.String(), dialect)
	}

	// 4. 根据方言追加 Upsert 子句
	switch strings.ToLower(dialect) {
	case "mysql":
		// MySQL: ON DUPLICATE KEY UPDATE name=VALUES(name), age=VALUES(age)
		query.WriteString(" ON DUPLICATE KEY UPDATE ")
		var updates []string
		for _, col := range updateCols {
			updates = append(updates, fmt.Sprintf("%s=VALUES(%s)", col, col))
		}
		query.WriteString(strings.Join(updates, ", "))

	case "postgres", "postgresql", "sqlite", "sqlite3":
		// PG/SQLite 必须明确指定冲突的主键
		if len(pks) == 0 {
			return "", nil, errors.New("sqlx: upsert in postgres/sqlite requires at least one primary key")
		}
		// PG/SQLite: ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, age=EXCLUDED.age
		query.WriteString(fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET ", strings.Join(pks, ", ")))
		var updates []string
		for _, col := range updateCols {
			updates = append(updates, fmt.Sprintf("%s=EXCLUDED.%s", col, col))
		}
		query.WriteString(strings.Join(updates, ", "))

	default:
		return "", nil, fmt.Errorf("sqlx: unsupported dialect %q for upsert", dialect)
	}

	return query.String(), values, nil
}

// parseModelForUpsert 解析结构体，支持递归解析嵌套（匿名）结构体
func parseModelForUpsert(v any) (columns []string, values []any, pks []string, err error) {
	val := reflect.ValueOf(v)

	// 1. 处理指针类型
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, nil, nil, errors.New("sqlx: nil pointer passed to Save")
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, nil, nil, errors.New("sqlx: Save requires a struct or struct pointer")
	}

	// 2. 使用切片指针收集数据，方便在递归中追加
	err = extractStructFields(val, &columns, &values, &pks)
	if err != nil {
		return nil, nil, nil, err
	}

	if len(columns) == 0 {
		return nil, nil, nil, errors.New("sqlx: no valid columns found in struct")
	}

	return columns, values, pks, nil
}
// extractStructFields 递归提取结构体字段
func extractStructFields(val reflect.Value, columns *[]string, values *[]any, pks *[]string) error {
	// 确保传入的是结构体 (防备指针等)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil // 遇到空的嵌套指针结构体，直接跳过不提取
		}
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// 跳过未导出的私有字段
		if !field.IsExported() {
			continue
		}

		// 解析 db 标签
		dbTag := field.Tag.Get("db")
		if dbTag == "-" {
			continue // 忽略明确标记为 "-" 的字段，即使它是嵌套结构体也整体忽略
		}

		// 【核心逻辑】处理匿名/嵌套结构体 (例如 BaseModel)
		if field.Anonymous {
			// 如果它是结构体，或者是指向结构体的指针，递归解析
			kind := fieldVal.Kind()
			if kind == reflect.Struct || (kind == reflect.Ptr && fieldVal.Type().Elem().Kind() == reflect.Struct) {
				err := extractStructFields(fieldVal, columns, values, pks)
				if err != nil {
					return err
				}
				continue // 递归解析完毕后，继续看下一个字段
			}
		}

		// 常规字段处理逻辑
		colName := field.Name
		isPk := false

		if dbTag != "" {
			parts := strings.Split(dbTag, ",")
			if parts[0] != "" {
				colName = parts[0] // 使用 tag 指定的列名
			}
			
			// 检查是否包含 pk 标识
			for _, part := range parts[1:] {
				if strings.TrimSpace(part) == "pk" {
					isPk = true
					break
				}
			}
		} else {
			// 无 db 标签时，降级使用字段名的小写
			colName = strings.ToLower(colName)
		}

		// 补充主键识别：如果字段名为 ID/Id 且尚未标记为主键
		if !isPk && strings.EqualFold(field.Name, "ID") {
			isPk = true
		}

		// 追加数据到切片指针引用的底层数组
		*columns = append(*columns, colName)
		*values = append(*values, fieldVal.Interface())
		if isPk {
			*pks = append(*pks, colName)
		}
	}

	return nil
}

// compileDoNothing 处理只有主键时的退化情况
func compileDoNothing(baseQuery, dialect string) (string, []any, error) {
	switch strings.ToLower(dialect) {
	case "mysql":
		// MySQL 没有优雅的 DO NOTHING，通常用更新主键自身来作为 Hack
		// 注意：这会导致自增 ID 增加 (InnoDB 行为)
		return baseQuery + " ON DUPLICATE KEY UPDATE id=id", nil, nil
	case "postgres", "postgresql", "sqlite", "sqlite3":
		return baseQuery + " ON CONFLICT DO NOTHING", nil, nil
	default:
		return "", nil, fmt.Errorf("sqlx: unsupported dialect %q", dialect)
	}
}
