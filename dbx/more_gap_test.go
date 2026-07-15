package dbx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ml444/gkit/dbx"
)

type softUpdatedRow struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	UpdatedAt int64  `json:"updated_at"`
	DeletedAt int64  `json:"deleted_at"`
}

func (softUpdatedRow) TableName() string     { return "soft_updated" }
func (s softUpdatedRow) GetDeletedAt() int64 { return s.DeletedAt }
func (s softUpdatedRow) GetUpdatedAt() int64 { return s.UpdatedAt }

type plainDriver struct{}

func (plainDriver) Find(ctx context.Context, b *dbx.QueryBuilder, dest any) error { return nil }
func (plainDriver) First(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	return dbx.ErrRecordNotFound
}
func (plainDriver) Count(ctx context.Context, b *dbx.QueryBuilder) (int64, error) { return 0, nil }
func (plainDriver) Create(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	return 1, nil
}
func (plainDriver) CreateInBatches(ctx context.Context, b *dbx.QueryBuilder, values any, batchSize int) (int64, error) {
	return 1, nil
}
func (plainDriver) Save(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	return 1, nil
}
func (plainDriver) Update(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	return 0, errors.New("upd")
}
func (plainDriver) UpdateColumn(ctx context.Context, b *dbx.QueryBuilder, field string, value any) (int64, error) {
	return 0, errors.New("upc")
}
func (plainDriver) Delete(ctx context.Context, b *dbx.QueryBuilder) (int64, error) { return 0, nil }
func (plainDriver) Scan(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	return dbx.ErrRecordNotFound
}
func (plainDriver) Transaction(ctx context.Context, fn func(dbx.Driver) error, opts ...dbx.TxOption) error {
	return fn(plainDriver{})
}

type plainConn struct{ d dbx.Driver }

func (c plainConn) Driver(ctx context.Context) dbx.Driver { return c.d }

func TestProtoUpdatedAndPlainDriverTx(t *testing.T) {
	_, conn := testConn()
	s := dbx.NewScope(conn, softUpdatedRow{})
	if err := s.Update(map[string]any{"name": "n"}, "id = ?", int64(1)); err != nil {
		t.Fatal(err)
	}

	pc := plainConn{d: plainDriver{}}
	if err := dbx.TxGo(context.Background(), pc, func(d dbx.Driver) error {
		return dbx.NewScope(dbx.StaticConn(d), &testRow{}).Create(&testRow{ID: 1})
	}); err != nil {
		t.Fatal(err)
	}
	s2 := dbx.NewScope(pc, &testRow{})
	_ = s2.Update(map[string]any{"name": "x"})
	_ = s2.UpdateColumn("name", "y")
	_ = s2.UpdateColumnWithIncr("age", 1)

	repo := dbx.NewT[srcModel](func() dbx.Conn { return conn })
	if err := repo.BatchCreate(context.Background(), &srcModel{ID: 8, Name: "solo"}); err != nil {
		t.Fatal(err)
	}
	var list []srcModel
	_ = repo.ListAll(context.Background(), &list, map[string]string{"name": "alice"})

	fc := failCipher{err: errors.New("d")}
	er := dbx.NewT[encryptRow](func() dbx.Conn { return conn }, dbx.SetSpecifyFieldCipherMap(map[string]dbx.FieldCipher{
		"name": {StructField: "Name", Cipher: fc},
	}))
	_ = er.CheckAndCrypto([]*encryptRow{{Name: "a"}}, dbx.CipherKindEncrypt, false)
	_ = er.CheckAndCrypto(map[string]any{"name": "a"}, dbx.CipherKindEncrypt, false)

	d, _ := testConn()
	su := &dbx.ScopeUpdateItem{Model: &encryptRow{ID: 1}, Where: map[string]any{"name": "a"}, Updates: nil}
	_ = su.Preload(er, d)
	_ = su.Execute(er, d)
	ss := &dbx.ScopeSaveItem{Model: &encryptRow{ID: 1, Name: "a"}, Where: map[string]any{"name": "a"}}
	_ = ss.Preload(er, d)
	_ = ss.Execute(er, d)
	si := &dbx.SaveItem{Model: &testRow{ID: 1}, Where: map[string]any{"id": int64(1)}}
	_ = si.Preload(d)
	upd := &dbx.UpdateItem{Model: &testRow{}, Where: map[string]any{"id": int64(1)}, Updates: nil}
	_ = upd.Execute(d)

	er2 := dbx.NewT[encryptRow](func() dbx.Conn { return conn }, dbx.SetSpecifyFieldCipherMap(map[string]dbx.FieldCipher{
		"name": {StructField: "Name", Cipher: failCipher{err: errors.New("q")}},
	}))
	_, _ = er2.ExistByWhere(context.Background(), "name = ?", "alice")
	_ = er2.ListAll(context.Background(), &[]encryptRow{}, "name = ?")

	d3, c3 := testConn()
	fr := dbx.NewT[srcModel](func() dbx.Conn { return c3 }, dbx.SetNotFoundErrCode(3))
	_ = fr.GetOne(context.Background(), &srcModel{}, uint64(404))
	_ = d3

	// force GetOne success already covered; trigger Find error on ListAll
	d3.FindErr = errors.New("listfail")
	_ = fr.ListAll(context.Background(), &list, nil)
	_, _ = fr.ListWithPagination(context.Background(), &list, nil, 1, 1)
}
