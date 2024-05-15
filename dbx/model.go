package dbx

import (
	"reflect"
	"strings"
	"unicode"

	"gorm.io/gorm"

	"github.com/ml444/gkit/dbx/paging"
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
		if !strings.Contains(gormTag, "primarykey") {
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
	if reflect.TypeOf(m) == reflect.TypeOf(x.tModel) {
		return x.Scope().Create(m)
	} else {
		return NewScope(x.getDB(), m).Create(m)
	}
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

func (x *T) Update(m interface{}, whereMap map[string]interface{}) error {
	if im, ok := (m).(IModel); ok {
		return x.Scope().Where(whereMap).Update(im.ToORM())
	}
	return x.Scope().Where(whereMap).Update(m)
}

func (x *T) DeleteById(pk uint64) error {
	return x.Scope().Eq("id", pk).Delete()
}

func (x *T) DeleteByWhere(whereMap map[string]interface{}) error {
	return x.Scope().Where(whereMap).Delete()
}

func (x *T) ExistByWhere(whereMap map[string]interface{}) (bool, error) {
	count, err := x.Scope().Where(whereMap).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (x *T) Count(whereMap map[string]interface{}) (int64, error) {
	if len(whereMap) == 0 {
		return x.Scope().Count()
	}
	return x.Scope().Where(whereMap).Count()
}

func (x *T) GetOne(pk uint64) (interface{}, error) {
	tm := x.CloneTModel()
	err := x.Scope().SetNotFoundErr(x.NotFoundErrCode).Eq("id", pk).First(tm)
	if res, ok := tm.(ITModel); ok {
		return res.ToSource(), err
	}
	return tm, err
}

func (x *T) GetOneByWhere(whereMap map[string]interface{}) (interface{}, error) {
	m := x.CloneTModel()
	err := x.Scope().SetNotFoundErr(x.NotFoundErrCode).Where(whereMap).First(&m)
	if err != nil {
		return nil, err
	}
	if res, ok := m.(ITModel); ok {
		return res.ToSource(), nil
	}
	return m, nil
}

func (x *T) ListAll(opts interface{}) (interface{}, error) {
	scope := x.Scope()
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
	listType := reflect.SliceOf(reflect.TypeOf(x.tModel))
	tList := reflect.New(listType)
	err := scope.Find(tList.Interface())
	if err != nil {
		return nil, err
	}
	result := reflect.New(reflect.SliceOf(reflect.TypeOf(x.model))).Elem()
	for i := 0; i < tList.Elem().Len(); i++ {
		m := tList.Elem().Index(i).Interface()
		if res, ok := m.(ITModel); ok {
			m = res.ToSource()
		}
		result = reflect.Append(result, reflect.ValueOf(m))
	}
	return result.Interface(), nil
}

func (x *T) ListWithPaginate(paginate *paging.Paginate, opts interface{}) (interface{}, *paging.Paginate, error) {
	var err error
	scope := x.Scope()
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
	listType := reflect.SliceOf(reflect.TypeOf(x.tModel))
	tList := reflect.New(listType)
	var newPaginate *paging.Paginate
	newPaginate, err = scope.PaginateQuery(paginate, tList.Interface())
	if err != nil {
		return nil, nil, err
	}
	result := reflect.New(reflect.SliceOf(reflect.TypeOf(x.model))).Elem()
	for i := 0; i < tList.Elem().Len(); i++ {
		m := tList.Elem().Index(i).Interface()
		if res, ok := m.(ITModel); ok {
			m = res.ToSource()
		}
		result = reflect.Append(result, reflect.ValueOf(m))
	}
	return result.Interface(), newPaginate, nil
}
