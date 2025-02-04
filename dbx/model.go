package dbx

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/ml444/gkit/dbx/pagination"
	"github.com/ml444/gkit/log"
)

type cipherKind int

const (
	cipherKindEncrypt cipherKind = 0
	cipherKindDecrypt cipherKind = 1
)

type ICipher interface {
	Encrypt(plaintext any) (any, error)
	Decrypt(ciphertext any) (any, error)
}

type IModel interface {
	ToORM() ITModel
}
type ITModel interface {
	ToSource() IModel
	ForceTModel() bool
}

type GenerateIDFunc func() uint64

type T struct {
	getDB           func() *gorm.DB
	model           interface{}
	ormModel        interface{}
	forceTModel     bool
	NotFoundErrCode int32
	BatchCreateSize int
	PrimaryKey      string

	// encrypt config
	EncryptFieldMap map[string]ICipher // {dbFieldName:structFieldName}
	// tableEncryptor  ICipher
	NeedEncrypt    bool
	DisableDecrypt bool

	IdGenerator GenerateIDFunc
}

func NewT[M any](fn func() *gorm.DB, opts ...TOption) *T {
	t := T{
		getDB:           fn,
		BatchCreateSize: 100,
		PrimaryKey:      "id",
		EncryptFieldMap: map[string]ICipher{},
	}
	m := new(M)

	im, ok := any(m).(IModel)
	if ok {
		t.model = m
		ormM := im.ToORM()
		t.ormModel = ormM
		t.forceTModel = ormM.ForceTModel()
	} else {
		t.model = m
		t.ormModel = m
	}

	t.init()
	for _, opt := range opts {
		opt(&t)
	}
	if t.NeedEncrypt && len(t.EncryptFieldMap) == 0 {
		log.Warn("no fields was found that needed to be encrypted")
		t.NeedEncrypt = false
	}

	return &t
}

func (x *T) init() {
	t := reflect.TypeOf(x.ormModel)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("model must be struct, but got %v", t.Kind()))
	}

	var pkList []string
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
		if strings.Contains(strings.ToLower(gormTag), "primaryKey") {
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				pkList = append(pkList, strings.Split(jsonTag, ",")[0])
			} else {
				pkList = append(pkList, camelToSnake(field.Name))
			}
		}

	}
	// NOTE: Composite primary keys are not processed for now.
	if len(pkList) == 1 {
		x.PrimaryKey = pkList[0]
	}
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

//	func (x *T) CloneTModel() interface{} {
//		mT := reflect.TypeOf(x.ormModel)
//		if mT.Kind() == reflect.Ptr {
//			mT = mT.Elem()
//		}
//		return reflect.New(mT).Interface()
//	}
func (x *T) getModel() interface{} {
	var mT reflect.Type
	if x.forceTModel {
		mT = reflect.TypeOf(x.ormModel)
	} else {
		mT = reflect.TypeOf(x.model)
	}
	if mT.Kind() == reflect.Ptr {
		mT = mT.Elem()
	}
	return reflect.New(mT).Interface()
}

func (x *T) Scope() *Scope {
	return NewScope(x.getDB(), x.getModel())
}

func (x *T) Create(m interface{}) (err error) {
	err = x.CheckAndCrypto(m, cipherKindEncrypt, true)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return err
	}
	if x.forceTModel {
		if im, ok := (m).(IModel); ok {
			ormModel := im.ToORM()
			err = x.Scope().Create(ormModel)
			if err != nil {
				log.Errorf("err: %v\n", err)
				return err
			}
			return copier.Copy(m, ormModel)
		}
	}
	return x.Scope().Create(m)
}

func (x *T) BatchCreate(list any) (err error) {
	err = x.CheckAndCrypto(list, cipherKindEncrypt, true)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return err
	}
	listV := reflect.Indirect(reflect.ValueOf(list))
	if listV.Kind() == reflect.Interface {
		listV = reflect.ValueOf(listV.Interface())
	}
	switch listV.Kind() {
	case reflect.Array, reflect.Slice:
		mT := listV.Type().Elem()
		if x.forceTModel && mT == reflect.TypeOf(x.model) {
			ormList := reflect.New(reflect.SliceOf(reflect.TypeOf(x.ormModel)))
			for i := 0; i < listV.Len(); i++ {
				ormElement := listV.Index(i).Interface().(IModel).ToORM()
				newOrmList := reflect.Append(ormList.Elem(), reflect.ValueOf(ormElement))
				ormList.Elem().Set(newOrmList)
			}
			valList := ormList.Interface()
			err = x.Scope().CreateInBatches(valList, x.BatchCreateSize)
			if err != nil {
				log.Errorf("err: %v\n", err)
				return err
			}
			return copier.Copy(list, valList)
		}
		return x.Scope().CreateInBatches(list, x.BatchCreateSize)
	default:
		return x.Scope().Create(list)
	}
}

func (x *T) Save(m interface{}) (err error) {
	err = x.CheckAndCrypto(m, cipherKindEncrypt, true)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return err
	}
	if x.forceTModel {
		if im, ok := (m).(IModel); ok {
			ormModel := im.ToORM()
			err = x.Scope().Save(ormModel)
			if err != nil {
				log.Errorf("err: %v\n", err)
				return err
			}
			return copier.Copy(m, ormModel)
		}
	}
	return x.Scope().Save(m)
}

func (x *T) Update(m interface{}, query interface{}, args ...interface{}) (rows int64, err error) {
	err = x.CheckAndCrypto(m, cipherKindEncrypt, false)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return 0, err
	}
	scope := x.Scope()
	if query != nil {
		err = x.CheckAndCrypto(query, cipherKindEncrypt, false)
		if err != nil {
			log.Errorf("err: %v\n", err)
			return 0, err
		}
		scope = scope.Where(query, args...)
	}
	if im, ok := (m).(IModel); ok && x.forceTModel {
		err = scope.Update(im.ToORM())
	} else {
		err = scope.Update(m)
	}
	return scope.RowsAffected, err
}

func (x *T) UpdateByPk(m interface{}, pk any) (rows int64, err error) {
	if x.PrimaryKey == "" {
		return 0, errors.New("unable to find a unique primary key field")
	}
	return x.Update(m, x.PrimaryKey, pk)
}

func (x *T) DeleteByPk(pk any) error {
	if x.PrimaryKey == "" {
		return errors.New("unable to find a unique primary key field")
	}
	return x.Scope().Eq(x.PrimaryKey, pk).Delete()
}

func (x *T) DeleteByWhere(query interface{}, args ...interface{}) error {
	err := x.CheckAndCrypto(query, cipherKindEncrypt, false)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return err
	}
	return x.Scope().Where(query, args...).Delete()
}

func (x *T) ExistByWhere(args ...interface{}) (bool, error) {
	count, err := x.Count(args...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (x *T) Count(args ...interface{}) (int64, error) {
	if len(args) == 0 {
		return x.Scope().Count()
	}
	query := args[0]
	args = args[1:]
	err := x.CheckAndCrypto(query, cipherKindEncrypt, false)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return 0, err
	}
	return x.Scope().Where(query, args...).Count()
}

func (x *T) GetOne(pk any, m interface{}) (err error) {
	if x.PrimaryKey == "" {
		return errors.New("unable to find a unique primary key field")
	}
	return x.GetOneByWhere(m, x.PrimaryKey, pk)
}

func (x *T) GetOneByWhere(m interface{}, query interface{}, args ...interface{}) (err error) {
	tx := NewScope(x.getDB(), m)
	if query != nil {
		err = x.CheckAndCrypto(query, cipherKindEncrypt, false)
		if err != nil {
			return err
		}
		tx = tx.Where(query, args...)
	}
	if x.NotFoundErrCode != 0 {
		tx = tx.SetNotFoundErr(x.NotFoundErrCode)
	}
	if im, ok := (m).(IModel); ok && x.forceTModel {
		mV := im.ToORM()
		err = tx.First(mV)
		if err != nil {
			log.Errorf("err: %v\n", err)
			return err
		}
		err = copier.Copy(m, mV)
	} else {
		err = tx.First(m)
	}
	if err != nil {
		log.Errorf("err: %v\n", err)
		return err
	}
	if !x.DisableDecrypt {
		err = x.CheckAndCrypto(m, cipherKindDecrypt, false)
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

func (x *T) processOpts(scope *Scope, opts interface{}) (*Scope, error) {
	if opts == nil {
		return scope, nil
	}
	switch o := opts.(type) {
	case *Scope:
		return o, nil
	case Scope:
		return &o, nil
	case QueryOpts:
		if len(o.Where) > 0 {
			err := x.CheckAndCrypto(o.Where, cipherKindEncrypt, false)
			if err != nil {
				return scope, err
			}
		}
		scope = scope.Query(&o)
	case *QueryOpts:
		if len(o.Where) > 0 {
			err := x.CheckAndCrypto(o.Where, cipherKindEncrypt, false)
			if err != nil {
				return scope, err
			}
		}
		scope = scope.Query(o)
	case map[string]any:
		if len(o) > 0 {
			err := x.CheckAndCrypto(o, cipherKindEncrypt, false)
			if err != nil {
				return scope, err
			}
			scope = scope.Where(o)
		}
	default:
		optsV := reflect.ValueOf(opts)
		if optsV.Kind() == reflect.Map {
			if optsV.Len() > 0 {
				err := x.CheckAndCrypto(o, cipherKindEncrypt, false)
				if err != nil {
					return scope, err
				}
				scope = scope.Where(o)
			}
			return scope, nil
		}
		err := fmt.Errorf("unknown opts type: %v", reflect.TypeOf(opts))
		return scope, err
	}
	return scope, nil
}

func (x *T) doBeforce(opts interface{}, listPtr interface{}) (scope *Scope, valList interface{}, needCopy bool, err error) {
	m, err := x.validateListAndGetModel(listPtr)
	if err != nil {
		log.Error(err)
		return
	}

	if x.forceTModel && !reflect.DeepEqual(m, x.ormModel) {
		needCopy = true
		valList = reflect.New(reflect.SliceOf(reflect.TypeOf(x.ormModel))).Interface()
	} else {
		valList = listPtr
	}
	scope, err = x.processOpts(x.Scope(), opts)
	if err != nil {
		log.Error(err)
		return
	}
	return
}

func (x *T) doAfter(needCopy bool, listPtr, valList interface{}) (err error) {
	if needCopy {
		err = copier.CopyWithOption(listPtr, valList, copier.Option{IgnoreEmpty: true})
		if err != nil {
			log.Error(err)
			return err
		}
	}

	if !x.DisableDecrypt {
		err = x.CheckAndCrypto(listPtr, cipherKindDecrypt, false)
	}
	return err
}

func (x *T) ListAll(opts interface{}, listPtr interface{}) error {
	scope, valList, needCopy, err := x.doBeforce(opts, listPtr)
	if err != nil {
		log.Error(err)
		return err
	}

	err = scope.Find(valList)
	if err != nil {
		log.Error(err)
		return err
	}

	return x.doAfter(needCopy, listPtr, valList)
}

func (x *T) ListWithPagination(paginate *pagination.Pagination, opts any, listPtr any) (*pagination.Pagination, error) {
	scope, valList, needCopy, err := x.doBeforce(opts, listPtr)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var newPagination *pagination.Pagination
	newPagination, err = scope.PaginationQuery(paginate, listPtr)
	if err != nil {
		return nil, err
	}
	return newPagination, x.doAfter(needCopy, listPtr, valList)
}

func (x *T) CheckAndCrypto(m any, kind cipherKind, isCreate bool) error {
	if m == nil || (len(x.EncryptFieldMap) == 0 && !isCreate) {
		return nil
	}
	mV := reflect.ValueOf(m)
	if mV.Kind() == reflect.Ptr {
		mV = mV.Elem()
	}

	switch mV.Kind() {
	case reflect.Struct:
		if isCreate && x.IdGenerator != nil && x.PrimaryKey != "" {
			pkValue := mV.FieldByName(x.PrimaryKey)
			if pkValue.IsZero() && pkValue.CanSet() {
				pkValue.SetUint(x.IdGenerator())
			}
		}
		for fieldName, cipher := range x.EncryptFieldMap {
			fieldValue := mV.FieldByName(fieldName)
			if fieldValue.Kind() == reflect.Interface {
				fieldValue = fieldValue.Elem()
			}
			if !fieldValue.IsValid() || fieldValue.IsZero() {
				continue
			}
			if fieldValue.CanSet() {
				cryptoFunc := cipher.Encrypt
				if kind == cipherKindDecrypt {
					cryptoFunc = cipher.Decrypt
				}
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
			err := x.CheckAndCrypto(mV.Index(i).Interface(), kind, isCreate)
			if err != nil {
				return err
			}
		}
	case reflect.Map:
		if isCreate && x.IdGenerator != nil && x.PrimaryKey != "" {
			mV.SetMapIndex(reflect.ValueOf(x.PrimaryKey), reflect.ValueOf(x.IdGenerator()))
		}
		for dbField, cipher := range x.EncryptFieldMap {
			fieldValue := mV.MapIndex(reflect.ValueOf(dbField))
			if fieldValue.Kind() == reflect.Interface {
				fieldValue = fieldValue.Elem()
			}
			if !fieldValue.IsValid() || fieldValue.IsZero() {
				continue
			}

			cryptoFunc := cipher.Encrypt
			if kind == cipherKindDecrypt {
				cryptoFunc = cipher.Decrypt
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

// TOption is a function that takes a pointer to a T and modifies it.
type TOption func(*T)

func SetCreateBatchSize(size int) TOption {
	return func(t *T) {
		t.BatchCreateSize = size
	}
}

// SetTableCipher sets the Encryptor field of the table.
// If only the table-level encryptor is configured,
// the encryptor is used to encrypt all fields that need to be encrypted
func SetTableCipher(tableCipher ICipher) TOption {
	return func(x *T) {
		mT := reflect.TypeOf(x.ormModel)
		if mT.Kind() == reflect.Ptr {
			mT = mT.Elem()
		}

		if mT.Kind() != reflect.Struct {
			panic(fmt.Sprintf("model must be struct, but got %v", mT.Kind()))
		}

		for i := 0; i < mT.NumField(); i++ {
			field := mT.Field(i)

			if !field.IsExported() {
				continue
			}

			gormTag := field.Tag.Get("gorm")
			if gormTag == "" {
				continue
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
					var dbFieldName string
					if jsonTag := field.Tag.Get("json"); jsonTag != "" {
						dbFieldName = strings.Split(jsonTag, ",")[0]
					} else {
						dbFieldName = camelToSnake(field.Name)
					}
					if x.EncryptFieldMap == nil {
						x.EncryptFieldMap = make(map[string]ICipher)
					}
					x.EncryptFieldMap[dbFieldName] = tableCipher
					x.EncryptFieldMap[field.Name] = tableCipher
				}
			}
		}
	}
}

// SetDisableDecrypt enable decryption of encrypted fields when get data.
func SetDisableDecrypt() TOption {
	return func(t *T) {
		t.DisableDecrypt = true
	}
}

type FieldCipher struct {
	StructField string
	Cipher      ICipher
}

// SetSpecifyFieldCipherMap Set a cipher that specifies different ciphers
// to be used for different fields, and stores both database fields and
// db fields in EncryptFieldMap.
// {dbFieldName: FieldCipher}
func SetSpecifyFieldCipherMap(fieldMap map[string]FieldCipher) TOption {
	return func(t *T) {
		if len(fieldMap) == 0 {
			return
		}
		if t.EncryptFieldMap == nil {
			t.EncryptFieldMap = make(map[string]ICipher)
		}
		t.NeedEncrypt = true
		for dbFieldName, fc := range fieldMap {
			t.EncryptFieldMap[dbFieldName] = fc.Cipher

			if fc.StructField != "" {
				t.EncryptFieldMap[fc.StructField] = fc.Cipher
			}
		}
	}
}

// SetNotFoundErrCode set the NotFoundErrCode when record not found
func SetNotFoundErrCode(notFoundErrCode int32) TOption {
	return func(t *T) {
		t.NotFoundErrCode = notFoundErrCode
	}
}

// SetPrimaryKey is a function that takes a string and sets the PrimaryKey field of the T struct
func SetPrimaryKey(pk string) TOption {
	return func(t *T) {
		t.PrimaryKey = pk
	}
}

func SetGenerateIDFunc(idFunc GenerateIDFunc) TOption {
	return func(t *T) {
		t.IdGenerator = idFunc
	}
}
