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

func (s *Scope) PaginationQuery(list any, page, size uint32) (*pagination.Pagination, error) {
	return s.PaginationQueryWithOpt(list, s.HandlePagination(page, size))
}

// PaginationQueryWithOpt runs paginated find using standard LIMIT/OFFSET.
func (s *Scope) PaginationQueryWithOpt(list any, opt *pagination.Pagination) (*pagination.Pagination, error) {
	if s == nil || s.driver == nil {
		return nil, errors.New("invalid scope or transaction")
	}
	if opt == nil {
		opt = pagination.NewDefaultPagination()
	}
	normalizePagination(opt)

	var total int64
	if !opt.SkipCount {
		var err error
		total, err = s.Count()
		if err != nil {
			return nil, err
		}
	}

	offset := opt.Offset()
	limit := int(opt.Size)
	b := s.builder.Clone()
	b.Limit = limit
	b.Offset = offset
	if err := s.driver.Find(s.context(), b, list); err != nil {
		return nil, err
	}

	return &pagination.Pagination{
		Page:      opt.Page,
		Size:      opt.Size,
		Total:     total,
		SkipCount: opt.SkipCount,
	}, nil
}

func (s *Scope) QueryPagination(list any, page, size uint32, skipCount bool) (total int64, err error) {
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

func (s *Scope) ScrollQuery(list any, cursor string, size uint32, cursorField ...string) (*pagination.Scroll, error) {
	if s == nil || s.driver == nil {
		return nil, errors.New("invalid scope or transaction")
	}
	field := "id"
	if len(cursorField) > 0 && cursorField[0] != "" {
		field = cursorField[0]
	}
	opt := s.HandlePagination(1, size)
	limit := int(opt.Size)

	ns := s.Order(field + " ASC").Limit(limit)
	if cursor != "" {
		cursorVal, err := parseScrollCursor(cursor)
		if err != nil {
			return nil, err
		}
		ns = ns.Where(field+" > ?", cursorVal)
	}
	b := ns.builder.Clone()
	if err := ns.driver.Find(ns.context(), b, list); err != nil {
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

func fieldByColumn(v reflect.Value, column string) reflect.Value {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if col := columnNameFromTag(field); col != "" && strings.EqualFold(col, column) {
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

func columnNameFromTag(field reflect.StructField) string {
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		return strings.Split(jsonTag, ",")[0]
	}
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
