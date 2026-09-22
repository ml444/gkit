package dbx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ml444/gkit/dbx/crypto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type badModel int

func TestInitAndRebuildNonStruct(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for non-struct model")
		}
	}()
	_ = NewT[badModel](func() Conn { return stubTxConn{d: stubDriver{}} })
}

func TestWithTCryptoTagDiscovery(t *testing.T) {
	type encryptNoJSON struct {
		ID     uint64 `gorm:"primaryKey"`
		Secret string `gorm:"encrypt:storage"`
		Skip   string `gorm:"encrypt:false"`
	}
	repo := NewT[encryptNoJSON](func() Conn { return stubTxConn{d: stubDriver{}} },
		WithTCrypto(crypto.Config{KeyRing: testKeyRing()}),
	)
	if repo.crypto == nil {
		t.Fatal("expected processor from tags")
	}
}

func TestWithTCryptoExplicitFields(t *testing.T) {
	repo := NewT[testRow](func() Conn { return stubTxConn{d: stubDriver{}} },
		WithTCrypto(crypto.Config{
			KeyRing: testKeyRing(),
			Fields:  []crypto.FieldSpec{{Column: "name", Struct: "Name", Mode: crypto.ModeSearchable}},
		}),
	)
	if repo.crypto == nil {
		t.Fatal("missing processor")
	}
}

func TestCryptoMapDecrypt(t *testing.T) {
	repo := NewT[encryptRow](func() Conn { return stubTxConn{d: stubDriver{}} }, nameSearchableCrypto())
	enc, err := repo.encryptForWrite(map[string]any{"name": "world"}, false)
	if err != nil {
		t.Fatal(err)
	}
	m := enc.(map[string]any)
	if err := repo.decryptAfterRead(m); err != nil || m["name"] != "world" {
		t.Fatalf("%#v %v", m, err)
	}
}

type failCopyORM struct {
	ID    uint64 `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	force bool
}

func (m *failCopyORM) ToSource() IModel                    { return &failCopySrc{ID: m.ID, Name: m.Name} }
func (m *failCopyORM) ForceTModel() bool                   { return m.force }
func (m *failCopyORM) CopyToSource(dst IModel) error       { return errors.New("copy fail") }
func (m *failCopyORM) CopyToSourceIgnoreEmpty(dst IModel) error {
	return errors.New("copy fail")
}

type failCopySrc struct {
	ID   uint64
	Name string
}

func (m *failCopySrc) ToORM() ITModel {
	return &failCopyORM{ID: m.ID, Name: m.Name, force: true}
}

func TestCopyORMListErrorBranches(t *testing.T) {
	var nilIface any
	if err := copyORMListToSource(&nilIface, []*failCopyORM{}, false); err == nil {
		t.Fatal("nil interface")
	}
	var notSlice any = 1
	if err := copyORMListToSource(&notSlice, []*failCopyORM{}, false); err == nil {
		t.Fatal("non-slice interface")
	}
	var holder any = []failCopySrc{{}}
	if err := copyORMListToSource(&holder, []*failCopyORM{{ID: 1, force: true}}, false); err == nil {
		t.Fatal("copy via interface should fail")
	}
	var num int
	if err := copyORMListToSource(&num, []*failCopyORM{}, false); err == nil {
		t.Fatal("non-slice holder")
	}
	var values []failCopySrc
	if err := copyORMListToSource(&values, []any{1}, false); err == nil {
		t.Fatal("non ITModel")
	}
	var ptrs []*int
	orms := []*failCopyORM{{ID: 1, force: true}}
	if err := copyORMListToSource(&ptrs, orms, false); err == nil {
		t.Fatal("dst not IModel pointer")
	}
	var failPtrs []*failCopySrc
	if err := copyORMListToSource(&failPtrs, orms, false); err == nil {
		t.Fatal("copy pointer dst fail")
	}
	if err := copyORMListToSource(&values, orms, false); err == nil {
		t.Fatal("copy value dst fail")
	}
	goodORMs := []*copiedORM{{ID: 1, Name: "a", force: true}}
	var badPtrs []*struct{ X int }
	if err := copyORMListToSource(&badPtrs, goodORMs, false); err == nil {
		t.Fatal("dst pointer not IModel")
	}
	type notModel struct{ ID uint64 }
	var notModels []notModel
	if err := copyORMListToSource(&notModels, goodORMs, false); err == nil {
		t.Fatal("dst value not IModel")
	}
}

func TestDoAfterCopyError(t *testing.T) {
	repo := NewT[failCopySrc](func() Conn { return stubTxConn{d: stubDriver{}} })
	list := []failCopySrc{{}}
	ormList := []*failCopyORM{{ID: 1, Name: "x", force: true}}
	if err := repo.doAfter(true, &list, ormList); err == nil {
		t.Fatal("expected copy error")
	}
}

type createFailDriver struct{ stubDriver }

func (d createFailDriver) Create(ctx context.Context, b *QueryBuilder, v any) (int64, error) {
	return 0, errors.New("create fail")
}
func (d createFailDriver) CreateInBatches(ctx context.Context, b *QueryBuilder, values any, batchSize int) (int64, error) {
	return 0, errors.New("batch fail")
}
func (d createFailDriver) Save(ctx context.Context, b *QueryBuilder, v any) (int64, error) {
	return 0, errors.New("save fail")
}
func (d createFailDriver) Transaction(ctx context.Context, fn func(Driver) error, opts ...TxOption) error {
	return fn(d)
}

func TestForceModelCreateSaveBatchErrors(t *testing.T) {
	conn := stubTxConn{d: createFailDriver{}}
	repo := NewT[copiedSource](func() Conn { return conn })
	ctx := context.Background()
	if err := repo.Create(ctx, &copiedSource{ID: 1, Name: "a"}, "name"); err == nil {
		t.Fatal("create")
	}
	if err := repo.Save(ctx, &copiedSource{ID: 1, Name: "a"}, "name"); err == nil {
		t.Fatal("save")
	}
	batch := []*copiedSource{{ID: 2, Name: "b"}}
	if err := repo.BatchCreate(ctx, &batch, "name"); err == nil {
		t.Fatal("batch")
	}
}

func TestBatchCreateEncryptAndInterface(t *testing.T) {
	conn := stubTxConn{d: stubDriver{}}
	repo := NewT[storageEncryptRow](func() Conn { return conn }, dualFieldCrypto())
	if err := repo.BatchCreate(context.Background(), []*storageEncryptRow{{Name: "a", Secret: "s"}}); err != nil {
		t.Fatal(err)
	}
	ok := NewT[testRow](func() Conn { return conn })
	var wrapped any = []*testRow{{ID: 1, Name: "a"}}
	if err := ok.BatchCreate(context.Background(), &wrapped); err != nil {
		t.Fatal(err)
	}
}

func TestSaveEncryptError(t *testing.T) {
	conn := stubTxConn{d: stubDriver{}}
	repo := NewT[storageEncryptRow](func() Conn { return conn }, dualFieldCrypto())
	if err := repo.Save(context.Background(), &storageEncryptRow{Name: "a", Secret: "s"}, "id"); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteByWhere(context.Background(), map[string]any{"secret": "s"}); err == nil {
		t.Fatal("expected storage query encrypt fail")
	}
}

func TestUpdateQueryEncryptAndGetOneProcessOpts(t *testing.T) {
	conn := stubTxConn{d: stubDriver{}}
	repo := NewT[storageEncryptRow](func() Conn { return conn }, dualFieldCrypto())
	ctx := context.Background()
	if _, err := repo.Update(ctx, map[string]any{"name": "z"}, map[string]any{"secret": "a"}); err == nil {
		t.Fatal("query encrypt on storage field")
	}
	if err := repo.GetOneByWhere(ctx, &storageEncryptRow{}, QueryOpts{Where: map[string]any{"secret": "a"}}); err == nil {
		t.Fatal("processOpts encrypt QueryOpts")
	}
	if err := repo.GetOneByWhere(ctx, &storageEncryptRow{}, &QueryOpts{Where: map[string]any{"secret": "a"}}); err == nil {
		t.Fatal("processOpts encrypt *QueryOpts")
	}
	if err := repo.GetOneByWhere(ctx, &storageEncryptRow{}, map[string]any{"secret": "a"}); err == nil {
		t.Fatal("processOpts encrypt map")
	}
	if err := repo.GetOneByWhere(ctx, &storageEncryptRow{}, map[string]string{"secret": "a"}); err == nil {
		t.Fatal("processOpts encrypt string map")
	}
}

func TestValidateListPointerElem(t *testing.T) {
	repo := NewT[testRow](func() Conn { return stubTxConn{d: stubDriver{}} })
	var list []*testRow
	if _, err := repo.validateListAndGetModel(&list); err != nil {
		t.Fatal(err)
	}
}

type countErrDriver struct{ stubDriver }

func (countErrDriver) Count(ctx context.Context, b *QueryBuilder) (int64, error) {
	return 0, errors.New("count boom")
}

type firstErrDriver struct{ stubDriver }

func (firstErrDriver) First(ctx context.Context, b *QueryBuilder, dest any) error {
	return errors.New("first boom")
}

func TestPaginationCountErrorAndExistError(t *testing.T) {
	s := NewScope(stubTxConn{d: countErrDriver{}}, &testRow{})
	var rows []testRow
	if _, err := s.PaginationQueryWithOpt(&rows, nil); err == nil {
		t.Fatal("count error")
	}
	s = NewScope(stubTxConn{d: firstErrDriver{}}, &testRow{})
	if _, err := s.Exist(); err == nil {
		t.Fatalf("exist count error")
	}
}

func TestWhereNonMapFallback(t *testing.T) {
	s := NewScope(stubTxConn{d: stubDriver{}}, &testRow{}).Where(42, "x")
	if len(s.Builder().Wheres) == 0 {
		t.Fatal("where missing")
	}
}

func TestResetTimeAndProtoUpsert(t *testing.T) {
	msg := mustDynamicUpsertMessage(t)
	stub := &protoUpsertStub{msg: msg}
	s := NewScope(stubTxConn{d: stubDriver{}}, &testRow{})
	s.resetTime = true
	s.ResetSysDateTimeField(stub)
	if err := s.Create(stub); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(stub); err != nil {
		t.Fatal(err)
	}
	if err := s.Update(map[string]any{"name": "x"}); err != nil {
		t.Fatal(err)
	}
}

type protoUpsertStub struct {
	msg protoreflect.Message
}

func (p *protoUpsertStub) ProtoReflect() protoreflect.Message { return p.msg }
func (p *protoUpsertStub) GetCreatedAt() int64                { return 1 }
func (p *protoUpsertStub) GetUpdatedAt() int64                { return 1 }
func (p *protoUpsertStub) GetDeletedAt() int64                { return 0 }

func mustDynamicUpsertMessage(t *testing.T) protoreflect.Message {
	t.Helper()
	file := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("dbx_upsert.proto"),
		Package: proto.String("dbx.test"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Upsert"),
			Field: []*descriptorpb.FieldDescriptorProto{
				{
					Name:     proto.String("created_at"),
					Number:   proto.Int32(1),
					Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum(),
					JsonName: proto.String("created_at"),
				},
				{
					Name:     proto.String("updated_at"),
					Number:   proto.Int32(2),
					Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum(),
					JsonName: proto.String("updated_at"),
				},
				{
					Name:     proto.String("deleted_at"),
					Number:   proto.Int32(3),
					Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum(),
					JsonName: proto.String("deleted_at"),
				},
			},
		}},
	}
	fd, err := protodesc.NewFile(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	md := fd.Messages().Get(0)
	msg := dynamicpb.NewMessage(md)
	fields := md.Fields()
	msg.Set(fields.ByName("created_at"), protoreflect.ValueOfUint32(9))
	msg.Set(fields.ByName("updated_at"), protoreflect.ValueOfUint32(9))
	msg.Set(fields.ByName("deleted_at"), protoreflect.ValueOfUint32(9))
	return msg
}

func TestScrollPointerLastAndPascalColumn(t *testing.T) {
	d := &seedFindDriver{rows: []map[string]any{{"id": int64(1), "name": "a", "age": int64(1), "deleted_at": uint32(0)}}}
	s := NewScope(stubTxConn{d: d}, &testRow{})
	var rows []*testRow
	if _, err := s.ScrollQuery(&rows, "0", 10); err != nil {
		t.Fatal(err)
	}
	type weird struct {
		FooBar int `json:""`
	}
	v := reflect.ValueOf(weird{FooBar: 7})
	// double underscore skips camelToSnake match then hits snakeToPascal exact name
	if got := fieldByColumn(v, "foo__bar"); !got.IsValid() || got.Int() != 7 {
		t.Fatalf("pascal path = %v", got)
	}
}

type seedFindDriver struct {
	stubDriver
	rows []map[string]any
}

func (d seedFindDriver) Find(ctx context.Context, b *QueryBuilder, dest any) error {
	return assignTestRows(dest, d.rows)
}

func assignTestRows(dest any, rows []map[string]any) error {
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Ptr {
		return fmt.Errorf("dest ptr")
	}
	sv := dv.Elem()
	elemType := sv.Type().Elem()
	isPtr := elemType.Kind() == reflect.Ptr
	if isPtr {
		elemType = elemType.Elem()
	}
	out := reflect.MakeSlice(sv.Type(), 0, len(rows))
	for _, row := range rows {
		val := reflect.New(elemType)
		rv := val.Elem()
		for i := 0; i < rv.NumField(); i++ {
			f := rv.Type().Field(i)
			col := f.Name
			if j := f.Tag.Get("json"); j != "" {
				col = j
			}
			if v, ok := row[col]; ok {
				rv.Field(i).Set(reflect.ValueOf(v).Convert(rv.Field(i).Type()))
			}
		}
		if isPtr {
			out = reflect.Append(out, val)
		} else {
			out = reflect.Append(out, val.Elem())
		}
	}
	sv.Set(out)
	return nil
}

func TestTxCreateMultiModelsError(t *testing.T) {
	err := TxCreateMultiModels(context.Background(), stubTxConn{d: createFailDriver{}}, &testRow{ID: 1})
	if err == nil {
		t.Fatal("expected create error")
	}
}

func TestSaveUpdateItemNonNotFoundAndNilExecute(t *testing.T) {
	ctx := context.Background()
	d := firstErrDriver{}
	item := &SaveItem{Model: &testRow{ID: 1}, Where: map[string]any{"id": int64(1)}}
	if err := item.Preload(ctx, d); err == nil {
		t.Fatal("preload non-not-found")
	}
	if err := (&SaveItem{}).Execute(ctx, d); err != nil {
		t.Fatal(err)
	}

	conn := stubTxConn{d: firstErrDriver{}}
	repo := NewT[storageEncryptRow](func() Conn { return conn }, dualFieldCrypto())
	su := &ScopeUpdateItem{Model: &storageEncryptRow{ID: 1, Name: "a"}, Where: map[string]any{"id": uint64(1)}, Updates: map[string]any{"secret": "z"}}
	if err := su.Execute(ctx, repo, d); err == nil {
		t.Fatal("updates encrypt on storage field")
	}
	su2 := &ScopeUpdateItem{Model: nil, Where: map[string]any{"id": uint64(1)}, Updates: map[string]any{}}
	if err := su2.Execute(ctx, repo, d); err == nil {
		t.Fatal("model update without updates")
	}

	ss := &ScopeSaveItem{Model: &storageEncryptRow{ID: 1}, Where: map[string]any{"id": uint64(1)}}
	if err := ss.Preload(ctx, repo, firstErrDriver{}); err == nil {
		t.Fatal("scopesave preload")
	}
	if err := (&ScopeSaveItem{}).Execute(ctx, repo, d); err != nil {
		t.Fatal(err)
	}
	ss2 := &ScopeSaveItem{Model: &storageEncryptRow{ID: 1, Name: "a", Secret: "s"}}
	if err := ss2.Execute(ctx, repo, d); err != nil {
		t.Fatal("scopesave execute")
	}
}
