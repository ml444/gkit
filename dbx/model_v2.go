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
	primaryKeyIndex   int

	crypto         *crypto.Processor
	disableDecrypt bool

	modelType    reflect.Type
	ormModelType reflect.Type
	needORMCopy  bool

	IdGenerator GenerateIDFunc
}

// NewT constructs a T repository. Invalid crypto options return an error.
func NewT[M any](fn func() Conn, opts ...T2Option) *T {
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

func (x *T) initModelMeta()  {
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
	if len(pkList) == 1 {
		x.PrimaryKey = pkList[0]
	}
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
	x.needORMCopy = x.forceTModel && modelT != ormT
}

func (x *T) Clone(opts ...T2Option) *T {
	t := &T{
		getConn:           x.getConn,
		model:             x.model,
		ormModel:          x.ormModel,
		forceTModel:       x.forceTModel,
		IgnoreNotFoundErr: x.IgnoreNotFoundErr,
		NotFoundErrCode:   x.NotFoundErrCode,
		BatchCreateSize:   x.BatchCreateSize,
		PrimaryKey:        x.PrimaryKey,
		primaryKeyIndex:   x.primaryKeyIndex,
		crypto:            x.crypto,
		disableDecrypt:    x.disableDecrypt,
		modelType:         x.modelType,
		ormModelType:      x.ormModelType,
		needORMCopy:       x.needORMCopy,
		IdGenerator:       x.IdGenerator,
	}
	for _, opt := range opts {
		opt(t)
	}
	t.cacheTypes()
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
	return NewScope(x.getConn(), x.getModel()).WithContext(ctx)
}

func (x *T) generateIDs(m any) {
	if x.IdGenerator == nil || x.primaryKeyIndex < 0 || m == nil {
		return
	}
	mV := reflect.ValueOf(m)
	for mV.Kind() == reflect.Ptr {
		if mV.IsNil() {
			return
		}
		mV = mV.Elem()
	}
	switch mV.Kind() {
	case reflect.Struct:
		pkValue := mV.Field(x.primaryKeyIndex)
		if pkValue.IsValid() && pkValue.IsZero() && pkValue.CanSet() && pkValue.Kind() == reflect.Uint64 {
			pkValue.SetUint(x.IdGenerator())
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < mV.Len(); i++ {
			x.generateIDs(mV.Index(i).Interface())
		}
	case reflect.Map:
		if x.PrimaryKey == "" {
			return
		}
		mV.SetMapIndex(reflect.ValueOf(x.PrimaryKey), reflect.ValueOf(x.IdGenerator()))
	}
}

func (x *T) encryptForWrite(m any, isCreate bool) (any, error) {
	if isCreate {
		x.generateIDs(m)
	}
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
	// MutateCopy returns a new value; copy fields back into caller's object when possible.
	if out == nil || reflect.DeepEqual(out, m) {
		return nil
	}
	return assignCryptoResult(m, out)
}

func assignCryptoResult(dst, src any) error {
	dv := reflect.ValueOf(dst)
	sv := reflect.ValueOf(src)
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
	return nil
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
	m, err = x.encryptForWrite(m, true)
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
	list, err = x.encryptForWrite(list, true)
	if err != nil {
		return err
	}
	listV := reflect.Indirect(reflect.ValueOf(list))
	if listV.Kind() == reflect.Interface {
		listV = reflect.ValueOf(listV.Interface())
	}
	scope := x.Scope(ctx)
	if len(omitFields) > 0 {
		scope = scope.Omit(omitFields...)
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
			return copyORMListToSource(list, valList, false)
		}
		return scope.CreateInBatches(list, x.BatchCreateSize)
	default:
		return scope.Create(list)
	}
}

func (x *T) Save(ctx context.Context, m any, omitFields ...string) (err error) {
	m, err = x.encryptForWrite(m, true)
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
	m, err = x.encryptForWrite(m, false)
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
	m, err = x.encryptForWrite(m, false)
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
	query, err := x.encryptQuery(query)
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
	return reflect.New(elemType).Interface(), nil
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
		err = copyORMListToSource(listPtr, valList, true)
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

func copyFromORM(dst IModel, src ITModel, ignoreEmpty bool) error {
	if ignoreEmpty {
		return src.CopyToSourceIgnoreEmpty(dst)
	}
	return src.CopyToSource(dst)
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
	for i := 0; i < n; i++ {
		src, ok := valV.Index(i).Interface().(ITModel)
		if !ok {
			return reflect.Value{}, fmt.Errorf("copyORMListToSource: val element %d is not ITModel", i)
		}
		dstElem := newDst.Index(i)
		if dstElem.Kind() == reflect.Ptr {
			if dstElem.IsNil() {
				dstElem.Set(reflect.New(dstElem.Type().Elem()))
			}
			dstIM, ok := dstElem.Interface().(IModel)
			if !ok {
				return reflect.Value{}, fmt.Errorf("copyORMListToSource: dst element %d is not IModel", i)
			}
			if err := copyFromORM(dstIM, src, ignoreEmpty); err != nil {
				return reflect.Value{}, err
			}
			continue
		}
		dstIM, ok := dstElem.Addr().Interface().(IModel)
		if !ok {
			return reflect.Value{}, fmt.Errorf("copyORMListToSource: dst element %d is not IModel", i)
		}
		if err := copyFromORM(dstIM, src, ignoreEmpty); err != nil {
			return reflect.Value{}, err
		}
	}
	return newDst, nil
}
