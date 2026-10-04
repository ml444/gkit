package dbx

import (
	"context"
	"testing"
)

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

	badSU := &ScopeUpdateItem{Model: &storageEncryptRow{ID: 1}, Where: map[string]any{"secret": "s"}}
	if err := badSU.Preload(ctx, repo, d); err == nil {
		t.Fatal("scoped update preload should reject storage encrypted where")
	}
	badModelUpdate := &ScopeUpdateItem{Model: &encryptRow{ID: 1, Name: "a"}}
	if err := badModelUpdate.Execute(ctx, brokenEncryptionRepo(t), d); err == nil {
		t.Fatal("scoped update execute should return encrypt error")
	}
	badSave := &ScopeSaveItem{Model: &storageEncryptRow{ID: 1}, Where: map[string]any{"secret": "s"}}
	if err := badSave.Preload(ctx, repo, d); err == nil {
		t.Fatal("scoped save preload should reject storage encrypted where")
	}
	if err := (&ScopeSaveItem{Model: &encryptRow{ID: 2, Name: "x"}}).Execute(ctx, brokenEncryptionRepo(t), d); err == nil {
		t.Fatal("scoped save execute should return encrypt error")
	}
}
