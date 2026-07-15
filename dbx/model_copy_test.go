package dbx_test

import (
	"context"
	"testing"

	"github.com/ml444/gkit/dbx"
)

type srcModel struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func (m *srcModel) ToORM() dbx.ITModel {
	return &ormModel{ID: m.ID, Name: m.Name, force: true}
}

type ormModel struct {
	ID   uint64 `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`

	force bool
}

func (ormModel) TableName() string { return "orm_models" }

func (o *ormModel) ToSource() dbx.IModel {
	return &srcModel{ID: o.ID, Name: o.Name}
}

func (o *ormModel) ForceTModel() bool { return o.force }

func (o *ormModel) CopyToSource(dst dbx.IModel) error {
	m := dst.(*srcModel)
	m.ID, m.Name = o.ID, o.Name
	return nil
}

func (o *ormModel) CopyToSourceIgnoreEmpty(dst dbx.IModel) error {
	m := dst.(*srcModel)
	if o.ID != 0 {
		m.ID = o.ID
	}
	if o.Name != "" {
		m.Name = o.Name
	}
	return nil
}

func TestRepositoryCopiesForcedORMModels(t *testing.T) {
	d, conn := testConn()
	repo := dbx.NewT[srcModel](func() dbx.Conn { return conn })
	ctx := context.Background()

	created := &srcModel{ID: 1, Name: "alice"}
	if err := repo.Create(ctx, created); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got := d.Tables["orm_models"][0]["name"]; got != "alice" {
		t.Fatalf("stored name = %v", got)
	}

	var one srcModel
	if err := repo.GetOne(ctx, &one, uint64(1)); err != nil {
		t.Fatalf("GetOne: %v", err)
	}
	if one != (srcModel{ID: 1, Name: "alice"}) {
		t.Fatalf("GetOne result = %#v", one)
	}

	list := []srcModel{{Name: "keep"}}
	if err := repo.ListAll(ctx, &list, nil); err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(list) != 1 || list[0] != (srcModel{ID: 1, Name: "alice"}) {
		t.Fatalf("ListAll result = %#v", list)
	}

	batch := []*srcModel{{ID: 2, Name: "bob"}, {ID: 3, Name: "cara"}}
	if err := repo.BatchCreate(ctx, &batch); err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}
	if len(d.Tables["orm_models"]) != 3 || batch[1].Name != "cara" {
		t.Fatalf("BatchCreate result = %#v", batch)
	}

	if rows, err := repo.Update(ctx, &srcModel{Name: "updated"}, "id = ?", uint64(1)); err != nil || rows != 1 {
		t.Fatalf("Update = %d, %v", rows, err)
	}
	if rows, err := repo.UpdateByPk(ctx, &srcModel{Name: "again"}, uint64(1)); err != nil || rows != 1 {
		t.Fatalf("UpdateByPk = %d, %v", rows, err)
	}
	if err := repo.Save(ctx, &srcModel{ID: 4, Name: "saved"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	var paged []srcModel
	if _, err := repo.ListWithPagination(ctx, &paged, nil, 1, 2); err != nil || len(paged) != 2 {
		t.Fatalf("ListWithPagination = %#v, %v", paged, err)
	}
	if count, err := repo.Count(ctx, "id = ?", uint64(1)); err != nil || count != 1 {
		t.Fatalf("Count query = %d, %v", count, err)
	}
	if err := repo.DeleteByPk(ctx, uint64(1)); err != nil {
		t.Fatalf("DeleteByPk: %v", err)
	}
	if err := repo.DeleteByWhere(ctx, "id = ?", uint64(2)); err != nil {
		t.Fatalf("DeleteByWhere: %v", err)
	}
}
