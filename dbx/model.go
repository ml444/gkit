package dbx

import (
	"context"
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

type cryptoField struct {
	index  int
	cipher ICipher
}

type T struct {
	getDB             func() *gorm.DB
	model             any
	ormModel          any
	forceTModel       bool
	IgnoreNotFoundErr bool
	NotFoundErrCode   int32
	BatchCreateSize   int
	PrimaryKey        string
	primaryKeyIndex   int

	// encrypt config
	EncryptFieldMap map[string]ICipher // db or struct field name -> cipher
	encryptDBFields map[string]ICipher // db column name -> cipher (for map queries)
	cryptoFields    []cryptoField      // struct field indices for encryption
	NeedEncrypt     bool
	DisableDecrypt  bool

	modelType    reflect.Type
	ormModelType reflect.Type
	needORMCopy  bool

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
	t.rebuildFieldCache()

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
	x.primaryKeyIndex = -1
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
		if strings.Contains(gormTag, "primaryKey") {
			x.primaryKeyIndex = i
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

func (x *T) rebuildFieldCache() {
	ormT := reflect.TypeOf(x.ormModel)
	if ormT.Kind() == reflect.Ptr {
		ormT = ormT.Elem()
	}
	modelT := reflect.TypeOf(x.model)
	if modelT.Kind() == reflect.Ptr {
		modelT = modelT.Elem()
	}
	x.ormModelType = ormT
	x.modelType = modelT
	x.needORMCopy = x.forceTModel && modelT != ormT

	x.cryptoFields = nil
	x.encryptDBFields = make(map[string]ICipher)
	if ormT.Kind() != reflect.Struct {
		return
	}

	seenIndex := make(map[int]ICipher)
	for i := 0; i < ormT.NumField(); i++ {
		field := ormT.Field(i)
		if !field.IsExported() {
			continue
		}
		dbFieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			dbFieldName = strings.Split(jsonTag, ",")[0]
		} else {
			dbFieldName = camelToSnake(field.Name)
		}

		cipher, ok := x.EncryptFieldMap[field.Name]
		if !ok {
			cipher, ok = x.EncryptFieldMap[dbFieldName]
		}
		if !ok {
			continue
		}
		if _, exists := seenIndex[i]; !exists {
			seenIndex[i] = cipher
			x.cryptoFields = append(x.cryptoFields, cryptoField{index: i, cipher: cipher})
		}
		x.encryptDBFields[dbFieldName] = cipher
	}

	for name, cipher := range x.EncryptFieldMap {
		if _, ok := x.encryptDBFields[name]; ok {
			continue
		}
		if _, ok := ormT.FieldByName(name); ok {
			continue
		}
		x.encryptDBFields[name] = cipher
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

func (x *T) Clone(opts ...TOption) *T {
	t := &T{
		getDB:             x.getDB,
		model:             x.model,
		ormModel:          x.ormModel,
		forceTModel:       x.forceTModel,
		IgnoreNotFoundErr: x.IgnoreNotFoundErr,
		NotFoundErrCode:   x.NotFoundErrCode,
		BatchCreateSize:   x.BatchCreateSize,
		PrimaryKey:        x.PrimaryKey,
		primaryKeyIndex:   x.primaryKeyIndex,
		EncryptFieldMap:   x.EncryptFieldMap,
		encryptDBFields:   x.encryptDBFields,
		cryptoFields:      x.cryptoFields,
		NeedEncrypt:       x.NeedEncrypt,
		DisableDecrypt:    x.DisableDecrypt,
		modelType:         x.modelType,
		ormModelType:      x.ormModelType,
		needORMCopy:       x.needORMCopy,
		IdGenerator:       x.IdGenerator,
	}
	for _, opt := range opts {
		opt(t)
	}
	if len(opts) > 0 {
		t.rebuildFieldCache()
	}
	return t
}

func (x *T) getModel() any {
	mT := x.modelType
	if x.forceTModel {
		mT = x.ormModelType
	}
	return reflect.New(mT).Interface()
}

func (x *T) Scope(ctx context.Context) *Scope {
	return NewScope(x.getDB(), x.getModel()).WithContext(ctx)
}

func (x *T) Create(ctx context.Context, m any, omitFields ...string) (err error) {
	if err = x.CheckAndCrypto(m, cipherKindEncrypt, true); err != nil {
		return err
	}
	scope := x.Scope(ctx)
	if len(omitFields) > 0 {
		scope.Omit(omitFields...)
	}
	if x.forceTModel {
		if im, ok := (m).(IModel); ok {
			ormModel := im.ToORM()
			if err = scope.Create(ormModel); err != nil {
				return err
			}
			return copier.Copy(m, ormModel)
		}
	}
	return scope.Create(m)
}

func (x *T) BatchCreate(ctx context.Context, list any, omitFields ...string) (err error) {
	if err = x.CheckAndCrypto(list, cipherKindEncrypt, true); err != nil {
		return err
	}
	listV := reflect.Indirect(reflect.ValueOf(list))
	if listV.Kind() == reflect.Interface {
		listV = reflect.ValueOf(listV.Interface())
	}
	scope := x.Scope(ctx)
	if len(omitFields) > 0 {
		scope.Omit(omitFields...)
	}
	switch listV.Kind() {
	case reflect.Array, reflect.Slice:
		mT := listV.Type().Elem()
		elemType := mT
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
		if x.forceTModel && elemType == x.modelType {
			ormElemType := reflect.TypeOf(x.ormModel)
			ormList := reflect.MakeSlice(reflect.SliceOf(ormElemType), listV.Len(), listV.Len())
			for i := 0; i < listV.Len(); i++ {
				ormElement := listV.Index(i).Interface().(IModel).ToORM()
				ormList.Index(i).Set(reflect.ValueOf(ormElement))
			}
			valList := ormList.Interface()
			if err = scope.CreateInBatches(valList, x.BatchCreateSize); err != nil {
				return err
			}
			return copier.Copy(list, valList)
		}
		return scope.CreateInBatches(list, x.BatchCreateSize)
	default:
		return scope.Create(list)
	}
}

func (x *T) Save(ctx context.Context, m any, omitFields ...string) (err error) {
	if err = x.CheckAndCrypto(m, cipherKindEncrypt, true); err != nil {
		return err
	}
	scope := x.Scope(ctx)
	if len(omitFields) > 0 {
		scope.Omit(omitFields...)
	}
	if x.forceTModel {
		if im, ok := (m).(IModel); ok {
			ormModel := im.ToORM()
			if err = scope.Save(ormModel); err != nil {
				return err
			}
			return copier.Copy(m, ormModel)
		}
	}
	return scope.Save(m)
}

func (x *T) Update(ctx context.Context, m any, query any, args ...any) (rows int64, err error) {
	if err = x.CheckAndCrypto(m, cipherKindEncrypt, false); err != nil {
		return 0, err
	}
	scope := x.Scope(ctx)
	if query != nil {
		if err = x.CheckAndCrypto(query, cipherKindEncrypt, false); err != nil {
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

func (x *T) UpdateByPk(ctx context.Context, m any, pk any, selectFields ...string) (rows int64, err error) {
	if x.PrimaryKey == "" {
		return 0, errors.New("unable to find a unique primary key field")
	}
	if err = x.CheckAndCrypto(m, cipherKindEncrypt, false); err != nil {
		return 0, err
	}
	scope := x.Scope(ctx)
	scope.Eq(x.PrimaryKey, pk)
	if len(selectFields) > 0 {
		scope.Select(selectFields...)
	}
	if im, ok := (m).(IModel); ok && x.forceTModel {
		err = scope.Update(im.ToORM())
	} else {
		err = scope.Update(m)
	}
	return scope.RowsAffected, err
}

func (x *T) DeleteByPk(ctx context.Context, pk any) error {
	if x.PrimaryKey == "" {
		return errors.New("unable to find a unique primary key field")
	}
	return x.Scope(ctx).Where(x.PrimaryKey, pk).Delete()
}

func (x *T) DeleteByWhere(ctx context.Context, query any, args ...any) error {
	if err := x.CheckAndCrypto(query, cipherKindEncrypt, false); err != nil {
		return err
	}
	return x.Scope(ctx).Where(query, args...).Delete()
}

func (x *T) ExistByWhere(ctx context.Context, args ...any) (bool, error) {
	scope := x.Scope(ctx)
	if len(args) == 0 {
		return scope.Exist()
	}
	query := args[0]
	args = args[1:]
	err := x.CheckAndCrypto(query, cipherKindEncrypt, false)
	if err != nil {
		return false, err
	}
	scope, err = x.processOpts(scope, query, args...)
	if err != nil {
		return false, err
	}
	return scope.Exist()
}

func (x *T) Count(ctx context.Context, args ...any) (int64, error) {
	if len(args) == 0 {
		return x.Scope(ctx).Count()
	}
	query := args[0]
	args = args[1:]
	if err := x.CheckAndCrypto(query, cipherKindEncrypt, false); err != nil {
		return 0, err
	}
	return x.Scope(ctx).Where(query, args...).Count()
}

func (x *T) GetOne(ctx context.Context, m any, pk any) (err error) {
	if x.PrimaryKey == "" {
		return errors.New("unable to find a unique primary key field")
	}
	return x.GetOneByWhere(ctx, m, x.PrimaryKey, pk)
}

func (x *T) GetOneByWhere(ctx context.Context, m any, query any, args ...any) (err error) {
	scope := x.Scope(ctx)
	if x.IgnoreNotFoundErr {
		scope.IgnoreNotFoundErr()
	}
	if query != nil {
		if scope, err = x.processOpts(scope, query, args...); err != nil {
			return err
		}
	}
	if !x.IgnoreNotFoundErr && x.NotFoundErrCode != 0 {
		scope = scope.SetNotFoundErr(x.NotFoundErrCode)
	}
	if im, ok := (m).(IModel); ok && x.forceTModel {
		mV := im.ToORM()
		if err = scope.First(mV); err != nil {
			return err
		}
		err = copier.Copy(m, mV)
	} else {
		err = scope.First(m)
	}
	if err != nil {
		return err
	}
	if !x.DisableDecrypt {
		err = x.CheckAndCrypto(m, cipherKindDecrypt, false)
	}
	return err
}

func (x *T) validateListAndGetModel(listPtr any) (any, error) {
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

func (x *T) processOpts(scope *Scope, opts any, args ...any) (*Scope, error) {
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
			return scope.Where(o), nil
		}
	case string:
		if o != "" {
			encArgs, encErr := x.encryptQueryArgs(args)
			if encErr != nil {
				return scope, encErr
			}
			return scope.Where(o, encArgs...), nil
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

func (x *T) doBefore(ctx context.Context, opts any, listPtr any) (scope *Scope, valList any, needCopy bool, err error) {
	if _, err = x.validateListAndGetModel(listPtr); err != nil {
		log.Error(err)
		return
	}

	if x.needORMCopy {
		needCopy = true
		valList = reflect.New(reflect.SliceOf(reflect.TypeOf(x.ormModel))).Interface()
	} else {
		valList = listPtr
	}
	scope, err = x.processOpts(x.Scope(ctx), opts)
	if err != nil {
		log.Error(err)
		return
	}
	return
}

func (x *T) doAfter(needCopy bool, listPtr, valList any) (err error) {
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

func (x *T) ListAll(ctx context.Context, listPtr any, opts any) error {
	scope, valList, needCopy, err := x.doBefore(ctx, opts, listPtr)
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

func (x *T) ListWithPagination(ctx context.Context, listPtr any, opts any, page, size uint32) (*pagination.Pagination, error) {
	scope, valList, needCopy, err := x.doBefore(ctx, opts, listPtr)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var newPagination *pagination.Pagination
	newPagination, err = scope.PaginationQuery(valList, page, size)
	if err != nil {
		return nil, err
	}
	return newPagination, x.doAfter(needCopy, listPtr, valList)
}

func (x *T) CheckAndCrypto(m any, kind cipherKind, isCreate bool) error {
	if m == nil || (len(x.cryptoFields) == 0 && len(x.encryptDBFields) == 0 && !isCreate) {
		return nil
	}
	mV := reflect.ValueOf(m)
	if mV.Kind() == reflect.Ptr {
		mV = mV.Elem()
	}

	switch mV.Kind() {
	case reflect.Struct:
		if isCreate && x.IdGenerator != nil && x.primaryKeyIndex >= 0 {
			pkValue := mV.Field(x.primaryKeyIndex)
			if pkValue.IsValid() && pkValue.IsZero() && pkValue.CanSet() && pkValue.Kind() == reflect.Uint64 {
				pkValue.SetUint(x.IdGenerator())
			}
		}
		for _, cf := range x.cryptoFields {
			fieldValue := mV.Field(cf.index)
			if fieldValue.Kind() == reflect.Interface {
				fieldValue = fieldValue.Elem()
			}
			if !fieldValue.IsValid() || fieldValue.IsZero() {
				continue
			}
			if fieldValue.CanSet() {
				cryptoFunc := cf.cipher.Encrypt
				if kind == cipherKindDecrypt {
					cryptoFunc = cf.cipher.Decrypt
				}
				encryptedValue, err := cryptoFunc(fieldValue.Interface())
				if err != nil {
					log.Errorf("crypto field err: %v", err)
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
		for dbField, cipher := range x.encryptDBFields {
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
				log.Errorf("crypto field err: %v", err)
				return err
			}
			mV.SetMapIndex(reflect.ValueOf(dbField), reflect.ValueOf(encryptedValue))
		}
	}

	return nil
}

func (x *T) encryptQueryArgs(args []any) ([]any, error) {
	if len(args) == 0 || len(x.encryptDBFields) == 0 {
		return args, nil
	}
	ciphers := x.uniqueCiphers()
	if len(ciphers) != 1 {
		return args, nil
	}
	cipher := ciphers[0]
	out := make([]any, len(args))
	for i, arg := range args {
		out[i] = arg
		if s, ok := arg.(string); ok && s != "" {
			enc, err := cipher.Encrypt(s)
			if err != nil {
				return nil, err
			}
			out[i] = enc
		}
	}
	return out, nil
}

func (x *T) uniqueCiphers() []ICipher {
	seen := make(map[ICipher]struct{})
	var ciphers []ICipher
	for _, c := range x.encryptDBFields {
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		ciphers = append(ciphers, c)
	}
	return ciphers
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
						log.Errorf("parse encrypt tag: %v", err)
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

func SetIgnoreNotFoundErr() TOption {
	return func(t *T) {
		t.IgnoreNotFoundErr = true
	}
}
