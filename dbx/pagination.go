package dbx

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/ml444/gkit/dbx/pagination"
)

const (
	DefaultLimit int = 2000
	MaxLimit     int = 100000
)

func (s *Scope) PaginationQuery(list interface{}, page, size uint32) (*pagination.Pagination, error) {
	return s.PaginationQueryWithOpt(list, s.HandlePagination(page, size))
}

// PaginationQueryWithOpt runs paginated find using a pre-built Pagination option.
func (s *Scope) PaginationQueryWithOpt(list interface{}, opt *pagination.Pagination) (*pagination.Pagination, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("invalid scope or transaction")
	}
	if opt == nil {
		opt = pagination.NewDefaultPagination()
	}
	normalizePagination(opt)

	var total int64
	if !opt.SkipCount {
		if err := s.DB.Count(&total).Error; err != nil {
			return nil, err
		}
	}

	// Apply limit/offset on a chained query without mutating s.DB, so the
	// same Scope can be safely reused (e.g. for count or another page).
	offset := opt.Offset()
	if err := s.DB.Limit(int(opt.Size)).Offset(offset).Find(list).Error; err != nil {
		return nil, err
	}

	return &pagination.Pagination{
		Page:      opt.Page,
		Size:      opt.Size,
		Total:     total,
		SkipCount: opt.SkipCount,
	}, nil
}

// QueryPagination is deprecated; use PaginationQuery or PaginationQueryWithOpt.
func (s *Scope) QueryPagination(list interface{}, page, size uint32, skipCount bool) (total int64, err error) {
	opt := s.HandlePagination(page, size)
	opt.SkipCount = skipCount
	p, err := s.PaginationQueryWithOpt(list, opt)
	if err != nil {
		return 0, err
	}
	return p.Total, nil
}

func (s *Scope) HandlePagination(page, size uint32) *pagination.Pagination {
	opt := &pagination.Pagination{
		Page: page,
		Size: size,
	}
	normalizePagination(opt)
	return opt
}

func normalizePagination(opt *pagination.Pagination) {
	if opt.Page == 0 {
		opt.Page = 1
	}
	if opt.Size == 0 {
		opt.Size = uint32(DefaultLimit)
	} else if opt.Size > uint32(MaxLimit) {
		opt.Size = uint32(MaxLimit)
	}
}

// ScrollQuery fetches the next page using keyset pagination on the primary key column (default "id").
func (s *Scope) ScrollQuery(list interface{}, cursor string, size uint32, cursorField ...string) (*pagination.Scroll, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("invalid scope or transaction")
	}
	field := "id"
	if len(cursorField) > 0 && cursorField[0] != "" {
		field = cursorField[0]
	}
	opt := s.HandlePagination(1, size)
	limit := int(opt.Size)

	q := s.DB.Order(field + " ASC").Limit(limit)
	if cursor != "" {
		cursorVal, err := parseScrollCursor(cursor)
		if err != nil {
			return nil, err
		}
		q = q.Where(field+" > ?", cursorVal)
	}
	if err := q.Find(list).Error; err != nil {
		return nil, err
	}

	scroll := &pagination.Scroll{Size: opt.Size}
	listV := reflect.ValueOf(list)
	if listV.Kind() == reflect.Ptr {
		listV = listV.Elem()
	}
	if listV.Kind() == reflect.Slice && listV.Len() > 0 {
		last := listV.Index(listV.Len() - 1)
		if last.Kind() == reflect.Ptr {
			last = last.Elem()
		}
		if last.Kind() == reflect.Struct {
			// Resolve the cursor from the same column used for ORDER BY/WHERE
			// (default "id"), so a custom cursorField yields the correct cursor.
			cursorVal := fieldByColumn(last, field)
			if cursorVal.IsValid() {
				scroll.Cursor = formatScrollCursor(cursorVal)
			}
		}
	}
	return scroll, nil
}

func parseScrollCursor(cursor string) (any, error) {
	if cursor == "" {
		return nil, nil
	}
	if id, err := strconv.ParseUint(cursor, 10, 64); err == nil {
		return id, nil
	}
	return cursor, nil
}

// fieldByColumn finds the struct field that maps to the given DB column name.
// Match order:
//  1. explicit gorm:"column:..." tag
//  2. camelToSnake(field) == column (disambiguates e.g. foo_sar vs foos_ar)
//  3. field.Name == snakeToPascal(column) (case-sensitive)
//  4. simple names without underscores, e.g. column "id" -> field ID
func fieldByColumn(v reflect.Value, column string) reflect.Value {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if col := gormColumnName(field); col != "" && strings.EqualFold(col, column) {
			return v.Field(i)
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.EqualFold(camelToSnake(field.Name), column) {
			return v.Field(i)
		}
	}

	target := snakeToPascal(column)
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == target {
			return v.Field(i)
		}
	}

	if !strings.Contains(column, "_") {
		for i := 0; i < t.NumField(); i++ {
			if strings.EqualFold(t.Field(i).Name, column) {
				return v.Field(i)
			}
		}
	}

	return reflect.Value{}
}

func gormColumnName(field reflect.StructField) string {
	tag := field.Tag.Get("gorm")
	if tag == "" {
		return ""
	}
	for _, part := range strings.Split(tag, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		if len(p) == 1 {
			b.WriteString(strings.ToUpper(p))
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		b.WriteString(strings.ToLower(p[1:]))
	}
	return b.String()
}

func formatScrollCursor(field reflect.Value) string {
	switch field.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(field.Uint(), 10)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(field.Int(), 10)
	case reflect.String:
		return field.String()
	default:
		return ""
	}
}
