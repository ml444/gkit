package dbx

import (
	"errors"
	"reflect"
	"strings"
	"unicode"

	"gorm.io/gorm"

	"github.com/ml444/gkit/dbx/pagination"
)

type IModel interface {
	ToORM() ITModel
}
type ITModel interface {
	ToSource() IModel
}

type T struct {
	getDB           func() *gorm.DB
	model           interface{}
	tModel          interface{}
	NotFoundErrCode int32
	BatchCreateSize int
	PrimaryKey      string
}

func NewT(fn func() *gorm.DB, m interface{}, errCode int32) *T {
	t := T{
		getDB:           fn,
		NotFoundErrCode: errCode,
		BatchCreateSize: 100,
	}
	im, ok := (m).(IModel)
	if ok {
		t.model = im
		t.tModel = im.ToORM()
	} else {
		t.model = m
		t.tModel = m
	}
	t.PrimaryKey = findPrimaryKey(t.tModel)
	return &t
}

// findPrimaryKey function
func findPrimaryKey(m interface{}) (pk string) {
	pk = "id"
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		gormTag := field.Tag.Get("gorm")
		if !strings.Contains(strings.ToLower(gormTag), "primarykey") {
			continue
		}
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			pk = strings.Split(jsonTag, ",")[0]
			return
		}
		return camelToSnake(field.Name)
	}
	return
}

func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteByte('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func (x *T) SetCreateBatchSize(size int) {
	x.BatchCreateSize = size
}

func (x *T) CloneTModel() interface{} {
	mT := reflect.TypeOf(x.tModel)
	if mT.Kind() == reflect.Ptr {
		return reflect.New(mT.Elem()).Interface()
	}
	return reflect.New(mT).Interface()
}

func (x *T) Scope() *Scope {
	return NewScope(x.getDB(), x.CloneTModel())
}

func (x *T) Create(m interface{}) (err error) {
	return NewScope(x.getDB(), m).Create(m)
}

func (x *T) BatchCreate(list interface{}) (err error) {
	listV := reflect.Indirect(reflect.ValueOf(list))
	switch listV.Kind() {
	case reflect.Array, reflect.Slice:
		mT := listV.Type().Elem()
		if mT == reflect.TypeOf(x.tModel) {
			return x.Scope().CreateInBatches(list, x.BatchCreateSize)
		} else {
			return NewScope(x.getDB(), reflect.New(mT).Interface()).CreateInBatches(list, x.BatchCreateSize)
		}
	default:
		return x.Create(list)
	}
}

func (x *T) Update(m interface{}, whereMap map[string]interface{}) (rows int64, err error) {
	scope := x.Scope().Where(whereMap)
	if im, ok := (m).(IModel); ok {
		err = scope.Update(im.ToORM())
	} else {
		err = scope.Update(m)
	}
	rows = scope.RowsAffected
	return rows, err
}

func (x *T) DeleteByPk(pk interface{}) error {
	return NewScope(x.getDB(), x.model).Eq(x.PrimaryKey, pk).Delete()
}

func (x *T) DeleteByWhere(query interface{}, args ...interface{}) error {
	return NewScope(x.getDB(), x.model).Where(query, args...).Delete()
}

func (x *T) ExistByWhere(query interface{}, args ...interface{}) (bool, error) {
	count, err := NewScope(x.getDB(), x.model).Where(query, args...).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (x *T) Count(whereMap map[string]interface{}) (int64, error) {
	if len(whereMap) == 0 {
		return x.Scope().Count()
	}
	return NewScope(x.getDB(), x.model).Where(whereMap).Count()
}

func (x *T) GetOne(pk uint64, m interface{}) error {
	return NewScope(x.getDB(), m).SetNotFoundErr(x.NotFoundErrCode).Eq(x.PrimaryKey, pk).First(m)
}

func (x *T) GetOneByWhere(whereMap map[string]interface{}, m interface{}) error {
	return NewScope(x.getDB(), m).SetNotFoundErr(x.NotFoundErrCode).Where(whereMap).First(m)
}

func (x *T) validateListAndGetModel(listPtr interface{}) (interface{}, error) {
	listType := reflect.TypeOf(listPtr)
	if listType.Kind() != reflect.Ptr {
		return nil, errors.New("list must be pointer")
	}
	listType = listType.Elem()
	if listType.Kind() != reflect.Slice && listType.Kind() != reflect.Array {
		return nil, errors.New("list is not a slice or array")
	}
	elemType := listType.Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	m := reflect.New(elemType).Interface()
	return m, nil
}

func (x *T) ListAll(opts interface{}, listPtr interface{}) error {
	m, err := x.validateListAndGetModel(listPtr)
	if err != nil {
		return err
	}
	scope := NewScope(x.getDB(), m)
	if opts != nil {
		switch o := opts.(type) {
		case *Scope:
			scope = o
		case *QueryOpts:
			scope = scope.Query(o)
		case map[string]interface{}:
			if len(o) > 0 {
				scope = scope.Where(o)
			}
		}
	}
	return scope.Find(listPtr)
}

func (x *T) ListWithPagination(paginate *pagination.Pagination, opts interface{}, listPtr interface{}) (*pagination.Pagination, error) {
	m, err := x.validateListAndGetModel(listPtr)
	if err != nil {
		return nil, err
	}
	scope := NewScope(x.getDB(), m)
	if opts != nil {
		switch o := opts.(type) {
		case *Scope:
			scope = o
		case *QueryOpts:
			scope = scope.Query(o)
		case map[string]interface{}:
			if len(o) > 0 {
				scope = scope.Where(o)
			}
		}
	}
	var newPagination *pagination.Pagination
	newPagination, err = scope.PaginationQuery(paginate, listPtr)
	if err != nil {
		return nil, err
	}
	return newPagination, err
}
