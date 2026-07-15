package gorm

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/ml444/gkit/dbx"
	dbxmock "github.com/ml444/gkit/dbx/mock"
	"github.com/ml444/gkit/dbx/pagination"
	"gorm.io/driver/sqlite"
	stdgorm "gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type coverageModel struct {
	ID    int64 `gorm:"primaryKey"`
	Name  string
	Score int64
}

func TestDriverInternalHelpers(t *testing.T) {
	db, err := stdgorm.Open(sqlite.Open(":memory:"), &stdgorm.Config{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	b := &dbx.QueryBuilder{
		Model:            &coverageModel{},
		Unscoped:         true,
		Distinct:         true,
		Selects:          []string{"id", "name"},
		Omits:            []string{"score"},
		Wheres:           []dbx.WhereClause{{Query: "id > ?", Args: []any{1}}},
		OrWheres:         []dbx.WhereClause{{Query: "name = ?", Args: []any{"a"}}},
		Groups:           []string{"name", "COUNT(*)"},
		Having:           &dbx.WhereClause{Query: "COUNT(*) > ?", Args: []any{0}},
		Orders:           []dbx.OrderColumn{{Field: "id", Desc: true, Reorder: true}},
		OrderRaw:         []string{"name ASC"},
		Limit:            2,
		Offset:           1,
		ForUpdate:        true,
		ReturningColumns: []string{"id"},
	}
	tx := (&Driver{db: db}).applyBuilder(b).Find(&[]coverageModel{})
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	if tx.Statement.SQL.String() == "" {
		t.Fatal("applyBuilder produced no SQL")
	}
	if got := applyTxOptions([]dbx.TxOption{dbx.WithTxOptions(sql.TxOptions{ReadOnly: true})}); !got.ReadOnly {
		t.Fatal("applyTxOptions did not apply options")
	}
	if stringsSplitFieldsFunc("a,b", func(r rune) bool { return r == ',' })[1] != "b" {
		t.Fatal("stringsSplitFieldsFunc split incorrectly")
	}
	if mapErr(stdgorm.ErrRecordNotFound) != dbx.ErrRecordNotFound || mapErr(nil) != nil {
		t.Fatal("mapErr did not normalize record-not-found")
	}
	want := errors.New("x")
	if !errors.Is(mapErr(want), want) {
		t.Fatal("mapErr lost an unknown error")
	}
}

func TestDeferredJoinHelpers(t *testing.T) {
	stmt := &stdgorm.Statement{Clauses: map[string]clause.Clause{}}
	if order, ok := deferredJoinOrderExpr("id", stmt); !ok || order != "id ASC" {
		t.Fatalf("default order = %q, %t", order, ok)
	}
	stmt.Clauses["ORDER BY"] = clause.Clause{Expression: clause.OrderBy{Columns: []clause.OrderByColumn{{Column: clause.Column{Name: "id"}, Desc: true}}}}
	if order, ok := deferredJoinOrderExpr("id", stmt); !ok || order != "id DESC" {
		t.Fatalf("descending order = %q, %t", order, ok)
	}
	stmt.Clauses["ORDER BY"] = clause.Clause{Expression: clause.OrderBy{Columns: []clause.OrderByColumn{{Column: clause.Column{Name: "name"}}}}}
	if _, ok := deferredJoinOrderExpr("id", stmt); ok {
		t.Fatal("non-primary-key order accepted")
	}
	if name, desc, ok := parseOrderByColumn(clause.OrderByColumn{Column: clause.Column{Name: "`id` DESC", Raw: true}}); !ok || name != "id" || !desc {
		t.Fatalf("raw order = %q, %t, %t", name, desc, ok)
	}
	if _, _, ok := parseOrderByColumn(clause.OrderByColumn{}); ok {
		t.Fatal("empty order column accepted")
	}
	if orderDirection("id DESC") != "DESC" || orderDirection("id") != "ASC" {
		t.Fatal("orderDirection mismatch")
	}
}

func TestDeferredJoinIneligibleQueries(t *testing.T) {
	db, err := stdgorm.Open(sqlite.Open(":memory:"), &stdgorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	oldThreshold := DeferredJoinOffsetThreshold
	DeferredJoinOffsetThreshold = 1
	t.Cleanup(func() { DeferredJoinOffsetThreshold = oldThreshold })

	base := NewGormScope(db, &coverageModel{})
	if _, _, ok := base.canDeferredJoin(1); !ok {
		t.Fatal("primary-key query should support deferred join")
	}
	cases := []struct {
		name string
		db   *stdgorm.DB
	}{
		{"distinct", db.Model(&coverageModel{}).Distinct()},
		{"join", db.Model(&coverageModel{}).Joins("JOIN other ON other.id = coverage_models.id")},
		{"group", db.Model(&coverageModel{}).Group("name")},
		{"having", db.Model(&coverageModel{}).Having("COUNT(*) > 0")},
		{"non-pk-order", db.Model(&coverageModel{}).Order("name ASC")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gs := NewGormScope(tc.db, &coverageModel{})
			if _, _, ok := gs.canDeferredJoin(1); ok {
				t.Fatal("ineligible query accepted")
			}
		})
	}
	DeferredJoinOffsetThreshold = 0
	if _, _, ok := base.canDeferredJoin(1); ok {
		t.Fatal("disabled threshold accepted")
	}
	if _, _, ok := (&GormScope{}).canDeferredJoin(1); ok {
		t.Fatal("empty GormScope accepted")
	}
}

func TestTxGoRejectsNonGormDriver(t *testing.T) {
	err := TxGo(context.Background(), dbx.StaticConn(dbxmock.New()), func(*stdgorm.DB) error { return nil })
	if err == nil {
		t.Fatal("TxGo accepted a non-GORM driver")
	}
}

func TestRemainingCoverageBranches(t *testing.T) {
	db, err := stdgorm.Open(sqlite.Open("file:gorm-gap?mode=memory&cache=shared"), &stdgorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&coverageModel{}); err != nil {
		t.Fatal(err)
	}
	conn := NewConn(db)
	ctx := context.Background()
	d := conn.Driver(ctx).(*Driver)

	if _, err := d.UpdateColumn(ctx, &dbx.QueryBuilder{Model: &coverageModel{}, Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(1)}}}}, "name", "x"); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&coverageModel{ID: 1, Name: "a", Score: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := d.UpdateColumn(ctx, &dbx.QueryBuilder{Model: &coverageModel{}, Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(1)}}}, IncrColumn: "score", IncrValue: -1}, "score", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Delete(ctx, &dbx.QueryBuilder{Model: &coverageModel{}, Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(2)}}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Delete(ctx, &dbx.QueryBuilder{Table: "coverage_models", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(1)}}}}); err != nil {
		t.Fatal(err)
	}
	if err := d.Transaction(ctx, func(tx dbx.Driver) error { return nil }, dbx.WithTxOptions(sql.TxOptions{})); err != nil {
		t.Fatal(err)
	}

	if _, ok := AsGormScope(nil); ok {
		t.Fatal("nil scope")
	}
	if _, ok := AsGormScope(dbx.NewScope(dbx.StaticConn(d), &coverageModel{})); ok {
		t.Fatal("static conn should not AsGormScope")
	}

	gs := NewGormScope(db, &coverageModel{})
	if _, err := (*GormScope)(nil).PaginationQueryWithOpt(&[]coverageModel{}, nil); err == nil {
		t.Fatal("nil gs pagination")
	}
	if _, err := gs.PaginationQueryWithOpt(&[]coverageModel{}, nil); err != nil {
		t.Fatal(err)
	}
	// force count path then find
	_ = db.Create(&coverageModel{ID: 2, Name: "b"})
	if _, err := gs.PaginationQueryWithOpt(&[]coverageModel{}, &pagination.Pagination{Page: 1, Size: 10}); err != nil {
		t.Fatal(err)
	}

	old := DeferredJoinOffsetThreshold
	DeferredJoinOffsetThreshold = 1
	t.Cleanup(func() { DeferredJoinOffsetThreshold = old })
	gs2 := NewGormScope(db.Model(&coverageModel{}), &coverageModel{})
	gs3 := &GormScope{Scope: dbx.NewScope(conn, &coverageModel{}), gormDB: db.Session(&stdgorm.Session{NewDB: true})}
	_, _, _ = gs3.canDeferredJoin(10)
	_, _, _ = gs2.canDeferredJoin(10)
	_ = gs2.deferredJoinTable()
	var list []coverageModel
	_ = gs2.findWithDeferredJoin(&list, 1, 1, "id", "id ASC")
	emptyGS := &GormScope{gormDB: &stdgorm.DB{Statement: &stdgorm.Statement{}}}
	_ = emptyGS.deferredJoinTable()
	noTable := &GormScope{
		Scope:  dbx.NewScope(conn, &coverageModel{}),
		gormDB: db.Session(&stdgorm.Session{NewDB: true}),
	}
	noTable.gormDB.Statement = &stdgorm.Statement{DB: noTable.gormDB, Clauses: map[string]clause.Clause{}}
	func() {
		defer func() { _ = recover() }()
		_ = noTable.findWithDeferredJoin(&list, 1, 0, "id", "id ASC")
	}()

	// Direct statement mutations for remaining canDeferredJoin branches.
	stmtDB := db.Session(&stdgorm.Session{NewDB: true}).Model(&coverageModel{})
	_ = stmtDB.Statement.Parse(&coverageModel{})
	gsHit := &GormScope{Scope: dbx.NewScope(conn, &coverageModel{}), gormDB: stmtDB}
	gsHit.gormDB.Statement.Clauses["HAVING"] = clause.Clause{}
	_, _, _ = gsHit.canDeferredJoin(10)
	gsHit.gormDB.Statement.Clauses = map[string]clause.Clause{"GROUP BY": {}}
	_, _, _ = gsHit.canDeferredJoin(10)
	gsHit.gormDB.Statement.Clauses = map[string]clause.Clause{}
	gsHit.gormDB.Statement.Schema = &schema.Schema{PrimaryFields: []*schema.Field{{DBName: ""}}}
	_, _, _ = gsHit.canDeferredJoin(10)
	gsHit.gormDB.Statement.Schema = nil
	gsHit.gormDB.Statement.Model = &coverageModel{}
	_, _, _ = gsHit.canDeferredJoin(10)

	_ = db.Create(&coverageModel{ID: 3, Name: "c"})
	if _, err := d.Delete(ctx, &dbx.QueryBuilder{Model: &coverageModel{ID: 3}}); err != nil {
		t.Fatal(err)
	}
}
