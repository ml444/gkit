package dbx

import (
	"context"
	"errors"
	"reflect"
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
