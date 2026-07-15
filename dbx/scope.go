package dbx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/log"
)

const (
	ProtoMessageFieldCreatedAt = "created_at"
	ProtoMessageFieldUpdatedAt = "updated_at"
	ProtoMessageFieldDeletedAt = "deleted_at"
)

type ModelBase interface {
	ProtoReflect() protoreflect.Message
	GetId() uint64
}

type ProtoUpsertedAt interface {
	ProtoReflect() protoreflect.Message
	GetCreatedAt() int64
	GetUpdatedAt() int64
	ProtoDeletedAt
}

type ProtoUpdatedAt interface {
	GetUpdatedAt() int64
}

type ProtoDeletedAt interface {
	GetDeletedAt() int64
}

type ITable interface {
	TableName() string
}

type OrderColumn struct {
	Field   string
	Desc    bool
	Reorder bool
}

type Scope struct {
	conn   Conn
	driver Driver
	ctx    context.Context

	builder *QueryBuilder
	model   any

	NotFoundErr  error
	RowsAffected int64

	resetTime         bool
	ignoreNotFoundErr bool
	includeDeleted    bool
}

func NewScope(conn Conn, modelOrTable any) *Scope {
	return newScope(conn, conn.Driver(context.Background()), context.Background(), modelOrTable, false)
}

func NewScopeOfPure(conn Conn, model any) *Scope {
	return newScope(conn, conn.Driver(context.Background()), context.Background(), model, true)
}

func newScopeWithDriver(conn Conn, driver Driver, ctx context.Context, modelOrTable any) *Scope {
	return newScope(conn, driver, ctx, modelOrTable, false)
}

func newScope(conn Conn, driver Driver, ctx context.Context, modelOrTable any, pure bool) *Scope {
	if modelOrTable == nil {
		panic("modelOrTable is nil")
	}
	b := newQueryBuilder(modelOrTable)
	if !pure {
		if _, ok := modelOrTable.(ProtoDeletedAt); ok {
			b.addWhere("deleted_at = 0")
		}
	}
	return &Scope{
		conn:    conn,
		driver:  driver,
		ctx:     ctx,
		builder: b,
		model:   modelOrTable,
	}
}

func (s *Scope) fork() *Scope {
	if s == nil {
		return nil
	}
	ns := *s
	ns.builder = s.builder.Clone()
	return &ns
}

func (s *Scope) context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}

// Context returns the scope context.
func (s *Scope) Context() context.Context {
	return s.context()
}

func (s *Scope) SetNotFoundErr(notFoundErrCode int32) *Scope {
	ns := s.fork()
	ns.NotFoundErr = errorx.New(notFoundErrCode)
	return ns
}

func (s *Scope) IgnoreNotFoundErr() *Scope {
	ns := s.fork()
	ns.ignoreNotFoundErr = true
	return ns
}

func (s *Scope) ResetSysDateTimeField(v any) {
	if m, ok := v.(ProtoUpsertedAt); ok {
		fields := m.ProtoReflect().Descriptor().Fields()
		m.ProtoReflect().Set(fields.ByName(ProtoMessageFieldCreatedAt), protoreflect.ValueOfUint32(0))
		m.ProtoReflect().Set(fields.ByName(ProtoMessageFieldUpdatedAt), protoreflect.ValueOfUint32(0))
		m.ProtoReflect().Set(fields.ByName(ProtoMessageFieldDeletedAt), protoreflect.ValueOfUint32(0))
	}
}

func (s *Scope) Create(v any) error {
	if s.resetTime {
		s.ResetSysDateTimeField(v)
	}
	b := s.builder.Clone()
	rows, err := s.driver.Create(s.context(), b, v)
	s.RowsAffected = rows
	return err
}

func (s *Scope) Save(v any) error {
	if s.resetTime {
		s.ResetSysDateTimeField(v)
	}
	b := s.builder.Clone()
	rows, err := s.driver.Save(s.context(), b, v)
	s.RowsAffected = rows
	return err
}

func (s *Scope) CreateInBatches(values any, batchSize int) error {
	b := s.builder.Clone()
	rows, err := s.driver.CreateInBatches(s.context(), b, values, batchSize)
	s.RowsAffected = rows
	return err
}

func (s *Scope) Update(v any, conds ...any) error {
	ns := s.fork()
	if ns.resetTime {
		ns.ResetSysDateTimeField(v)
	} else {
		if vv, okk := v.(map[string]any); okk {
			if _, ok := ns.model.(ProtoUpdatedAt); ok {
				vv[ProtoMessageFieldUpdatedAt] = time.Now().Unix()
			}
		}
	}
	if len(conds) > 0 {
		ns = ns.Where(conds[0], conds[1:]...)
	}
	b := ns.builder.Clone()
	rows, err := ns.driver.Update(ns.context(), b, v)
	ns.RowsAffected = rows
	*s = *ns
	if err != nil {
		return err
	}
	if s.RowsAffected == 0 {
		log.Warnf("model: %v, RowsAffected: 0", v)
	}
	return nil
}

func (s *Scope) UpdateColumn(field string, value any) error {
	b := s.builder.Clone()
	rows, err := s.driver.UpdateColumn(s.context(), b, field, value)
	s.RowsAffected = rows
	if err != nil {
		return err
	}
	if s.RowsAffected == 0 {
		log.Warnf("value: %v, RowsAffected: 0", value)
	}
	return nil
}

func (s *Scope) UpdateColumnWithIncr(field string, v int64) error {
	if v == 0 {
		return nil
	}
	b := s.builder.Clone()
	b.IncrColumn = field
	b.IncrValue = v
	rows, err := s.driver.UpdateColumn(s.context(), b, field, nil)
	s.RowsAffected = rows
	if err != nil {
		return err
	}
	if s.RowsAffected == 0 {
		log.Warnf("model: %v, RowsAffected: 0", v)
		return ErrUpdateRowAffectedZero
	}
	return nil
}

func (s *Scope) Delete(conds ...any) error {
	if _, ok := s.model.(ProtoDeletedAt); ok && !s.includeDeleted {
		ns := s.fork()
		if len(conds) > 0 {
			ns = ns.Where(conds[0], conds[1:]...)
		}
		return ns.UpdateColumn(ProtoMessageFieldDeletedAt, time.Now().Unix())
	}
	ns := s.fork()
	if len(conds) > 0 {
		ns = ns.Where(conds[0], conds[1:]...)
	}
	b := ns.builder.Clone()
	rows, err := ns.driver.Delete(ns.context(), b)
	s.RowsAffected = rows
	return err
}

func (s *Scope) handleNotFound(err error) error {
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrRecordNotFound) {
		return err
	}
	if s.ignoreNotFoundErr {
		return nil
	}
	if s.NotFoundErr != nil {
		return s.NotFoundErr
	}
	return err
}

func (s *Scope) First(dest any, conds ...any) error {
	ns := s.fork()
	if len(conds) > 0 {
		ns = ns.Where(conds[0], conds[1:]...)
	}
	b := ns.builder.Clone()
	err := MapDriverError(ns.driver.First(ns.context(), b, dest))
	return ns.handleNotFound(err)
}

func (s *Scope) Scan(dest any, conds ...any) error {
	ns := s.fork()
	if len(conds) > 0 {
		ns = ns.Where(conds[0], conds[1:]...)
	}
	b := ns.builder.Clone()
	err := MapDriverError(ns.driver.Scan(ns.context(), b, dest))
	return ns.handleNotFound(err)
}

func (s *Scope) Exist(conds ...any) (bool, error) {
	ns := s.fork()
	if len(conds) > 0 {
		ns = ns.Where(conds[0], conds[1:]...)
	}
	n, err := ns.Count()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Scope) Find(dest any, conds ...any) error {
	ns := s.fork()
	if len(conds) > 0 {
		ns = ns.Where(conds[0], conds[1:]...)
	}
	b := ns.builder.Clone()
	return ns.driver.Find(ns.context(), b, dest)
}

func (s *Scope) Select(fields ...string) *Scope {
	ns := s.fork()
	ns.builder.Selects = append(ns.builder.Selects, fields...)
	return ns
}

func (s *Scope) Like(field string, value string) *Scope {
	return s.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
}

func (s *Scope) LikePrefix(field string, value string) *Scope {
	return s.Where(fmt.Sprintf("%s LIKE ?", field), value+"%")
}

func (s *Scope) LikeSuffix(field string, value string) *Scope {
	return s.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value)
}

func (s *Scope) NotLike(field string, value string) *Scope {
	return s.Where(fmt.Sprintf("%s NOT LIKE ?", field), "%"+value+"%")
}

func (s *Scope) IsNull(field string) *Scope {
	return s.Where(fmt.Sprintf("%s IS NULL", field))
}

func (s *Scope) IsNotNull(field string) *Scope {
	return s.Where(fmt.Sprintf("%s IS NOT NULL", field))
}

func (s *Scope) Where(query any, args ...any) *Scope {
	ns := s.fork()
	switch q := query.(type) {
	case map[string]any:
		ns.builder.addMapWhere(q)
	case map[string]string:
		m := make(map[string]any, len(q))
		for k, v := range q {
			m[k] = v
		}
		ns.builder.addMapWhere(m)
	case string:
		ns.builder.addWhere(q, args...)
	default:
		rv := reflect.ValueOf(query)
		if rv.IsValid() && rv.Kind() == reflect.Map {
			iter := rv.MapRange()
			for iter.Next() {
				ns.builder.addWhere(fmt.Sprintf("%s = ?", iter.Key().String()), iter.Value().Interface())
			}
		} else {
			ns.builder.addWhere(fmt.Sprint(query), args...)
		}
	}
	return ns
}

type OrWhereItem struct {
	Field string
	Value any
}

type QueryOpts struct {
	Selects        []string
	Where          map[string]any
	Between        map[string][2]any
	Like           map[string]string
	Or             []WhereClause
	OrLike         [][2]string	// [["field1","abc"],["field2", "def"]]
	OrBetween      map[string][2]any
	IsLikePrefix   bool
	IsOrLikePrefix bool
	GroupBys       []string
	OrderBys       []OrderColumn
	OrderBy        string
	Exp            []any
}

func (s *Scope) Query(opts *QueryOpts) *Scope {
	if opts == nil {
		return s
	}
	if len(opts.Selects) > 0 {
		s = s.Select(opts.Selects...)
	}
	if len(opts.Where) > 0 {
		s = s.Where(opts.Where)
	}
	if len(opts.Between) > 0 {
		for field, value := range opts.Between {
			s = s.Between(field, value[0], value[1])
		}
	}
	if len(opts.Like) > 0 {
		for field, value := range opts.Like {
			if opts.IsLikePrefix {
				s = s.LikePrefix(field, value)
			} else {
				s = s.Like(field, value)
			}
		}
	}
	if len(opts.Or) > 0 {
		s = s.MultiOr(opts.Or)
	}
	if len(opts.OrLike) > 0 {
		s = s.MultiOrLike(opts.OrLike, opts.IsOrLikePrefix)
	}
	if len(opts.OrBetween) > 0 {
		var queryList []string
		var valueList []any
		for field, values := range opts.OrBetween {
			queryList = append(queryList, fmt.Sprintf("(%s BETWEEN ? AND ?)", field))
			valueList = append(valueList, values[0], values[1])
		}
		query := strings.Join(queryList, " OR ")
		s = s.Where(query, valueList...)
	}
	if len(opts.GroupBys) > 0 {
		s = s.Groups(opts.GroupBys...)
	}
	if len(opts.OrderBys) > 0 {
		s = s.Orders(opts.OrderBys...)
	}
	if opts.OrderBy != "" {
		s = s.Order(opts.OrderBy)
	}
	if len(opts.Exp) > 0 {
		s = s.Where(opts.Exp[0], opts.Exp[1:]...)
	}
	return s
}

func isNonEmptySlice(v any) bool {
	_v := reflect.ValueOf(v)
	return _v.Kind() == reflect.Slice && _v.Len() > 0
}

func (s *Scope) In(field string, values any) *Scope {
	if !isNonEmptySlice(values) {
		return s
	}
	return s.Where(fmt.Sprintf("%s IN ?", field), values)
}

func (s *Scope) NotIn(field string, values any) *Scope {
	if !isNonEmptySlice(values) {
		return s
	}
	return s.Where(fmt.Sprintf("%s NOT IN ?", field), values)
}

func (s *Scope) Ne(field string, arg any) *Scope {
	return s.Where(fmt.Sprintf("%s != ?", field), arg)
}

func (s *Scope) Eq(field string, arg any) *Scope {
	return s.Where(fmt.Sprintf("%s = ?", field), arg)
}

func (s *Scope) Gt(field string, arg any) *Scope {
	return s.Where(fmt.Sprintf("%s > ?", field), arg)
}

func (s *Scope) Gte(field string, arg any) *Scope {
	return s.Where(fmt.Sprintf("%s >= ?", field), arg)
}

func (s *Scope) Lt(field string, arg any) *Scope {
	return s.Where(fmt.Sprintf("%s < ?", field), arg)
}

func (s *Scope) Lte(field string, arg any) *Scope {
	return s.Where(fmt.Sprintf("%s <= ?", field), arg)
}

func (s *Scope) Between(field string, arg1, arg2 any) *Scope {
	return s.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), arg1, arg2)
}

func (s *Scope) NotBetween(field string, arg1, arg2 any) *Scope {
	return s.Where(fmt.Sprintf("%s NOT BETWEEN ? AND ?", field), arg1, arg2)
}

func (s *Scope) Or(query any, args ...any) *Scope {
	ns := s.fork()
	switch q := query.(type) {
	case string:
		ns.builder.addOrWhere(q, args...)
	default:
		ns.builder.addOrWhere(fmt.Sprint(q), args...)
	}
	return ns
}

func (s *Scope) MultiOr(opts []WhereClause) *Scope {
	var values []any
	var orQueryList []string
	for _, opt := range opts {
		if !strings.Contains(opt.Query, "?") && len(opt.Args) > 0 {
			orQueryList = append(orQueryList, fmt.Sprintf("(%s = ?)", opt.Query))
			values = append(values, opt.Args...)
		} else {
			orQueryList = append(orQueryList, fmt.Sprintf("(%s)", opt.Query))
			values = append(values, opt.Args...)
		}
	}
	finalQuery := fmt.Sprintf("(%s)", strings.Join(orQueryList, " OR "))
	return s.Where(finalQuery, values...)
}

func (s *Scope) MultiOrLike(opts [][2]string, isPrefix bool) *Scope {
	var values []any
	var orQueryList []string
	for _, opt := range opts {
		// Note: Please ensure that opt[0] (field name) does not contain 
		// special characters, especially backticks `
		orQueryList = append(orQueryList, fmt.Sprintf("(`%s` LIKE ?)", opt[0]))
		if isPrefix {
			values = append(values, opt[1]+"%")
		} else {
			values = append(values, "%"+opt[1]+"%")
		}
	}
	finalQuery := fmt.Sprintf("(%s)", strings.Join(orQueryList, " OR "))
	return s.Where(finalQuery, values...)
}

func (s *Scope) Order(value string) *Scope {
	ns := s.fork()
	ns.builder.OrderRaw = append(ns.builder.OrderRaw, value)
	return ns
}

func (s *Scope) Orders(values ...OrderColumn) *Scope {
	ns := s.fork()
	ns.builder.Orders = append(ns.builder.Orders, values...)
	return ns
}

func (s *Scope) Group(name string) *Scope {
	ns := s.fork()
	ns.builder.Groups = append(ns.builder.Groups, name)
	return ns
}

func (s *Scope) Groups(names ...string) *Scope {
	ns := s.fork()
	ns.builder.Groups = append(ns.builder.Groups, names...)
	return ns
}

func (s *Scope) Having(query any, args ...any) *Scope {
	ns := s.fork()
	switch q := query.(type) {
	case string:
		ns.builder.Having = &WhereClause{Query: q, Args: args}
	default:
		ns.builder.Having = &WhereClause{Query: fmt.Sprint(q), Args: args}
	}
	return ns
}

func (s *Scope) Joins(query string, args ...any) *Scope {
	ns := s.fork()
	ns.builder.Joins = append(ns.builder.Joins, joinClause{Query: query, Args: args})
	return ns
}

func (s *Scope) Count() (int64, error) {
	b := s.builder.Clone()
	return s.driver.Count(s.context(), b)
}

func (s *Scope) Offset(offset int) *Scope {
	ns := s.fork()
	ns.builder.Offset = offset
	return ns
}

func (s *Scope) Limit(limit int) *Scope {
	ns := s.fork()
	ns.builder.Limit = limit
	return ns
}

func (s *Scope) Omit(value ...string) *Scope {
	ns := s.fork()
	ns.builder.Omits = append(ns.builder.Omits, value...)
	return ns
}

func (s *Scope) ReturnColumns(columns ...string) *Scope {
	ns := s.fork()
	ns.builder.ReturningColumns = append(ns.builder.ReturningColumns, columns...)
	return ns
}

func (s *Scope) WithContext(ctx context.Context) *Scope {
	ns := s.fork()
	ns.ctx = ctx
	if ns.conn != nil {
		ns.driver = ns.conn.Driver(ctx)
	}
	return ns
}

// Builder returns a clone of the scope query builder (for driver extensions).
func (s *Scope) Builder() *QueryBuilder {
	if s == nil || s.builder == nil {
		return &QueryBuilder{}
	}
	return s.builder.Clone()
}

// Conn returns the connection associated with this scope.
func (s *Scope) Conn() Conn {
	return s.conn
}

// Model returns the bound model or table reference.
func (s *Scope) Model() any {
	return s.model
}

// SetIncludeDeleted marks the scope to include soft-deleted rows (used by gorm Unscoped).
func (s *Scope) SetIncludeDeleted() *Scope {
	ns := s.fork()
	ns.includeDeleted = true
	ns.builder.removeSoftDeleteFilter()
	return ns
}

// SetForUpdate requests row-level locking where the driver supports it.
func (s *Scope) SetForUpdate() *Scope {
	ns := s.fork()
	ns.builder.ForUpdate = true
	return ns
}

func (s *Scope) Transaction(fn func(Driver) error, opts ...TxOption) error {
	return s.driver.Transaction(s.context(), fn, opts...)
}
