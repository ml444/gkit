package dbx

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ml444/gkit/dbx/crypto"
)

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

func TestBatchCreateWithInterfaceValue(t *testing.T) {
	conn := stubTxConn{d: stubDriver{}}
	repo := NewT[testRow](func() Conn { return conn })
	var wrapped any = []*testRow{{ID: 1, Name: "a"}}
	if err := repo.BatchCreate(context.Background(), wrapped); err != nil {
		t.Fatal(err)
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
	if _, ok := repo.getModel().(*copiedORM); !ok {
		t.Fatalf("forced model type = %T", repo.getModel())
	}
	idRepo := NewT[copiedORM](func() Conn { return StaticConn(stubDriver{}) }, WithTGenerateIDFunc(func() uint64 { return 9 }))
	valuesMap := map[string]any{}
	idRepo.generateIDs(valuesMap)
	if valuesMap[idRepo.PrimaryKey] != uint64(9) {
		t.Fatalf("map ID generation = %#v", valuesMap)
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

func dualFieldCrypto() T2Option {
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


func encryptRowCrypto(mutate ...crypto.MutatePolicy) T2Option {
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
	if err := repo.BatchCreate(ctx, list); err != nil {
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

	repo := NewT[storageEncryptRow](func() Conn { return conn }, dualFieldCrypto(), WithTGenerateIDFunc(func() uint64 { return 42 }))
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
		WithTGenerateIDFunc(func() uint64 { return 7 }),
	)
	row := &encryptRow{Name: "bob"}
	if err := okRepo.Create(ctx, row); err != nil || row.ID != 7 {
		t.Fatalf("id generator create: %#v %v", row, err)
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