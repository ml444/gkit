package dbx

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"gorm.io/gorm"

	"github.com/ml444/gkit/dbx/pagination"
	"github.com/ml444/gkit/log"
)

type IModel interface {
	ToORM() ITModel
}
type ITModel interface {
	ToSource() IModel
}

type ICypher interface {
	Encrypt(plaintext any) (any, error)
	Decrypt(ciphertext any) (any, error)
}

type T struct {
	getDB           func() *gorm.DB
	model           interface{}
	tModel          interface{}
	NotFoundErrCode int32
	BatchCreateSize int
	PrimaryKey      string

	// encrypt config
	EncryptStructFieldMap map[string]bool // {structFieldName:true}
	EncryptDBFieldMap     map[string]bool // {dbFieldName:true}
	Encryptor             ICypher
	NeedEncrypt           bool
	NeedDecrypt           bool
}

func NewT(fn func() *gorm.DB, m interface{}, errCode int32, opts ...TOption) *T {
	t := T{
		getDB:           fn,
		NotFoundErrCode: errCode,
		BatchCreateSize: 100,
		PrimaryKey:      "id",
	}
	im, ok := (m).(IModel)
	if ok {
		t.model = im
		t.tModel = im.ToORM()
	} else {
		t.model = m
		t.tModel = m
	}
	t.init()
	for _, opt := range opts {
		opt(&t)
	}
	if (t.NeedEncrypt || t.NeedDecrypt) && t.Encryptor == nil {
		panic("Encryptor is nil, must set Encryptor")
	}
	return &t
}

func (x *T) init() {
	t := reflect.TypeOf(x.tModel)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("model must be struct, but got %v", t.Kind()))
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		gormTag := field.Tag.Get("gorm")
		if gormTag == "" {
			continue
		}
		// Handle primary key
		if strings.Contains(strings.ToLower(gormTag), "primarykey") {
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				x.PrimaryKey = strings.Split(jsonTag, ",")[0]
			} else {
				x.PrimaryKey = camelToSnake(field.Name)
			}
		}

		// Find encrypted fields
		for _, s := range strings.Split(gormTag, ";") {
			ss := strings.Split(s, ":")
			if len(ss) == 2 && strings.TrimSpace(ss[0]) == "encrypt" {
				isTrue, err := strconv.ParseBool(ss[1])
				if err != nil {
					println(err.Error())
					continue
				}
				if !isTrue {
					continue
				}
				x.NeedEncrypt = true
				if x.EncryptStructFieldMap == nil {
					x.EncryptStructFieldMap = make(map[string]bool)
				}
				if x.EncryptDBFieldMap == nil {
					x.EncryptDBFieldMap = make(map[string]bool)
				}
				if jsonTag := field.Tag.Get("json"); jsonTag != "" {
					dbFieldName := strings.Split(jsonTag, ",")[0]
					x.EncryptStructFieldMap[field.Name] = true
					x.EncryptDBFieldMap[dbFieldName] = true
				} else {
					dbFieldName := camelToSnake(field.Name)
					x.EncryptStructFieldMap[field.Name] = true
					x.EncryptDBFieldMap[dbFieldName] = true
				}
			}
		}
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
		mT = mT.Elem()
	}
	return reflect.New(mT).Interface()
}

func (x *T) Scope() *Scope {
	return NewScope(x.getDB(), x.CloneTModel())
}

func (x *T) Create(m interface{}) (err error) {
	err = x.CheckAndCrypto(m, x.Encryptor.Encrypt)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return err
	}
	return NewScope(x.getDB(), m).Create(m)
}

func (x *T) BatchCreate(list interface{}) (err error) {
	err = x.CheckAndCrypto(list, x.Encryptor.Encrypt)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return err
	}
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
	err = x.CheckAndCrypto(m, x.Encryptor.Encrypt)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return 0, err
	}
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
	tx := NewScope(x.getDB(), m)
	if x.NotFoundErrCode != 0 {
		tx = tx.SetNotFoundErr(x.NotFoundErrCode)
	}
	err := tx.Eq(x.PrimaryKey, pk).First(m)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return err
	}
	if x.NeedDecrypt {
		err = x.CheckAndCrypto(m, x.Encryptor.Decrypt)
	}
	return err
}

func (x *T) GetOneByWhere(whereMap map[string]interface{}, m interface{}) (err error) {
	tx := NewScope(x.getDB(), m)
	if x.NotFoundErrCode != 0 {
		tx = tx.SetNotFoundErr(x.NotFoundErrCode)
	}
	err = tx.Where(whereMap).First(m)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return err
	}
	if x.NeedDecrypt {
		err = x.CheckAndCrypto(m, x.Encryptor.Decrypt)
	}
	return err
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
	err = scope.Find(listPtr)
	if err != nil {
		return err
	}
	if x.NeedDecrypt {
		err = x.CheckAndCrypto(listPtr, x.Encryptor.Decrypt)
	}
	return nil
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
	if x.NeedDecrypt {
		err = x.CheckAndCrypto(listPtr, x.Encryptor.Decrypt)
	}
	return newPagination, err
}

func (x *T) CheckAndCrypto(m interface{}, cryptoFunc func(s any) (any, error)) error {
	if !x.NeedEncrypt {
		return nil
	}
	mV := reflect.ValueOf(m)
	if mV.Kind() == reflect.Ptr {
		mV = mV.Elem()
	}

	switch mV.Kind() {
	case reflect.Struct:
		for fieldName, ok := range x.EncryptStructFieldMap {
			if !ok {
				continue
			}
			fieldValue := mV.FieldByName(fieldName)
			if fieldValue.Kind() == reflect.Interface {
				fieldValue = fieldValue.Elem()
			}
			if !fieldValue.IsValid() || fieldValue.IsZero() {
				continue
			}
			if fieldValue.CanSet() {
				encryptedValue, err := cryptoFunc(fieldValue.Interface())
				if err != nil {
					println("crypto field err: ", err.Error())
					return err
				}
				fieldValue.Set(reflect.ValueOf(encryptedValue))
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < mV.Len(); i++ {
			err := x.CheckAndCrypto(mV.Index(i).Interface(), cryptoFunc)
			if err != nil {
				return err
			}
		}
	case reflect.Map:
		for dbField, ok := range x.EncryptDBFieldMap {
			if !ok {
				continue
			}
			fieldValue := mV.MapIndex(reflect.ValueOf(dbField))
			if fieldValue.Kind() == reflect.Interface {
				fieldValue = fieldValue.Elem()
			}
			if !fieldValue.IsValid() || fieldValue.IsZero() {
				continue
			}

			encryptedValue, err := cryptoFunc(fieldValue.Interface())
			if err != nil {
				println("crypto field err: ", err.Error())
				return err
			}
			mV.SetMapIndex(reflect.ValueOf(dbField), reflect.ValueOf(encryptedValue))
		}
	}

	return nil
}

func (x *T) Encrypt(plaintext any) (any, error) {
	if x.Encryptor == nil {
		return nil, errors.New("encryptor is nil, must set encryptor first")
	}
	return x.Encryptor.Encrypt(plaintext)
}
func (x *T) Decrypt(ciphertext any) (any, error) {
	if x.Encryptor == nil {
		return nil, errors.New("encryptor is nil, must set encryptor first")
	}
	return x.Encryptor.Decrypt(ciphertext)
}

// TOption is a function that takes a pointer to a T and modifies it.
type TOption func(*T)

// WithEncryptor is a function that takes an ICypher and sets the Encryptor field of the T struct
func WithEncryptor(encryptor ICypher) TOption {
	return func(t *T) {
		t.Encryptor = encryptor
	}
}

// WithNeedEncrypt enable some specified fields to be encrypted
func WithNeedEncrypt(needEncrypt bool) TOption {
	return func(t *T) {
		t.NeedEncrypt = needEncrypt
	}
}

// WithNeedDecrypt enable decryption of encrypted fields when get data.
func WithNeedDecrypt(needDecrypt bool) TOption {
	return func(t *T) {
		t.NeedDecrypt = needDecrypt
	}
}

// WithEncryptFieldMap is a function that takes a map of struct field names
// to database field names and sets the encryptStructFieldMap field of the T struct.
// {structFieldName: dbFieldName}
func WithEncryptFieldMap(encryptStructFieldMap map[string]string) TOption {
	return func(t *T) {
		if len(encryptStructFieldMap) == 0 {
			return
		}
		if t.EncryptStructFieldMap == nil {
			t.EncryptStructFieldMap = make(map[string]bool)
		}
		if t.EncryptDBFieldMap == nil {
			t.EncryptDBFieldMap = make(map[string]bool)
		}
		t.NeedEncrypt = true
		for fieldName, dbField := range encryptStructFieldMap {
			t.EncryptDBFieldMap[dbField] = true
			t.EncryptStructFieldMap[fieldName] = true
		}
	}
}

// WithNotFoundErrCode set the NotFoundErrCode when record not found
func WithNotFoundErrCode(notFoundErrCode int32) TOption {
	return func(t *T) {
		t.NotFoundErrCode = notFoundErrCode
	}
}

// WithPrimaryKey is a function that takes a string and sets the PrimaryKey field of the T struct
func WithPrimaryKey(pk string) TOption {
	return func(t *T) {
		t.PrimaryKey = pk
	}
}
