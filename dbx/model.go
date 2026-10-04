package dbx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/ml444/gkit/dbx/crypto"
	"github.com/ml444/gkit/dbx/pagination"
	"github.com/ml444/gkit/log"
)

type IModel interface {
	ToORM() ITModel
}
type ITModel interface {
	ToSource() IModel
	ForceTModel() bool
	CopyToSource(dst IModel) error
	CopyToSourceIgnoreEmpty(dst IModel) error
}

// ListCopyFunc copies a non-pointer ORM slice into a source slice pointer.
// Implementations should build a new result and assign dst only after all
// elements have been copied successfully.
type ListCopyFunc func(dst, src any, ignoreEmpty bool) error

// T is the repository with optional dbx/crypto field encryption.
type T struct {
	getConn           func() Conn
	model             any
	ormModel          any
	forceTModel       bool
	IgnoreNotFoundErr bool
	NotFoundErrCode   int32
	BatchCreateSize   int
	PrimaryKey        string

	crypto         *crypto.Processor
	disableDecrypt bool

	modelType    reflect.Type
	ormModelType reflect.Type
	ormListType  reflect.Type
	needORMCopy  bool
	newModel     func() any
	newORMList   func() any
	copyList     ListCopyFunc
}

// NewT constructs a T repository. Invalid crypto options return an error.
func NewT[M any](fn func() Conn, opts ...TOption) *T {
	t := &T{
		getConn:         fn,
		BatchCreateSize: 100,
		PrimaryKey:      "id",
	}
	m := new(M)
	if im, ok := any(m).(IModel); ok {
		t.model = m
		ormM := im.ToORM()
		t.ormModel = ormM
		t.forceTModel = ormM.ForceTModel()
	} else {
		t.model = m
		t.ormModel = m
	}
	t.initModelMeta()
	for _, opt := range opts {
		opt(t)
	}
	t.cacheTypes()
	return t
}

func (x *T) initModelMeta() {
	t := reflect.TypeOf(x.ormModel)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("model must be struct, but got %v", t.Kind()))
	}

	var primaryKeys []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		gormTag := field.Tag.Get("gorm")
		if isPrimaryKeyTag(gormTag) {
			primaryKeys = append(primaryKeys, modelColumnName(field, gormTag))
		}
	}
	if len(primaryKeys) > 1 {
		x.PrimaryKey = ""
	} else if len(primaryKeys) == 1 {
		x.PrimaryKey = primaryKeys[0]
	}
}

func isPrimaryKeyTag(tag string) bool {
	for _, part := range strings.Split(tag, ";") {
		key := strings.ToLower(strings.TrimSpace(strings.SplitN(part, ":", 2)[0]))
		if key == "primarykey" || key == "primary_key" {
			return true
		}
	}
	return false
}

func modelColumnName(field reflect.StructField, gormTag string) string {
	for _, part := range strings.Split(gormTag, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "column:") {
			return strings.TrimSpace(part[len("column:"):])
		}
	}
	if jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]; jsonTag != "" && jsonTag != "-" {
		return jsonTag
	}
	return camelToSnake(field.Name)
}

func (x *T) cacheTypes() {
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
	x.ormListType = reflect.SliceOf(reflect.TypeOf(x.ormModel))
	x.needORMCopy = x.forceTModel && modelT != ormT
	modelFactoryType := x.modelType
	if x.forceTModel {
		modelFactoryType = x.ormModelType
	}
	x.newModel = func() any { return reflect.New(modelFactoryType).Interface() }
	ormListType := x.ormListType
	x.newORMList = func() any { return reflect.New(ormListType).Interface() }
}

func (x *T) Clone(opts ...TOption) *T {
	t := &T{
		getConn:           x.getConn,
		model:             x.model,
		ormModel:          x.ormModel,
		forceTModel:       x.forceTModel,
		IgnoreNotFoundErr: x.IgnoreNotFoundErr,
		NotFoundErrCode:   x.NotFoundErrCode,
		BatchCreateSize:   x.BatchCreateSize,
		PrimaryKey:        x.PrimaryKey,
		crypto:            x.crypto,
		disableDecrypt:    x.disableDecrypt,
		modelType:         x.modelType,
		ormModelType:      x.ormModelType,
		ormListType:       x.ormListType,
		needORMCopy:       x.needORMCopy,
		copyList:          x.copyList,
	}
	for _, opt := range opts {
		opt(t)
	}
	t.cacheTypes()
	return t
}

func (x *T) getModel() any {
	return x.newModel()
}

func (x *T) Scope(ctx context.Context) *Scope {
	return NewScope(x.getConn(), x.getModel()).WithContext(ctx)
}

func (x *T) encryptForWrite(m any) (any, error) {
	if x.crypto == nil {
		return m, nil
	}
	return x.crypto.EncryptValue(m)
}

func (x *T) decryptAfterRead(m any) error {
	if x.crypto == nil || x.disableDecrypt {
		return nil
	}
	out, err := x.crypto.DecryptValue(m)
	if err != nil {
		return err
	}
	if out == nil || x.crypto.MutatesInPlace() {
		return nil
	}
	return assignCryptoResult(m, out)
}

func assignCryptoResult(dst, src any) error {
	dv := reflect.ValueOf(dst)
	sv := reflect.ValueOf(src)
	if !dv.IsValid() || !sv.IsValid() {
		return errors.New("assign crypto result: source and destination must be non-nil")
	}
	if dv.Kind() == reflect.Ptr && !dv.IsNil() && sv.Kind() == reflect.Ptr && !sv.IsNil() {
		if dv.Elem().Type() == sv.Elem().Type() && dv.Elem().CanSet() {
			dv.Elem().Set(sv.Elem())
			return nil
		}
	}
	if dv.Kind() == reflect.Ptr && sv.Kind() != reflect.Ptr {
		if dv.Elem().CanSet() && dv.Elem().Type() == sv.Type() {
			dv.Elem().Set(sv)
			return nil
		}
	}
	// slices / maps: if dst is pointer to slice/map, replace
	if dv.Kind() == reflect.Ptr && !dv.IsNil() && dv.Elem().CanSet() {
		if sv.Type().AssignableTo(dv.Elem().Type()) {
			dv.Elem().Set(sv)
			return nil
		}
		if sv.Kind() == reflect.Ptr && !sv.IsNil() && sv.Elem().Type().AssignableTo(dv.Elem().Type()) {
			dv.Elem().Set(sv.Elem())
			return nil
		}
	}
	return fmt.Errorf("assign crypto result: cannot assign %T to %T", src, dst)
}

func (x *T) encryptQuery(query any) (any, error) {
	if x.crypto == nil || query == nil {
		return query, nil
	}
	switch q := query.(type) {
	case map[string]any:
		return x.crypto.EncryptQueryMap(q)
	case QueryOpts:
		if len(q.Where) > 0 {
			w, err := x.crypto.EncryptQueryMap(q.Where)
			if err != nil {
				return nil, err
			}
			q.Where = w
		}
		return q, nil
	case *QueryOpts:
		if q != nil && len(q.Where) > 0 {
			w, err := x.crypto.EncryptQueryMap(q.Where)
			if err != nil {
				return nil, err
			}
			q.Where = w
		}
		return q, nil
	default:
		// Encrypt struct/map updates via EncryptValue (both modes).
		rv := reflect.ValueOf(query)
		if rv.Kind() == reflect.Map {
			enc, err := x.crypto.EncryptValue(query)
			return enc, err
		}
		return query, nil
	}
}

func (x *T) Create(ctx context.Context, m any, omitFields ...string) (err error) {
	m, err = x.encryptForWrite(m)
	if err != nil {
		return err
	}
	scope := x.Scope(ctx)
	if len(omitFields) > 0 {
		scope = scope.Omit(omitFields...)
	}
	if x.forceTModel {
		if im, ok := m.(IModel); ok {
			ormModel := im.ToORM()
			if err = scope.Create(ormModel); err != nil {
				return err
			}
			return copyFromORM(im, ormModel, false)
		}
	}
	return scope.Create(m)
}

func (x *T) BatchCreate(ctx context.Context, list any, omitFields ...string) (err error) {
	if err = x.validateList(list); err != nil {
		return err
	}
	if err = validateBatchElements(reflect.ValueOf(list).Elem()); err != nil {
		return err
	}
	list, err = x.encryptForWrite(list)
	if err != nil {
		return err
	}
	listV := reflect.ValueOf(list).Elem()
	scope := x.Scope(ctx)
	if len(omitFields) > 0 {
		scope = scope.Omit(omitFields...)
	}
	elemType := listV.Type().Elem()
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if x.forceTModel && elemType == x.modelType {
		valList, buildErr := x.buildORMList(listV)
		if buildErr != nil {
			return buildErr
		}
		if err = scope.CreateInBatches(valList, x.BatchCreateSize); err != nil {
			return err
		}
		return x.copyORMListToSource(list, valList, false)
	}
	return scope.CreateInBatches(list, x.BatchCreateSize)
}

func validateBatchElements(listV reflect.Value) error {
	for i := 0; i < listV.Len(); i++ {
		if isNilReflectValue(listV.Index(i)) {
			return fmt.Errorf("BatchCreate: element %d is nil", i)
		}
	}
	return nil
}

func (x *T) buildORMList(listV reflect.Value) (any, error) {
	ormList := reflect.MakeSlice(x.ormListType, listV.Len(), listV.Len())
	ormElemType := x.ormListType.Elem()
	for i := 0; i < listV.Len(); i++ {
		model, err := modelFromListElement(listV.Index(i))
		if err != nil {
			return nil, fmt.Errorf("BatchCreate: element %d: %w", i, err)
		}
		ormModel := model.ToORM()
		if ormModel == nil {
			return nil, fmt.Errorf("BatchCreate: element %d: ToORM returned nil", i)
		}
		ormValue := reflect.ValueOf(ormModel)
		if !ormValue.Type().AssignableTo(ormElemType) {
			return nil, fmt.Errorf("BatchCreate: element %d: ToORM returned %v, want %v", i, ormValue.Type(), ormElemType)
		}
		ormList.Index(i).Set(ormValue)
	}
	return ormList.Interface(), nil
}

func modelFromListElement(v reflect.Value) (IModel, error) {
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil, errors.New("model is nil")
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil, errors.New("model is nil")
	}
	if model, ok := reflect.TypeAssert[IModel](v); ok {
		return model, nil
	}
	if v.Kind() != reflect.Ptr && v.CanAddr() {
		if model, ok := reflect.TypeAssert[IModel](v.Addr()); ok {
			return model, nil
		}
	}
	return nil, fmt.Errorf("%v does not implement IModel", v.Type())
}

func (x *T) Save(ctx context.Context, m any, omitFields ...string) (err error) {
	m, err = x.encryptForWrite(m)
	if err != nil {
		return err
	}
	scope := x.Scope(ctx)
	if len(omitFields) > 0 {
		scope = scope.Omit(omitFields...)
	}
	if x.forceTModel {
		if im, ok := m.(IModel); ok {
			ormModel := im.ToORM()
			if err = scope.Save(ormModel); err != nil {
				return err
			}
			return copyFromORM(im, ormModel, false)
		}
	}
	return scope.Save(m)
}

func (x *T) Update(ctx context.Context, m any, query any, args ...any) (rows int64, err error) {
	m, err = x.encryptForWrite(m)
	if err != nil {
		return 0, err
	}
	scope := x.Scope(ctx)
	if query != nil {
		query, err = x.encryptQuery(query)
		if err != nil {
			return 0, err
		}
		scope = scope.Where(query, args...)
	}
	if im, ok := m.(IModel); ok && x.forceTModel {
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
	m, err = x.encryptForWrite(m)
	if err != nil {
		return 0, err
	}
	scope := x.Scope(ctx)
	scope = scope.Eq(x.PrimaryKey, pk)
	if len(selectFields) > 0 {
		scope = scope.Select(selectFields...)
	}
	if im, ok := m.(IModel); ok && x.forceTModel {
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
	query, err := x.encryptQuery(query)
	if err != nil {
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
	var err error
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
	query, err := x.encryptQuery(query)
	if err != nil {
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
		scope = scope.IgnoreNotFoundErr()
	}
	if query != nil {
		if scope, err = x.processOpts(scope, query, args...); err != nil {
			return err
		}
	}
	if !x.IgnoreNotFoundErr && x.NotFoundErrCode != 0 {
		scope = scope.SetNotFoundErr(x.NotFoundErrCode)
	}
	if im, ok := m.(IModel); ok && x.forceTModel {
		mV := im.ToORM()
		if err = scope.First(mV); err != nil {
			return err
		}
		err = copyFromORM(im, mV, false)
	} else {
		err = scope.First(m)
	}
	if err != nil {
		return err
	}
	return x.decryptAfterRead(m)
}

func (x *T) validateList(listPtr any) error {
	listV := reflect.ValueOf(listPtr)
	if !listV.IsValid() || listV.Kind() != reflect.Ptr || listV.IsNil() {
		return errors.New("list must be a non-nil pointer")
	}
	if listV.Elem().Kind() != reflect.Slice {
		return errors.New("list must be a pointer to a slice")
	}
	return nil
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
		enc, err := x.encryptQuery(o)
		if err != nil {
			return scope, err
		}
		qo := enc.(QueryOpts)
		scope = scope.Query(&qo)
	case *QueryOpts:
		enc, err := x.encryptQuery(o)
		if err != nil {
			return scope, err
		}
		scope = scope.Query(enc.(*QueryOpts))
	case map[string]any:
		if len(o) > 0 {
			enc, err := x.encryptQuery(o)
			if err != nil {
				return scope, err
			}
			return scope.Where(enc), nil
		}
	case string:
		if o != "" {
			// Phase 1: do not auto-encrypt raw SQL args.
			return scope.Where(o, args...), nil
		}
	default:
		optsV := reflect.ValueOf(opts)
		if optsV.Kind() == reflect.Map {
			if optsV.Len() > 0 {
				enc, err := x.encryptQuery(o)
				if err != nil {
					return scope, err
				}
				scope = scope.Where(enc)
			}
			return scope, nil
		}
		return scope, fmt.Errorf("unknown opts type: %v", reflect.TypeOf(opts))
	}
	return scope, nil
}

func (x *T) doBefore(ctx context.Context, opts any, listPtr any) (scope *Scope, valList any, needCopy bool, err error) {
	if err = x.validateList(listPtr); err != nil {
		log.Error(err)
		return
	}
	if x.needORMCopy {
		needCopy = true
		valList = x.newORMList()
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
		err = x.copyORMListToSource(listPtr, valList, true)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return x.decryptAfterRead(listPtr)
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

func (x *T) ListWithPagination(ctx context.Context, listPtr any, opts any, page, size uint32) (*pagination.PaginationResult, error) {
	scope, valList, needCopy, err := x.doBefore(ctx, opts, listPtr)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var newPagination *pagination.PaginationResult
	newPagination, err = scope.PaginationQuery(valList, page, size)
	if err != nil {
		return nil, err
	}
	return newPagination, x.doAfter(needCopy, listPtr, valList)
}

func (x *T) ListWithScroll(ctx context.Context, listPtr any, opts any, cursorField, cursor string, size uint32, isDesc bool) (*pagination.ScrollResult, error) {
	scope, valList, needCopy, err := x.doBefore(ctx, opts, listPtr)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var newScroll *pagination.ScrollResult
	newScroll, err = scope.ScrollQuery(valList, cursorField, cursor, size, isDesc)
	if err != nil {
		return nil, err
	}
	return newScroll, x.doAfter(needCopy, listPtr, valList)
}

func camelToSnake(s string) string {
	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			// 当前大写字母前面：
			// 1. 前一个是小写：直接加 _
			// 2. 前一个是大写，但后一个是小写：
			//    说明这是缩写的最后一个字母，例如 IDValue -> ID_Value
			if i > 0 &&
				(unicode.IsLower(runes[i-1]) ||
					(i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
				result.WriteByte('_')
			}
		}

		result.WriteRune(unicode.ToLower(r))
	}

	return result.String()
}

func copyFromORM(dst IModel, src ITModel, ignoreEmpty bool) error {
	if ignoreEmpty {
		return src.CopyToSourceIgnoreEmpty(dst)
	}
	return src.CopyToSource(dst)
}

func (x *T) copyORMListToSource(listPtr, valList any, ignoreEmpty bool) error {
	if x.copyList != nil {
		valV := reflect.ValueOf(valList)
		for valV.IsValid() && valV.Kind() == reflect.Ptr {
			if valV.IsNil() {
				return errors.New("copyORMListToSource: ORM list is nil")
			}
			valV = valV.Elem()
		}
		if !valV.IsValid() || valV.Kind() != reflect.Slice {
			return fmt.Errorf("copyORMListToSource: ORM list must be a slice, got %T", valList)
		}
		return x.copyList(listPtr, valV.Interface(), ignoreEmpty)
	}
	return copyORMListToSource(listPtr, valList, ignoreEmpty)
}

func copyORMListToSource(listPtr, valList any, ignoreEmpty bool) error {
	listPV := reflect.ValueOf(listPtr)
	if listPV.Kind() != reflect.Ptr {
		return fmt.Errorf("copyORMListToSource: listPtr must be pointer, got %T", listPtr)
	}

	valV := reflect.ValueOf(valList)
	for valV.Kind() == reflect.Ptr {
		valV = valV.Elem()
	}
	if valV.Kind() != reflect.Slice {
		return fmt.Errorf("copyORMListToSource: valList must be slice, got %v", valV.Kind())
	}

	holder := listPV.Elem()
	if holder.Kind() == reflect.Interface {
		if holder.IsNil() {
			return fmt.Errorf("copyORMListToSource: nil interface list")
		}
		inner := reflect.ValueOf(holder.Interface())
		if inner.Kind() != reflect.Slice {
			return fmt.Errorf("copyORMListToSource: interface holds %v, want slice", inner.Kind())
		}
		newDst, err := buildCopiedSlice(inner.Type(), valV, ignoreEmpty)
		if err != nil {
			return err
		}
		holder.Set(newDst)
		return nil
	}
	if holder.Kind() != reflect.Slice {
		return fmt.Errorf("copyORMListToSource: listPtr must be pointer to slice, got %T", listPtr)
	}
	newDst, err := buildCopiedSlice(holder.Type(), valV, ignoreEmpty)
	if err != nil {
		return err
	}
	holder.Set(newDst)
	return nil
}

func buildCopiedSlice(dstType reflect.Type, valV reflect.Value, ignoreEmpty bool) (reflect.Value, error) {
	n := valV.Len()
	newDst := reflect.MakeSlice(dstType, n, n)
	dstElemType := dstType.Elem()
	dstElemIsPtr := dstElemType.Kind() == reflect.Ptr
	iModelType := reflect.TypeFor[IModel]()
	if n > 0 && dstElemIsPtr {
		if !dstElemType.Implements(iModelType) {
			return reflect.Value{}, fmt.Errorf("buildCopiedSlice: dst element type %v is not IModel", dstElemType)
		}
	} else if n > 0 && !reflect.PointerTo(dstElemType).Implements(iModelType) {
		return reflect.Value{}, fmt.Errorf("buildCopiedSlice: dst element type %v is not IModel", dstElemType)
	}
	for i := 0; i < n; i++ {
		srcValue := valV.Index(i)
		if isNilReflectValue(srcValue) {
			return reflect.Value{}, fmt.Errorf("buildCopiedSlice: val element %d is nil", i)
		}
		src, ok := reflect.TypeAssert[ITModel](srcValue)
		if !ok {
			return reflect.Value{}, fmt.Errorf("buildCopiedSlice: val element %d is not ITModel", i)
		}
		dstElem := newDst.Index(i)
		if dstElemIsPtr {
			dstElem.Set(reflect.New(dstElemType.Elem()))
			dstIM, _ := reflect.TypeAssert[IModel](dstElem)
			if err := copyFromORM(dstIM, src, ignoreEmpty); err != nil {
				return reflect.Value{}, err
			}
			continue
		}
		dstIM, _ := reflect.TypeAssert[IModel](dstElem.Addr())
		if err := copyFromORM(dstIM, src, ignoreEmpty); err != nil {
			return reflect.Value{}, err
		}
	}
	return newDst, nil
}

func isNilReflectValue(v reflect.Value) bool {
	for v.IsValid() && v.Kind() == reflect.Interface {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return false
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return v.IsNil()
	default:
		return false
	}
}
