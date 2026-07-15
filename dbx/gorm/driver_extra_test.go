package gorm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ml444/gkit/dbx"
	gormdriver "github.com/ml444/gkit/dbx/gorm"
	dbxmock "github.com/ml444/gkit/dbx/mock"
)

func TestDriverBuilderOptionsAndErrors(t *testing.T) {
	db := openGorm(t)
	if err := db.Create(&[]gormRow{{Name: "a", Score: 1}, {Name: "b", Score: 2}}).Error; err != nil {
		t.Fatal(err)
	}
	conn := gormdriver.NewConn(db)
	s := dbx.NewScope(conn, &gormRow{})
	var rows []gormRow
	complex := s.Select("gorm_rows.id", "gorm_rows.name", "gorm_rows.score").Omit("score").Or("gorm_rows.name = ?", "a").
		Joins("LEFT JOIN gorm_rows other ON other.id = gorm_rows.id").Group("gorm_rows.name").
		Having("COUNT(*) >= ?", 1).Order("gorm_rows.name ASC").Orders(dbx.OrderColumn{Field: "gorm_rows.id", Desc: true})
	if err := complex.Find(&rows); err != nil {
		t.Fatalf("builder query: %v", err)
	}
	if err := s.Eq("name", "a").UpdateColumnWithIncr("score", 3); err != nil {
		t.Fatalf("increment: %v", err)
	}
	if err := s.Eq("name", "a").UpdateColumnWithIncr("score", -1); err != nil {
		t.Fatalf("decrement: %v", err)
	}
	var changed gormRow
	if err := s.Eq("name", "a").First(&changed); err != nil || changed.Score != 3 {
		t.Fatalf("increment result = %+v, %v", changed, err)
	}
	if err := s.Eq("name", "a").Save(&gormRow{ID: changed.ID, Name: "saved", Score: 7}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := s.Eq("name", "saved").Scan(&changed); err != nil || changed.Score != 7 {
		t.Fatalf("scan result = %+v, %v", changed, err)
	}
	// SetForUpdate and include-deleted are builder-only options; constructing the
	// query exercises their GORM translation without relying on SQLite lock syntax.
	_ = s.SetForUpdate().SetIncludeDeleted()
	if _, ok := gormdriver.RawDB(conn.Driver(context.Background())); !ok {
		t.Fatal("RawDB did not unwrap gorm driver")
	}
	manager, ok := dbx.AsTxManager(conn)
	if !ok {
		t.Fatal("grom connection is not a transaction manager")
	}
	tx, err := manager.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := gormdriver.RawDB(tx); !ok {
		t.Fatal("RawDB did not unwrap transaction driver")
	}
	if err := manager.Rollback(tx); err != nil {
		t.Fatal(err)
	}
	if _, ok := gormdriver.RawDB(dbxmock.New()); ok {
		t.Fatal("RawDB accepted mock driver")
	}
	var absent gormRow
	if err := s.Eq("id", -1).First(&absent); !errors.Is(err, dbx.ErrRecordNotFound) {
		t.Fatalf("First missing = %v", err)
	}
}
