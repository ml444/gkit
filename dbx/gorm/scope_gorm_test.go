package gorm_test

import (
	"testing"

	"github.com/ml444/gkit/dbx"
	gormdriver "github.com/ml444/gkit/dbx/gorm"
	dbxmock "github.com/ml444/gkit/dbx/mock"
	"github.com/ml444/gkit/dbx/pagination"
	"gorm.io/gorm/clause"
)

func TestGormScopeWrappersAndPagination(t *testing.T) {
	db := openGorm(t)
	if err := db.Create(&[]gormRow{{Name: "a", Score: 1}, {Name: "b", Score: 2}}).Error; err != nil {
		t.Fatal(err)
	}
	s := dbx.NewScope(gormdriver.NewConn(db), &gormRow{})
	gs, ok := gormdriver.AsGormScope(s)
	if !ok || gs == nil {
		t.Fatal("AsGormScope did not recognize gorm connection")
	}
	if _, ok := gormdriver.AsGormScope(dbx.NewScope(dbxmock.New().Conn(), "items")); ok {
		t.Fatal("AsGormScope recognized mock connection")
	}
	if _, ok := gormdriver.AsGormScope(nil); ok {
		t.Fatal("AsGormScope accepted nil scope")
	}
	if gormdriver.TxConn(nil) != nil || gormdriver.WrapScope(nil, db) != nil {
		t.Fatal("nil scope helpers should return nil")
	}
	if gormdriver.NewGormScope(db, &gormRow{}) == nil {
		t.Fatal("NewGormScope returned nil")
	}
	gs.Preload("Missing").Clauses(clause.OnConflict{DoNothing: true})
	if gs.Association("Missing") == nil {
		t.Fatal("Association returned nil")
	}
	if gs.Unscoped() == nil {
		t.Fatal("Unscoped returned nil")
	}
	gs = gormdriver.WrapScope(s, db)
	var rows []gormRow
	page, err := gs.PaginationQueryWithOpt(&rows, &pagination.Pagination{Page: 0, Size: 1})
	if err != nil || page.Page != 1 || page.Size != 1 || page.Total != 2 || len(rows) != 1 {
		t.Fatalf("PaginationQueryWithOpt = %+v, rows=%d, err=%v", page, len(rows), err)
	}
	page, err = gs.PaginationQueryWithOpt(&rows, &pagination.Pagination{SkipCount: true, Size: uint32(dbx.MaxLimit + 1)})
	if err != nil || !page.SkipCount || page.Size != uint32(dbx.MaxLimit) {
		t.Fatalf("skip-count pagination = %+v, %v", page, err)
	}
	page, err = gs.PaginationQueryWithOpt(&rows, &pagination.Pagination{})
	if err != nil || page.Page != 1 || page.Size != uint32(dbx.DefaultLimit) {
		t.Fatalf("default pagination = %+v, %v", page, err)
	}
	var nilScope *gormdriver.GormScope
	if _, err := nilScope.PaginationQueryWithOpt(&rows, nil); err == nil {
		t.Fatal("nil scope accepted pagination")
	}
}
