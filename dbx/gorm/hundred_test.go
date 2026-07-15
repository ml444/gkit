package gorm

import (
	"context"
	"testing"

	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/dbx/pagination"
	"gorm.io/driver/sqlite"
	stdgorm "gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

func TestBeginClosedDBAndDeferredJoinEdges(t *testing.T) {
	db, err := stdgorm.Open(sqlite.Open("file:gorm-hundred?mode=memory&cache=shared"), &stdgorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&coverageModel{}); err != nil {
		t.Fatal(err)
	}
	conn := NewConn(db).(*Conn)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	_ = sqlDB.Close()
	if _, err := conn.Begin(context.Background()); err == nil {
		t.Fatal("expected begin error on closed db")
	}

	db2, err := stdgorm.Open(sqlite.Open("file:gorm-hundred2?mode=memory&cache=shared"), &stdgorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db2.AutoMigrate(&coverageModel{}); err != nil {
		t.Fatal(err)
	}
	conn2 := NewConn(db2)
	old := DeferredJoinOffsetThreshold
	DeferredJoinOffsetThreshold = 1
	t.Cleanup(func() { DeferredJoinOffsetThreshold = old })

	// Model() non-nil that fails Parse (string table name)
	gs := &GormScope{
		Scope:  dbx.NewScope(conn2, "coverage_models"),
		gormDB: &stdgorm.DB{Config: db2.Config, Statement: &stdgorm.Statement{DB: db2, Clauses: map[string]clause.Clause{}}},
	}
	_, _, _ = gs.canDeferredJoin(10)

	// Model() nil, stmt.Model Parse fails
	gsNilModel := &GormScope{
		Scope: &dbx.Scope{},
		gormDB: &stdgorm.DB{Config: db2.Config, Statement: &stdgorm.Statement{
			DB:      db2,
			Model:   make(chan int),
			Clauses: map[string]clause.Clause{},
		}},
	}
	_, _, _ = gsNilModel.canDeferredJoin(10)

	if order, ok := deferredJoinOrderExpr("id", &stdgorm.Statement{Clauses: map[string]clause.Clause{
		"ORDER BY": {Expression: clause.Eq{}},
	}}); ok || order != "" {
		t.Fatalf("bad order expression: %q %v", order, ok)
	}
	if order, ok := deferredJoinOrderExpr("id", &stdgorm.Statement{Clauses: map[string]clause.Clause{
		"ORDER BY": {Expression: clause.OrderBy{}},
	}}); !ok || order != "id ASC" {
		t.Fatalf("empty order columns: %q %v", order, ok)
	}
	if _, ok := deferredJoinOrderExpr("id", &stdgorm.Statement{Clauses: map[string]clause.Clause{
		"ORDER BY": {Expression: clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "id"}},
			{Column: clause.Column{Name: "name"}},
		}}},
	}}); ok {
		t.Fatal("multi order accepted")
	}
	if _, _, ok := parseOrderByColumn(clause.OrderByColumn{Column: clause.Column{Name: "   ", Raw: true}}); ok {
		t.Fatal("blank raw order accepted")
	}

	gs3 := &GormScope{Scope: dbx.NewScope(conn2, &coverageModel{}), gormDB: db2.Session(&stdgorm.Session{NewDB: true}).Model(&coverageModel{})}
	_ = gs3.gormDB.Statement.Parse(&coverageModel{})
	gs3.gormDB.Statement.Schema = &schema.Schema{PrimaryFields: []*schema.Field{{DBName: "id"}, {DBName: "name"}}}
	_, _, _ = gs3.canDeferredJoin(10)

	db3, err := stdgorm.Open(sqlite.Open("file:gorm-hundred3?mode=memory&cache=shared"), &stdgorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db3.AutoMigrate(&coverageModel{}); err != nil {
		t.Fatal(err)
	}
	gsPag := NewGormScope(db3, &coverageModel{})
	sql3, _ := db3.DB()
	_ = sql3.Close()
	var list []coverageModel
	if _, err := gsPag.PaginationQueryWithOpt(&list, &pagination.Pagination{Page: 1, Size: 10}); err == nil {
		t.Fatal("expected pagination count error")
	}
	if _, err := NewGormScope(db3, &coverageModel{}).PaginationQueryWithOpt(&list, &pagination.Pagination{Page: 1, Size: 10, SkipCount: true}); err == nil {
		t.Fatal("expected find error")
	}
}
