package dbx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ml444/gkit/dbx/crypto"
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

func nameSearchableCrypto(mutate ...crypto.MutatePolicy) TOption {
	policy := crypto.MutateInPlace
	if len(mutate) > 0 {
		policy = mutate[0]
	}
	return WithTCrypto(crypto.Config{
		KeyRing: testKeyRing(),
		Fields:  []crypto.FieldSpec{{Column: "name", Struct: "Name", Mode: crypto.ModeSearchable}},
		Mutate:  policy,
	})
}
func TestCryptoMapDecrypt(t *testing.T) {
	repo := NewT[encryptRow](func() Conn { return stubTxConn{d: stubDriver{}} }, nameSearchableCrypto())
	enc, err := repo.encryptForWrite(map[string]any{"name": "world"})
	if err != nil {
		t.Fatal(err)
	}
	m := enc.(map[string]any)
	if err := repo.decryptAfterRead(m); err != nil || m["name"] != "world" {
		t.Fatalf("%#v %v", m, err)
	}
}


type copiedSource struct {
	ID   uint64
	Name string
}

func (m *copiedSource) ToORM() ITModel { return &copiedORM{ID: m.ID, Name: m.Name, force: true} }

type copiedORM struct {
	ID    uint64 `gorm:"primaryKey"`
	Name  string
	force bool
}

func (m *copiedORM) ToSource() IModel  { return &copiedSource{ID: m.ID, Name: m.Name} }
func (m *copiedORM) ForceTModel() bool { return m.force }
func (m *copiedORM) CopyToSource(dst IModel) error {
	*dst.(*copiedSource) = copiedSource{ID: m.ID, Name: m.Name}
	return nil
}
func (m *copiedORM) CopyToSourceIgnoreEmpty(dst IModel) error {
	d := dst.(*copiedSource)
	if m.ID != 0 {
		d.ID = m.ID
	}
	if m.Name != "" {
		d.Name = m.Name
	}
	return nil
}

type testRow struct {
	ID        int64  `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	Age       int64  `json:"age"`
	DeletedAt uint32 `json:"deleted_at"`
}

func (testRow) TableName() string { return "test_rows" }

type encryptRow struct {
	ID   uint64 `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"column:name;encrypt:searchable"`
}

func (encryptRow) TableName() string { return "encrypt_rows" }

type storageEncryptRow struct {
	ID     uint64 `json:"id" gorm:"primaryKey"`
	Name   string `json:"name" gorm:"column:name;encrypt:searchable"`
	Secret string `json:"secret" gorm:"column:secret;encrypt:storage"`
}

func (storageEncryptRow) TableName() string { return "storage_encrypt_rows" }

func TestBatchCreateRequiresSlicePointer(t *testing.T) {
	conn := stubTxConn{d: stubDriver{}}
	repo := NewT[testRow](func() Conn { return conn })
	var wrapped any = []*testRow{{ID: 1, Name: "a"}}
	if err := repo.BatchCreate(context.Background(), wrapped); err == nil {
		t.Fatal("non-pointer slice should fail")
	}
	rows := wrapped.([]*testRow)
	if err := repo.BatchCreate(context.Background(), &rows); err != nil {
		t.Fatal(err)
	}
}

func TestWritesDoNotGenerateIDs(t *testing.T) {
	repo := NewT[encryptRow](func() Conn { return stubTxConn{d: stubDriver{}} })
	ctx := context.Background()

	created := &encryptRow{Name: "create"}
	if err := repo.Create(ctx, created); err != nil || created.ID != 0 {
		t.Fatalf("Create changed zero ID: %#v, %v", created, err)
	}
	saved := &encryptRow{Name: "save"}
	if err := repo.Save(ctx, saved); err != nil || saved.ID != 0 {
		t.Fatalf("Save changed zero ID: %#v, %v", saved, err)
	}
	batch := []*encryptRow{{Name: "batch"}}
	if err := repo.BatchCreate(ctx, &batch); err != nil || batch[0].ID != 0 {
		t.Fatalf("BatchCreate changed zero ID: %#v, %v", batch, err)
	}
}

func TestORMCopyHelpersAndModelOptions(t *testing.T) {
	orms := []*copiedORM{{ID: 1, Name: "one", force: true}, {ID: 2, Name: "two", force: true}}
	var values []copiedSource
	if err := copyORMListToSource(&values, orms, false); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(values, []copiedSource{{ID: 1, Name: "one"}, {ID: 2, Name: "two"}}) {
		t.Fatalf("copied values = %#v", values)
	}
	var pointers []*copiedSource
	if err := copyORMListToSource(&pointers, orms, true); err != nil || pointers[1].Name != "two" {
		t.Fatalf("copied pointers = %#v, %v", pointers, err)
	}
	var holder any = []copiedSource{}
	if err := copyORMListToSource(&holder, orms, false); err != nil {
		t.Fatal(err)
	}
	if got := holder.([]copiedSource); len(got) != 2 {
		t.Fatalf("interface list = %#v", holder)
	}
	for _, tc := range []struct {
		list, values any
	}{
		{[]copiedSource{}, orms},
		{&values, 1},
	} {
		if err := copyORMListToSource(tc.list, tc.values, false); err == nil {
			t.Fatal("invalid copy inputs should fail")
		}
	}

	repo := NewT[copiedSource](func() Conn { return StaticConn(stubDriver{}) })
	if repo.modelType != reflect.TypeOf(copiedSource{}) || repo.ormModelType != reflect.TypeOf(copiedORM{}) || !repo.needORMCopy {
		t.Fatalf("model cache = %#v", repo)
	}
	if repo.ormListType != reflect.TypeOf([]*copiedORM{}) {
		t.Fatalf("orm list type = %v", repo.ormListType)
	}
	if _, ok := repo.getModel().(*copiedORM); !ok {
		t.Fatalf("forced model type = %T", repo.getModel())
	}
}

type batchCallDriver struct {
	stubDriver
	batchCalls int
}

func (d *batchCallDriver) CreateInBatches(_ context.Context, _ *QueryBuilder, _ any, _ int) (int64, error) {
	d.batchCalls++
	return 1, nil
}

func TestBatchCreateValidatesAndConvertsBeforeWrite(t *testing.T) {
	drv := &batchCallDriver{}
	repo := NewT[copiedSource](func() Conn { return stubTxConn{d: drv} })

	nilElement := []*copiedSource{nil}
	if err := repo.BatchCreate(context.Background(), &nilElement); err == nil {
		t.Fatal("nil source element should fail")
	}
	if drv.batchCalls != 0 {
		t.Fatalf("database called before validation: %d", drv.batchCalls)
	}

	values := []copiedSource{{ID: 1, Name: "one"}}
	if err := repo.BatchCreate(context.Background(), &values); err != nil {
		t.Fatal(err)
	}
	if drv.batchCalls != 1 || values[0].Name != "one" {
		t.Fatalf("value batch = %#v, calls = %d", values, drv.batchCalls)
	}

	array := [1]copiedSource{}
	if err := repo.BatchCreate(context.Background(), &array); err == nil {
		t.Fatal("array batch should fail")
	}
}

func TestRegisteredListCopyFastPath(t *testing.T) {
	called := 0
	repo := NewT[copiedSource](func() Conn { return StaticConn(stubDriver{}) }, WithTListCopyFunc(
		func(dst, src any, ignoreEmpty bool) error {
			called++
			return copyCopiedListTyped(dst, src, ignoreEmpty)
		},
	))
	var out []copiedSource
	orms := []*copiedORM{{ID: 1, Name: "one", force: true}}
	if err := repo.doAfter(true, &out, &orms); err != nil {
		t.Fatal(err)
	}
	if called != 1 || len(out) != 1 || out[0].Name != "one" {
		t.Fatalf("typed copy calls = %d, out = %#v", called, out)
	}
	clone := repo.Clone()
	if clone.copyList == nil {
		t.Fatal("clone lost registered list copy function")
	}
}

func copyCopiedListTyped(dst, src any, ignoreEmpty bool) error {
	out := dst.(*[]copiedSource)
	in := src.([]*copiedORM)
	result := make([]copiedSource, len(in))
	for i, item := range in {
		if item == nil {
			return fmt.Errorf("nil ORM element %d", i)
		}
		if err := copyFromORM(&result[i], item, ignoreEmpty); err != nil {
			return err
		}
	}
	*out = result
	return nil
}

func TestAssignCryptoResultRejectsUnsupportedTargets(t *testing.T) {
	if err := assignCryptoResult(1, 2); err == nil {
		t.Fatal("unsupported crypto assignment should fail")
	}
	if err := assignCryptoResult(nil, nil); err == nil {
		t.Fatal("nil crypto assignment should fail")
	}
}

func TestModelOptionProcessingAndCryptoBranches(t *testing.T) {
	repo := NewT[copiedORM](func() Conn { return StaticConn(stubDriver{}) },
		WithTCrypto(crypto.Config{
			KeyRing: testKeyRing(),
			Fields:  []crypto.FieldSpec{{Column: "name", Struct: "Name", Mode: crypto.ModeSearchable}},
		}),
	)
	scope := repo.Scope(context.Background())
	for _, opts := range []any{
		nil,
		Scope{},
		QueryOpts{Where: map[string]any{"name": "a"}},
		&QueryOpts{Where: map[string]any{"name": "a"}},
		map[string]any{"name": "a"},
		"name = ?",
		map[string]string{"name": "a"},
	} {
		if _, err := repo.processOpts(scope, opts, "a"); err != nil {
			t.Fatalf("process %T: %v", opts, err)
		}
	}
	if _, err := repo.processOpts(scope, 42); err == nil {
		t.Fatal("unknown opts should fail")
	}

	storageRepo := NewT[storageEncryptRow](func() Conn { return StaticConn(stubDriver{}) }, dualFieldCrypto())
	if _, err := storageRepo.encryptQuery(map[string]any{"secret": "x"}); !errors.Is(err, crypto.ErrNotSearchable) {
		t.Fatalf("storage query encrypt = %v", err)
	}

	badTagRepo := NewT[struct {
		ID   uint64 `gorm:"primaryKey"`
		Name string `gorm:"encrypt:not-a-bool"`
	}](func() Conn { return StaticConn(stubDriver{}) }, WithTCrypto(crypto.Config{KeyRing: testKeyRing()}))
	if badTagRepo.crypto != nil {
		t.Fatal("v1 encrypt tag should not enable encryption")
	}
	noTagRepo := NewT[struct{ ID uint64 }](func() Conn { return StaticConn(stubDriver{}) }, WithTCrypto(crypto.Config{KeyRing: testKeyRing()}))
	if noTagRepo.crypto != nil {
		t.Fatal("missing encrypt tags should not enable encryption")
	}
}

func testKeyRing() crypto.KeyRing {
	return crypto.KeyRing{
		ActiveID: "1",
		Keys: map[string][]byte{
			"1": []byte("0123456789abcdef0123456789abcdef"),
			"0": []byte("abcdef0123456789abcdef0123456789"),
		},
	}
}

func dualFieldCrypto() TOption {
	return WithTCrypto(crypto.Config{
		KeyRing: testKeyRing(),
		Fields: []crypto.FieldSpec{
			{Column: "name", Struct: "Name", Mode: crypto.ModeSearchable},
			{Column: "secret", Struct: "Secret", Mode: crypto.ModeStorage},
		},
		Mutate: crypto.MutateInPlace,
	})
}

func testConn() (*stubDriver, Conn) {
	d := &stubDriver{}
	return d, stubTxConn{d: d}
}

func encryptRowCrypto(mutate ...crypto.MutatePolicy) TOption {
	policy := crypto.MutateInPlace
	if len(mutate) > 0 {
		policy = mutate[0]
	}
	return WithTCrypto(crypto.Config{
		KeyRing: testKeyRing(),
		Fields:  []crypto.FieldSpec{{Column: "name", Struct: "Name", Mode: crypto.ModeSearchable}},
		Mutate:  policy,
	})
}

type v2EncryptRow struct {
	ID    uint64 `json:"id" gorm:"primaryKey"`
	Phone string `json:"phone" gorm:"column:phone;encrypt:searchable"`
	Card  string `json:"id_card" gorm:"column:id_card;encrypt:storage"`
	Name  string `json:"name" gorm:"column:name"`
}

func (v2EncryptRow) TableName() string { return "v2_encrypt_rows" }

type captureDriver struct {
	stubDriver
	lastCreate any
	firstFill  any
	lastWhere  any
}

type captureExistDriver struct {
	stubDriver
	builder *QueryBuilder
}

func (d *captureExistDriver) First(_ context.Context, b *QueryBuilder, _ any) error {
	d.builder = b
	return ErrRecordNotFound
}

func (d *captureDriver) Create(ctx context.Context, b *QueryBuilder, v any) (int64, error) {
	d.lastCreate = cloneAny(v)
	return 1, nil
}

func (d *captureDriver) First(ctx context.Context, b *QueryBuilder, dest any) error {
	if d.firstFill == nil {
		return ErrRecordNotFound
	}
	return copyAny(dest, d.firstFill)
}

func (d *captureDriver) Find(ctx context.Context, b *QueryBuilder, dest any) error {
	if d.firstFill == nil {
		return nil
	}
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Ptr {
		return errors.New("dest must be pointer")
	}
	dv = dv.Elem()
	if dv.Kind() != reflect.Slice {
		return errors.New("dest must be slice")
	}
	elemType := dv.Type().Elem()
	item := reflect.New(elemType)
	if elemType.Kind() != reflect.Ptr {
		if err := copyAny(item.Interface(), d.firstFill); err != nil {
			return err
		}
		dv.Set(reflect.Append(dv, item.Elem()))
		return nil
	}
	// []*T
	inner := reflect.New(elemType.Elem())
	if err := copyAny(inner.Interface(), d.firstFill); err != nil {
		return err
	}
	dv.Set(reflect.Append(dv, inner))
	return nil
}

func cloneAny(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		cp := reflect.New(rv.Elem().Type())
		cp.Elem().Set(rv.Elem())
		return cp.Interface()
	}
	return v
}

func copyAny(dst, src any) error {
	dv := reflect.ValueOf(dst)
	sv := reflect.ValueOf(src)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return errors.New("dst must be non-nil pointer")
	}
	for sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	}
	dv = dv.Elem()
	if dv.Type() != sv.Type() {
		// allow *T vs T
		if sv.Kind() == reflect.Struct && dv.Kind() == reflect.Struct && dv.Type() == sv.Type() {
			dv.Set(sv)
			return nil
		}
		return errors.New("type mismatch")
	}
	dv.Set(sv)
	return nil
}

func TestNewT2FastPath(t *testing.T) {
	repo := NewT[encryptRow](func() Conn { return stubTxConn{d: stubDriver{}} })
	if repo.crypto != nil {
		t.Fatal("expected nil processor")
	}
}

func TestT2CreateGetSearchableWhere(t *testing.T) {
	drv := &captureDriver{}
	repo := NewT[v2EncryptRow](func() Conn { return stubTxConn{d: drv} },
		WithTCrypto(crypto.Config{
			KeyRing: testKeyRing(),
		}),
	)
	if repo.crypto == nil {
		t.Fatal("expected processor from tags")
	}

	ctx := context.Background()
	in := &v2EncryptRow{Phone: "138", Card: "secret-card", Name: "bob"}
	if err := repo.Create(ctx, in); err != nil {
		t.Fatal(err)
	}
	if in.Phone != "138" || in.Card != "secret-card" {
		t.Fatalf("MutateCopy should keep caller plaintext: %#v", in)
	}
	created := drv.lastCreate.(*v2EncryptRow)
	if !strings.HasPrefix(created.Phone, "v1:k1:") {
		t.Fatalf("phone not enveloped: %q", created.Phone)
	}
	if !strings.HasPrefix(created.Card, "v1:k1:") {
		t.Fatalf("card not enveloped: %q", created.Card)
	}
	if created.Name != "bob" {
		t.Fatalf("name = %q", created.Name)
	}

	drv.firstFill = created
	out := &v2EncryptRow{}
	if err := repo.GetOne(ctx, out, uint64(0)); err != nil {
		t.Fatal(err)
	}
	if out.Phone != "138" || out.Card != "secret-card" || out.Name != "bob" {
		t.Fatalf("decrypt = %#v", out)
	}

	// searchable where
	drv2 := &captureDriver{firstFill: created}
	repo2 := NewT[v2EncryptRow](func() Conn { return stubTxConn{d: drv2} },
		WithTCrypto(crypto.Config{KeyRing: testKeyRing()}),
	)
	got := &v2EncryptRow{}
	if err := repo2.GetOneByWhere(ctx, got, map[string]any{"phone": "138"}); err != nil {
		t.Fatal(err)
	}
	if got.Phone != "138" {
		t.Fatalf("got %#v", got)
	}

	if _, err := repo2.Count(ctx, map[string]any{"id_card": "x"}); !errors.Is(err, crypto.ErrNotSearchable) {
		t.Fatalf("want ErrNotSearchable, got %v", err)
	}
}

func TestT2DisableDecrypt(t *testing.T) {
	drv := &captureDriver{}
	repo := NewT[v2EncryptRow](func() Conn { return stubTxConn{d: drv} },
		WithTCrypto(crypto.Config{KeyRing: testKeyRing()}),
		WithTDecrypt(false),
	)
	ctx := context.Background()
	_ = repo.Create(ctx, &v2EncryptRow{Phone: "138", Card: "c"})
	drv.firstFill = drv.lastCreate
	out := &v2EncryptRow{}
	if err := repo.GetOne(ctx, out, uint64(1)); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.Phone, "v1:k") {
		t.Fatalf("expected envelope left intact: %#v", out)
	}
}

func TestT2BatchCreateAndList(t *testing.T) {
	drv := &captureDriver{}
	repo := NewT[v2EncryptRow](func() Conn { return stubTxConn{d: drv} },
		WithTCrypto(crypto.Config{KeyRing: testKeyRing()}),
	)
	ctx := context.Background()
	list := []*v2EncryptRow{{Phone: "1", Card: "a"}, {Phone: "2", Card: "b"}}
	if err := repo.BatchCreate(ctx, &list); err != nil {
		t.Fatal(err)
	}
	if list[0].Phone != "1" {
		t.Fatal("batch MutateCopy must keep plaintext")
	}
	// reuse lastCreate as list fill (single row ok for decrypt path)
	drv.firstFill = &v2EncryptRow{Phone: "v1:k1:placeholder", Card: "v1:k1:placeholder", Name: "x"}
	// build a real encrypted row via Create
	drv2 := &captureDriver{}
	repo2 := NewT[v2EncryptRow](func() Conn { return stubTxConn{d: drv2} },
		WithTCrypto(crypto.Config{KeyRing: testKeyRing()}),
	)
	_ = repo2.Create(ctx, &v2EncryptRow{Phone: "99", Card: "cc", Name: "n"})
	drv.firstFill = drv2.lastCreate
	var out []v2EncryptRow
	if err := repo.ListAll(ctx, &out, nil); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Phone != "99" || out[0].Card != "cc" {
		t.Fatalf("list decrypt = %#v", out)
	}
}

func TestExistByWhereEncryptsQueryOnce(t *testing.T) {
	drv := &captureExistDriver{}
	repo := NewT[encryptRow](func() Conn { return stubTxConn{d: drv} }, encryptRowCrypto())

	if _, err := repo.ExistByWhere(context.Background(), map[string]any{"name": "alice"}); err != nil {
		t.Fatal(err)
	}
	if drv.builder == nil || len(drv.builder.Wheres) != 1 || len(drv.builder.Wheres[0].Args) != 1 {
		t.Fatalf("captured where = %#v", drv.builder)
	}
	ciphertext, ok := drv.builder.Wheres[0].Args[0].(string)
	if !ok {
		t.Fatalf("encrypted query value = %#v", drv.builder.Wheres[0].Args[0])
	}
	plain, err := repo.crypto.DecryptValue(map[string]any{"name": ciphertext})
	if err != nil {
		t.Fatal(err)
	}
	if got := plain.(map[string]any)["name"]; got != "alice" {
		t.Fatalf("query was encrypted more than once: decrypted once = %#v", got)
	}
}

func TestT2InPlaceMutate(t *testing.T) {
	drv := &captureDriver{}
	repo := NewT[v2EncryptRow](func() Conn { return stubTxConn{d: drv} },
		WithTCrypto(crypto.Config{
			KeyRing: testKeyRing(),
			Mutate:  crypto.MutateInPlace,
		}),
	)
	in := &v2EncryptRow{Phone: "138", Card: "c"}
	if err := repo.Create(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(in.Phone, "v1:k") {
		t.Fatalf("InPlace should mutate caller: %#v", in)
	}
}

func TestT2MultiKeyDecrypt(t *testing.T) {
	prim := crypto.NewPrimitive()
	ring := testKeyRing()
	k0, _ := ring.Get("0")
	encPhone, err := prim.Encrypt(crypto.ModeSearchable, k0, "138")
	if err != nil {
		t.Fatal(err)
	}
	encCard, err := prim.Encrypt(crypto.ModeStorage, k0, "card")
	if err != nil {
		t.Fatal(err)
	}
	drv := &captureDriver{firstFill: &v2EncryptRow{Phone: encPhone, Card: encCard, Name: "n"}}
	repo := NewT[v2EncryptRow](func() Conn { return stubTxConn{d: drv} },
		WithTCrypto(crypto.Config{KeyRing: ring}),
	)
	out := &v2EncryptRow{}
	if err := repo.GetOne(context.Background(), out, uint64(1)); err != nil {
		t.Fatal(err)
	}
	if out.Phone != "138" || out.Card != "card" {
		t.Fatalf("old key decrypt = %#v", out)
	}
}

func TestT2InvalidConfig(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid keyring")
		}
	}()
	NewT[v2EncryptRow](func() Conn { return stubTxConn{d: stubDriver{}} },
		WithTCrypto(crypto.Config{
			KeyRing: crypto.KeyRing{},
			Fields:  []crypto.FieldSpec{{Column: "phone", Mode: crypto.ModeSearchable}},
		}),
	)
}

func TestModelErrorAndOptionBranches(t *testing.T) {
	_, conn := testConn()
	ctx := context.Background()

	noPK := NewT[testRow](func() Conn { return conn }, WithTPrimaryKey(""))
	if _, err := noPK.UpdateByPk(ctx, map[string]any{"name": "x"}, int64(1)); err == nil {
		t.Fatal("UpdateByPk empty pk")
	}
	if err := noPK.DeleteByPk(ctx, int64(1)); err == nil {
		t.Fatal("DeleteByPk empty pk")
	}
	if err := noPK.GetOne(ctx, &testRow{}, int64(1)); err == nil {
		t.Fatal("GetOne empty pk")
	}

	repo := NewT[storageEncryptRow](func() Conn { return conn }, dualFieldCrypto())
	if err := repo.DeleteByWhere(ctx, map[string]any{"secret": "a"}); err == nil {
		t.Fatal("DeleteByWhere storage query error")
	}
	if _, err := repo.ExistByWhere(ctx, map[string]any{"secret": "a"}); err == nil {
		t.Fatal("ExistByWhere storage query error")
	}
	if _, err := repo.Count(ctx, map[string]any{"secret": "a"}); err == nil {
		t.Fatal("Count storage query error")
	}

	okRepo := NewT[encryptRow](func() Conn { return conn },
		encryptRowCrypto(crypto.MutateInPlace),
	)
	row := &encryptRow{Name: "bob"}
	if err := okRepo.Create(ctx, row); err != nil || row.ID != 0 {
		t.Fatalf("Create should not generate an ID: %#v %v", row, err)
	}

	plain := NewT[testRow](func() Conn { return conn })
	if _, err := plain.ExistByWhere(ctx); err != nil {
		t.Fatal(err)
	}
	var bad any = 1
	if err := plain.ListAll(ctx, bad, nil); err == nil {
		t.Fatal("ListAll non-pointer")
	}
	var notSlice *int
	if err := plain.ListAll(ctx, notSlice, nil); err == nil {
		t.Fatal("ListAll non-slice")
	}
	var list []testRow
	if err := plain.ListAll(ctx, &list, QueryOpts{Where: map[string]any{"id": int64(7)}}); err != nil {
		t.Fatal(err)
	}
	if err := plain.ListAll(ctx, &list, &QueryOpts{Where: map[string]any{"id": int64(7)}}); err != nil {
		t.Fatal(err)
	}
	scope := NewScope(conn, &testRow{})
	if err := plain.ListAll(ctx, &list, scope); err != nil {
		t.Fatal(err)
	}
	if err := plain.ListAll(ctx, &list, *scope); err != nil {
		t.Fatal(err)
	}
	if err := plain.ListAll(ctx, &list, 123); err == nil {
		t.Fatal("unknown opts")
	}
	if err := plain.ListAll(ctx, &list, ""); err != nil {
		t.Fatal(err)
	}
	exists, err := plain.ExistByWhere(ctx, "id = ?", int64(7))
	if err != nil {
		t.Fatal(err)
	}
	_ = exists

	ignored := plain.Clone(WithTIgnoreNotFoundErr())
	var missing testRow
	if err := ignored.GetOneByWhere(ctx, &missing, "id = ?", int64(99999)); err != nil {
		t.Fatalf("ignore not found: %v", err)
	}
}

func BenchmarkModelListValidation(b *testing.B) {
	repo := NewT[testRow](func() Conn { return StaticConn(stubDriver{}) })
	list := []testRow{}
	b.ReportAllocs()
	for b.Loop() {
		if err := repo.validateList(&list); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCopyORMListToValues1000(b *testing.B) {
	orms := make([]*copiedORM, 1000)
	for i := range orms {
		orms[i] = &copiedORM{ID: uint64(i + 1), Name: "name", force: true}
	}
	var out []copiedSource
	b.ReportAllocs()
	for b.Loop() {
		if err := copyORMListToSource(&out, &orms, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRegisteredListCopyToValues1000(b *testing.B) {
	repo := NewT[copiedSource](func() Conn { return StaticConn(stubDriver{}) }, WithTListCopyFunc(copyCopiedListTyped))
	orms := make([]*copiedORM, 1000)
	for i := range orms {
		orms[i] = &copiedORM{ID: uint64(i + 1), Name: "name", force: true}
	}
	var out []copiedSource
	b.ReportAllocs()
	for b.Loop() {
		if err := repo.copyORMListToSource(&out, &orms, false); err != nil {
			b.Fatal(err)
		}
	}
}


type fallbackSourceModel struct {
	SourceKey uint64 `gorm:"column:shared_id"`
}

func (m *fallbackSourceModel) ToORM() ITModel { return &fallbackORMModel{Key: m.SourceKey} }

type fallbackORMModel struct {
	Key uint64 `gorm:"primaryKey;column:shared_id"`
}

func (m *fallbackORMModel) ToSource() IModel { return &fallbackSourceModel{SourceKey: m.Key} }
func (*fallbackORMModel) ForceTModel() bool  { return true }
func (m *fallbackORMModel) CopyToSource(dst IModel) error {
	dst.(*fallbackSourceModel).SourceKey = m.Key
	return nil
}
func (m *fallbackORMModel) CopyToSourceIgnoreEmpty(dst IModel) error {
	if m.Key != 0 {
		dst.(*fallbackSourceModel).SourceKey = m.Key
	}
	return nil
}

type nilToORMSource struct{}

func (*nilToORMSource) ToORM() ITModel { return nil }

type otherORMSource struct{}

func (*otherORMSource) ToORM() ITModel { return &fallbackORMModel{} }

type conditionalORMSource struct {
	ReturnNil bool
}

func (m *conditionalORMSource) ToORM() ITModel {
	if m.ReturnNil {
		return nil
	}
	return &copiedORM{force: true}
}

type findFailureDriver struct {
	stubDriver
	err error
}

func (d findFailureDriver) Find(context.Context, *QueryBuilder, any) error { return d.err }
func (d findFailureDriver) WithContext(context.Context) Driver             { return d }

func TestModelMetadataCoverageBranches(t *testing.T) {
	conn := func() Conn { return StaticConn(stubDriver{}) }

	withoutPK := NewT[struct{ Name string }](conn)
	if withoutPK.PrimaryKey != "id" {
		t.Fatalf("primary key = %q, want default id", withoutPK.PrimaryKey)
	}

	fallback := NewT[fallbackSourceModel](conn)
	if fallback.PrimaryKey != "shared_id" {
		t.Fatalf("fallback metadata = %#v", fallback)
	}

	jsonName := NewT[struct {
		Key uint64 `gorm:"primaryKey" json:"record_key,omitempty"`
	}](conn)
	if jsonName.PrimaryKey != "record_key" {
		t.Fatalf("JSON primary-key name = %q", jsonName.PrimaryKey)
	}

	snakeName := NewT[struct {
		APIKey uint64 `gorm:"column;primary_key:autoIncrement" json:"-"`
	}](conn)
	if snakeName.PrimaryKey != "api_key" {
		t.Fatalf("field-name primary key = %q", snakeName.PrimaryKey)
	}

	composite := NewT[struct {
		First  uint64 `gorm:"primaryKey"`
		Second uint64 `gorm:"primaryKey"`
	}](conn)
	if composite.PrimaryKey != "" {
		t.Fatalf("composite primary key = %q, want empty", composite.PrimaryKey)
	}

	unexported := NewT[struct {
		id   uint64 `gorm:"primaryKey"`
		Name string
	}](conn)
	if unexported.PrimaryKey != "id" {
		t.Fatalf("unexported field changed default primary key: %q", unexported.PrimaryKey)
	}
}

func brokenEncryptionRepo(t *testing.T) *T {
	t.Helper()
	ring := testKeyRing()
	repo := NewT[encryptRow](func() Conn { return StaticConn(stubDriver{}) }, WithTCrypto(crypto.Config{
		KeyRing: ring,
		Fields:  []crypto.FieldSpec{{Column: "name", Struct: "Name", Mode: crypto.ModeSearchable}},
	}))
	delete(ring.Keys, ring.ActiveID)
	return repo
}

func TestModelWriteEncryptionErrorBranches(t *testing.T) {
	ctx := context.Background()
	repo := brokenEncryptionRepo(t)

	if err := repo.Create(ctx, &encryptRow{Name: "create"}); err == nil {
		t.Fatal("Create encryption should fail")
	}
	if err := repo.Save(ctx, &encryptRow{Name: "save"}); err == nil {
		t.Fatal("Save encryption should fail")
	}
	rows := []*encryptRow{{Name: "batch"}}
	if err := repo.BatchCreate(ctx, &rows); err == nil {
		t.Fatal("BatchCreate encryption should fail")
	}
	if _, err := repo.Update(ctx, &encryptRow{Name: "update"}, nil); err == nil {
		t.Fatal("Update encryption should fail")
	}
	if _, err := repo.UpdateByPk(ctx, &encryptRow{Name: "update"}, uint64(1)); err == nil {
		t.Fatal("UpdateByPk encryption should fail")
	}
	if _, err := repo.processOpts(repo.Scope(ctx), map[string]string{"name": "query"}); err == nil {
		t.Fatal("generic map query encryption should fail")
	}
}

func TestDecryptAndAssignmentCoverageBranches(t *testing.T) {
	repo := NewT[encryptRow](func() Conn { return StaticConn(stubDriver{}) }, encryptRowCrypto())
	if err := repo.decryptAfterRead(&encryptRow{Name: "not-an-envelope"}); err == nil {
		t.Fatal("invalid ciphertext should fail")
	}

	structValue := copiedSource{ID: 1}
	if err := assignCryptoResult(&structValue, copiedSource{ID: 2}); err != nil || structValue.ID != 2 {
		t.Fatalf("value assignment = %#v, %v", structValue, err)
	}
	sliceValue := []int{1}
	if err := assignCryptoResult(&sliceValue, []int{2}); err != nil || sliceValue[0] != 2 {
		t.Fatalf("slice assignment = %#v, %v", sliceValue, err)
	}
	interfaceValue := any(nil)
	value := 3
	if err := assignCryptoResult(&interfaceValue, &value); err != nil || interfaceValue != &value {
		t.Fatalf("interface assignment = %#v, %v", interfaceValue, err)
	}
	type namedStruct struct{ Value int }
	unnamed := struct{ Value int }{}
	if err := assignCryptoResult(&unnamed, &namedStruct{Value: 4}); err != nil || unnamed.Value != 4 {
		t.Fatalf("pointer element assignment = %#v, %v", unnamed, err)
	}
}

func TestEncryptQueryDefaultNonMap(t *testing.T) {
	repo := NewT[encryptRow](func() Conn { return StaticConn(stubDriver{}) }, encryptRowCrypto())
	query := struct{ Name string }{Name: "plain"}
	out, err := repo.encryptQuery(query)
	if err != nil || out != query {
		t.Fatalf("query = %#v, %v", out, err)
	}
}

func TestBuildORMListAndModelElementCoverageBranches(t *testing.T) {
	repo := &T{ormListType: reflect.TypeOf([]*copiedORM{})}

	bad := []any{1}
	if _, err := repo.buildORMList(reflect.ValueOf(bad)); err == nil {
		t.Fatal("non-model list element should fail")
	}
	nilORM := []nilToORMSource{{}}
	if _, err := repo.buildORMList(reflect.ValueOf(nilORM)); err == nil {
		t.Fatal("nil ORM conversion should fail")
	}
	wrongORM := []otherORMSource{{}}
	if _, err := repo.buildORMList(reflect.ValueOf(wrongORM)); err == nil {
		t.Fatal("wrong ORM conversion type should fail")
	}

	var nilInterface any
	if _, err := modelFromListElement(reflect.ValueOf(&nilInterface).Elem()); err == nil {
		t.Fatal("nil interface model should fail")
	}
	if _, err := modelFromListElement(reflect.ValueOf((*copiedSource)(nil))); err == nil {
		t.Fatal("nil pointer model should fail")
	}
	if _, err := modelFromListElement(reflect.ValueOf(1)); err == nil {
		t.Fatal("non-addressable non-model should fail")
	}
	ints := []int{1}
	if _, err := modelFromListElement(reflect.ValueOf(ints).Index(0)); err == nil {
		t.Fatal("addressable non-model should fail")
	}
}

func TestBatchCreateORMBuildError(t *testing.T) {
	repo := NewT[conditionalORMSource](func() Conn { return StaticConn(stubDriver{}) })
	rows := []*conditionalORMSource{{ReturnNil: true}}
	if err := repo.BatchCreate(context.Background(), &rows); err == nil {
		t.Fatal("BatchCreate should return ORM build error")
	}
}

func TestListWithScrollCoverageBranches(t *testing.T) {
	ctx := context.Background()
	repo := NewT[testRow](func() Conn { return StaticConn(stubDriver{}) })
	if _, err := repo.ListWithScroll(ctx, []testRow{}, nil, "id", "", 10, false); err == nil {
		t.Fatal("non-pointer list should fail")
	}

	findErr := errors.New("find failed")
	failing := NewT[testRow](func() Conn { return StaticConn(findFailureDriver{err: findErr}) })
	var failed []testRow
	if _, err := failing.ListWithScroll(ctx, &failed, nil, "id", "", 10, false); !errors.Is(err, findErr) {
		t.Fatalf("scroll error = %v", err)
	}

	var rows []testRow
	result, err := repo.ListWithScroll(ctx, &rows, nil, "id", "", 10, false)
	if err != nil || result == nil {
		t.Fatalf("scroll result = %#v, %v", result, err)
	}
}

func TestRegisteredListCopyValidationBranches(t *testing.T) {
	repo := NewT[copiedSource](func() Conn { return StaticConn(stubDriver{}) }, WithTListCopyFunc(copyCopiedListTyped))
	var out []copiedSource
	var nilORMList *[]*copiedORM
	if err := repo.copyORMListToSource(&out, nilORMList, false); err == nil {
		t.Fatal("nil ORM list should fail")
	}
	if err := repo.copyORMListToSource(&out, 1, false); err == nil {
		t.Fatal("non-slice ORM list should fail")
	}
}

func TestIsNilReflectValueCoverageBranches(t *testing.T) {
	var nilInterface any
	if !isNilReflectValue(reflect.ValueOf(&nilInterface).Elem()) {
		t.Fatal("nil interface should be nil")
	}
	if isNilReflectValue(reflect.Value{}) {
		t.Fatal("invalid value should not be nil")
	}
	if isNilReflectValue(reflect.ValueOf(1)) {
		t.Fatal("integer should not be nil")
	}
}

type failCopyORM struct {
	ID    uint64 `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	force bool
}

func (m *failCopyORM) ToSource() IModel              { return &failCopySrc{ID: m.ID, Name: m.Name} }
func (m *failCopyORM) ForceTModel() bool             { return m.force }
func (m *failCopyORM) CopyToSource(dst IModel) error { return errors.New("copy fail") }
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
	if err := copyORMListToSource(&values, []*failCopyORM{nil}, false); err == nil {
		t.Fatal("nil ITModel")
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
	storageRows := []*storageEncryptRow{{Name: "a", Secret: "s"}}
	if err := repo.BatchCreate(context.Background(), &storageRows); err != nil {
		t.Fatal(err)
	}
	ok := NewT[testRow](func() Conn { return conn })
	rows := []*testRow{{ID: 1, Name: "a"}}
	if err := ok.BatchCreate(context.Background(), &rows); err != nil {
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
	if err := repo.validateList(&list); err != nil {
		t.Fatal(err)
	}
	var nilListPtr *[]*testRow
	if err := repo.validateList(nilListPtr); err == nil {
		t.Fatal("typed nil list pointer should fail")
	}
	if err := repo.validateList(nil); err == nil {
		t.Fatal("nil list should fail")
	}
}
