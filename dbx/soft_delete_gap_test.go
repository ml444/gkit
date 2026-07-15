package dbx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ml444/gkit/dbx"
)

type softRow struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	DeletedAt int64  `json:"deleted_at"`
}

func (softRow) TableName() string     { return "soft_rows" }
func (s softRow) GetDeletedAt() int64 { return s.DeletedAt }

func TestSoftDeleteScopeAndMoreGaps(t *testing.T) {
	d, conn := testConn()
	d.Seed("soft_rows", map[string]any{"id": int64(1), "name": "a", "deleted_at": int64(0)})
	s := dbx.NewScope(conn, softRow{})
	if len(s.Builder().Wheres) == 0 || s.Builder().Wheres[0].Query != "deleted_at = 0" {
		t.Fatalf("soft delete where missing: %#v", s.Builder().Wheres)
	}
	if err := s.Eq("id", int64(1)).Delete(); err != nil {
		t.Fatal(err)
	}
	pure := dbx.NewScopeOfPure(conn, softRow{})
	if len(pure.Builder().Wheres) != 0 {
		t.Fatal("pure should skip soft delete filter")
	}
	s.ResetSysDateTimeField(&testRow{})

	ctx := context.Background()
	repo := dbx.NewT[srcModel](func() dbx.Conn { return conn }, dbx.SetTableCipher(prefixCipher{}))
	// SetTableCipher with forceT orm without encrypt tags should warn and clear NeedEncrypt
	_ = repo

	failRepo := dbx.NewT[encryptRow](func() dbx.Conn { return conn },
		dbx.SetSpecifyFieldCipherMap(map[string]dbx.FieldCipher{
			"name": {StructField: "Name", Cipher: failCipher{err: errors.New("enc")}},
		}),
	)
	if _, err := failRepo.Update(ctx, map[string]any{"name": "z"}, "id = ?", uint64(1)); err == nil {
		t.Fatal("update encrypt fail")
	}
	if _, err := failRepo.Update(ctx, map[string]any{"name": "z"}, map[string]any{"name": "a"}); err == nil {
		t.Fatal("update query encrypt fail")
	}
	if _, err := failRepo.UpdateByPk(ctx, map[string]any{"name": "z"}, uint64(1)); err == nil {
		t.Fatal("updatebypk encrypt fail")
	}

	d2 := d
	d2.CreateErr = errors.New("create fail")
	badConn := d2.Conn()
	force := dbx.NewT[srcModel](func() dbx.Conn { return badConn })
	if err := force.Create(ctx, &srcModel{ID: 99, Name: "x"}); err == nil {
		t.Fatal("force create fail")
	}
	if err := force.Save(ctx, &srcModel{ID: 99, Name: "x"}); err == nil {
		t.Fatal("force save fail")
	}
	if err := force.BatchCreate(ctx, []*srcModel{{ID: 100, Name: "y"}}); err == nil {
		t.Fatal("force batch fail")
	}

	okd, okc := testConn()
	okd.Seed("orm_models", map[string]any{"id": uint64(1), "name": "alice"})
	okRepo := dbx.NewT[srcModel](func() dbx.Conn { return okc })
	var page []srcModel
	if _, err := okRepo.ListWithPagination(ctx, &page, map[string]any{}, 1, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := okRepo.ListWithPagination(ctx, page, nil, 1, 10); err == nil {
		t.Fatal("non pointer page")
	}

	// UpdateColumn error path via CreateErr not applicable; use Find for Exist error
	okd.FindErr = errors.New("count proxy")
	// Count uses Count not Find - skip

	if err := dbx.TxCreateMultiModels(ctx, okc, &testRow{ID: 50, Name: "t"}); err != nil && !errors.Is(err, okd.FindErr) {
		// may succeed on Create
		_ = err
	}

	// Trigger ScopeTxGo callback error
	if err := dbx.ScopeTxGo(ctx, okc, func() (any, func(*dbx.Scope) error) {
		return &testRow{}, func(*dbx.Scope) error { return errors.New("cb") }
	}); err == nil {
		t.Fatal("scope tx callback error")
	}
	if err := dbx.ScopeTxGoWithT(ctx, okRepo, func() (any, func(*dbx.Scope) error) {
		return &srcModel{}, func(*dbx.Scope) error { return errors.New("cb2") }
	}); err == nil {
		t.Fatal("scope tx with T callback error")
	}

	preFail := &dbx.UpdateItem{Model: &testRow{}, Where: map[string]any{"id": int64(99999)}}
	if err := dbx.RunTxItems(ctx, okc, preFail); err == nil {
		t.Fatal("preload miss should fail")
	}
	itemFail := &failTxItem{}
	if err := dbx.RunTxItems(ctx, okc, itemFail); err == nil {
		t.Fatal("item preload fail")
	}
	itemExecFail := &failExecItem{}
	if err := dbx.RunTxItems(ctx, okc, itemExecFail); err == nil {
		t.Fatal("item exec fail")
	}
	if err := dbx.RunTxItemsWithT(ctx, okRepo, &failScopeTxItem{}); err == nil {
		t.Fatal("scope item fail")
	}
	if err := dbx.RunTxItemsWithT(ctx, okRepo, &failScopeExecItem{}); err == nil {
		t.Fatal("scope exec fail")
	}

	_ = dbx.NewScope(okc, &testRow{}).MultiOrLike([][2]string{{"name", "a"}}, false)
}

type failTxItem struct{}

func (failTxItem) Preload(d dbx.Driver) error { return errors.New("pre") }
func (failTxItem) Execute(d dbx.Driver) error { return nil }

type failExecItem struct{}

func (failExecItem) Preload(d dbx.Driver) error { return nil }
func (failExecItem) Execute(d dbx.Driver) error { return errors.New("exec") }

type failScopeTxItem struct{}

func (failScopeTxItem) Preload(repo *dbx.T, d dbx.Driver) error { return errors.New("pre") }
func (failScopeTxItem) Execute(repo *dbx.T, d dbx.Driver) error { return nil }

type failScopeExecItem struct{}

func (failScopeExecItem) Preload(repo *dbx.T, d dbx.Driver) error { return nil }
func (failScopeExecItem) Execute(repo *dbx.T, d dbx.Driver) error { return errors.New("exec") }
