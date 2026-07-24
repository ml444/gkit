package dbx

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ml444/gkit/dbx/crypto"
)

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

func testKeyRing() crypto.KeyRing {
	return crypto.KeyRing{
		ActiveID: "1",
		Keys: map[string][]byte{
			"1": []byte("0123456789abcdef0123456789abcdef"),
			"0": []byte("abcdef0123456789abcdef0123456789"),
		},
	}
}

func TestNewT2FastPath(t *testing.T) {
	repo, err := NewT2[encryptRow](func() Conn { return stubTxConn{d: stubDriver{}} })
	if err != nil {
		t.Fatal(err)
	}
	if repo.crypto != nil {
		t.Fatal("expected nil processor")
	}
}

func TestT2CreateGetSearchableWhere(t *testing.T) {
	drv := &captureDriver{}
	repo, err := NewT2[v2EncryptRow](func() Conn { return stubTxConn{d: drv} },
		WithCrypto(crypto.Config{
			KeyRing: testKeyRing(),
			// tag discovery
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
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
	repo2, err := NewT2[v2EncryptRow](func() Conn { return stubTxConn{d: drv2} },
		WithCrypto(crypto.Config{KeyRing: testKeyRing()}),
	)
	if err != nil {
		t.Fatal(err)
	}
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
	repo, err := NewT2[v2EncryptRow](func() Conn { return stubTxConn{d: drv} },
		WithCrypto(crypto.Config{KeyRing: testKeyRing()}),
		WithDecrypt(false),
	)
	if err != nil {
		t.Fatal(err)
	}
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
	repo, err := NewT2[v2EncryptRow](func() Conn { return stubTxConn{d: drv} },
		WithCrypto(crypto.Config{KeyRing: testKeyRing()}),
	)
	if err != nil {
		t.Fatal(err)
	}
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
	repo2, _ := NewT2[v2EncryptRow](func() Conn { return stubTxConn{d: drv2} },
		WithCrypto(crypto.Config{KeyRing: testKeyRing()}),
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
	repo, err := NewT2[v2EncryptRow](func() Conn { return stubTxConn{d: drv} },
		WithCrypto(crypto.Config{
			KeyRing: testKeyRing(),
			Mutate:  crypto.MutateInPlace,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
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
	repo, err := NewT2[v2EncryptRow](func() Conn { return stubTxConn{d: drv} },
		WithCrypto(crypto.Config{KeyRing: ring}),
	)
	if err != nil {
		t.Fatal(err)
	}
	out := &v2EncryptRow{}
	if err := repo.GetOne(context.Background(), out, uint64(1)); err != nil {
		t.Fatal(err)
	}
	if out.Phone != "138" || out.Card != "card" {
		t.Fatalf("old key decrypt = %#v", out)
	}
}

func TestT2InvalidConfig(t *testing.T) {
	_, err := NewT2[v2EncryptRow](func() Conn { return stubTxConn{d: stubDriver{}} },
		WithCrypto(crypto.Config{
			KeyRing: crypto.KeyRing{},
			Fields:  []crypto.FieldSpec{{Column: "phone", Mode: crypto.ModeSearchable}},
		}),
	)
	if err == nil {
		t.Fatal("expected invalid keyring error")
	}
}
