package gorm_test

import (
	"testing"

	"gorm.io/driver/sqlite"
	stdgorm "gorm.io/gorm"

	"github.com/ml444/gkit/dbx"
	gormdriver "github.com/ml444/gkit/dbx/gorm"
	"github.com/ml444/gkit/dbx/pagination"
)

type djTestModel struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	CreatedAt uint32 `json:"created_at"`
	Name      string `json:"name"`
}

func (djTestModel) TableName() string { return "dj_test_model" }

func djTestDB(t *testing.T) dbx.Conn {
	t.Helper()
	db, err := stdgorm.Open(sqlite.Open("file:djtest?mode=memory&cache=shared"), &stdgorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&djTestModel{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("DELETE FROM dj_test_model").Error; err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	return gormdriver.NewConn(db)
}

func seedDJModels(t *testing.T, conn dbx.Conn, n int) {
	t.Helper()
	for i := 1; i <= n; i++ {
		if err := dbx.NewScope(conn, &djTestModel{}).Create(&djTestModel{Name: "user"}); err != nil {
			t.Fatalf("seed row %d: %v", i, err)
		}
	}
}

func wrapGormScope(t *testing.T, conn dbx.Conn, scope *dbx.Scope) *gormdriver.GormScope {
	t.Helper()
	gs, ok := gormdriver.AsGormScope(scope)
	if !ok {
		t.Fatal("expected gorm scope")
	}
	return gs
}

func TestCanDeferredJoin(t *testing.T) {
	conn := djTestDB(t)
	scope := dbx.NewScope(conn, &djTestModel{})

	prev := gormdriver.DeferredJoinOffsetThreshold
	gormdriver.DeferredJoinOffsetThreshold = gormdriver.DefaultDeferredJoinOffsetThreshold
	defer func() { gormdriver.DeferredJoinOffsetThreshold = prev }()

	gs := wrapGormScope(t, conn, scope)

	t.Run("below threshold", func(t *testing.T) {
		_, _, ok := gs.CanDeferredJoin(999)
		if ok {
			t.Fatal("expected false below threshold")
		}
	})

	t.Run("at threshold without order", func(t *testing.T) {
		pk, order, ok := gs.CanDeferredJoin(1000)
		if !ok || pk != "id" || order != "id ASC" {
			t.Fatalf("got pk=%q order=%q ok=%v", pk, order, ok)
		}
	})

	t.Run("non-pk order falls back", func(t *testing.T) {
		s := dbx.NewScope(conn, &djTestModel{}).Order("created_at DESC")
		gs2 := wrapGormScope(t, conn, s)
		_, _, ok := gs2.CanDeferredJoin(1000)
		if ok {
			t.Fatal("expected false for non-pk order")
		}
	})
}

func TestDeferredJoinResultConsistency(t *testing.T) {
	conn := djTestDB(t)
	const total = 105
	seedDJModels(t, conn, total)

	prev := gormdriver.DeferredJoinOffsetThreshold
	gormdriver.DeferredJoinOffsetThreshold = gormdriver.DefaultDeferredJoinOffsetThreshold
	defer func() { gormdriver.DeferredJoinOffsetThreshold = prev }()

	page, size := uint32(101), uint32(10)
	offset := int(size * (page - 1))

	var direct []*djTestModel
	if err := dbx.NewScope(conn, &djTestModel{}).Order("id ASC").Limit(int(size)).Offset(offset).Find(&direct); err != nil {
		t.Fatalf("direct query: %v", err)
	}

	scope := dbx.NewScope(conn, &djTestModel{}).Order("id ASC")
	gs := wrapGormScope(t, conn, scope)
	var deferred []*djTestModel
	if _, err := gs.PaginationQueryWithOpt(&deferred, &pagination.Pagination{Page: page, Size: size, SkipCount: true}); err != nil {
		t.Fatalf("deferred pagination: %v", err)
	}

	if len(direct) != len(deferred) {
		t.Fatalf("len mismatch: direct=%d deferred=%d", len(direct), len(deferred))
	}
	for i := range direct {
		if direct[i].ID != deferred[i].ID {
			t.Fatalf("row %d: direct id=%d deferred id=%d", i, direct[i].ID, deferred[i].ID)
		}
	}
}

func TestDeferredJoinScopeReuse(t *testing.T) {
	conn := djTestDB(t)
	seedDJModels(t, conn, 5)

	prev := gormdriver.DeferredJoinOffsetThreshold
	gormdriver.DeferredJoinOffsetThreshold = 1
	defer func() { gormdriver.DeferredJoinOffsetThreshold = prev }()

	scope := dbx.NewScope(conn, &djTestModel{})
	gs := wrapGormScope(t, conn, scope)
	var list []*djTestModel
	if _, err := gs.PaginationQueryWithOpt(&list, &pagination.Pagination{Page: 2, Size: 2, SkipCount: true}); err != nil {
		t.Fatalf("pagination: %v", err)
	}

	total, err := scope.Count()
	if err != nil {
		t.Fatalf("count after pagination: %v", err)
	}
	if total != 5 {
		t.Fatalf("count = %d, want 5", total)
	}
}
