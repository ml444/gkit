package gorm_test

/*
import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	user "github.com/ml444/gkit/cmd/protoc-gen-go-gorm/tests/user"
	"github.com/ml444/gkit/dbx"
	gormdriver "github.com/ml444/gkit/dbx/gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var benchDBSeq uint64

func benchSQLiteDSN() string {
	seq := atomic.AddUint64(&benchDBSeq, 1)
	return fmt.Sprintf("file:bench%d?mode=memory&cache=shared", seq)
}

func benchUserFixture() *user.User {
	return benchUserFixtureWithID(1)
}

func benchUserFixtureWithID(id uint64) *user.User {
	phone := fmt.Sprintf("bench-phone-%d", id)
	return &user.User{
		Id:          id,
		IsValidated: true,
		Name:        "bench-user",
		Age:         ptrUint32(30),
		CreatedAt:   20260101,
		UpdatedAt:   120000,
		DeletedAt:   0,
		State:       user.User_StateLogin,
		Tags:        []string{"a", "b"},
		GroupTags:   map[string]uint64{"g1": 1},
		Phone:       &phone,
	}
}

var benchConvSink any

func ptrUint32(v uint32) *uint32 { return &v }

func Benchmark_ToORM_only(b *testing.B) {
	src := benchUserFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchConvSink = src.ToORM()
	}
}

func Benchmark_ToSource_only(b *testing.B) {
	src := benchUserFixture()
	tm := src.ToORM().(*user.TUser)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchConvSink = tm.ToSource()
	}
}

func benchUserDB(b *testing.B) *gorm.DB {
	b.Helper()
	db, err := gorm.Open(sqlite.Open(benchSQLiteDSN()), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}
	// user.TUser uses MySQL FULLTEXT/time column tags; use a sqlite-compatible layout.
	if err := db.Exec(`CREATE TABLE my_user (
		id INTEGER PRIMARY KEY,
		is_validated numeric,
		name TEXT,
		age INTEGER,
		created_at TEXT,
		updated_at TEXT,
		deleted_at TEXT,
		detail1 TEXT,
		detail_blob1 BLOB,
		avatar BLOB,
		group_tags TEXT,
		client_login_info TEXT,
		state INTEGER,
		phone TEXT UNIQUE
	)`).Error; err != nil {
		b.Fatal(err)
	}
	// Seed with NULL timestamps so sqlite can scan rows for GetOne/List benchmarks.
	if err := db.Exec(`INSERT INTO my_user (id, is_validated, name, age, state, phone, group_tags)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, true, "bench-user", 30, int(user.User_StateLogin), "bench-seed-phone", `{"g1":1}`,
	).Error; err != nil {
		b.Fatal(err)
	}
	return db
}

func benchUserT(db *gorm.DB) *dbx.T {
	return dbx.NewT[user.User](func() dbx.Conn { return gormdriver.NewConn(db) })
}

func Benchmark_Create_full(b *testing.B) {
	db := benchUserDB(b)
	tbl := benchUserT(db)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u := benchUserFixtureWithID(uint64(i + 1000))
		if err := tbl.Create(ctx, u); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_GetOne_full(b *testing.B) {
	db := benchUserDB(b)
	tbl := benchUserT(db)
	ctx := context.Background()
	dst := &user.User{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := tbl.GetOne(ctx, dst, uint64(1)); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_ListPage_full(b *testing.B) {
	db := benchUserDB(b)
	tbl := benchUserT(db)
	ctx := context.Background()
	var list []*user.User
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list = list[:0]
		if _, err := tbl.ListWithPagination(ctx, &list, nil, 1, 20); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_BatchCreate_full(b *testing.B) {
	db := benchUserDB(b)
	tbl := benchUserT(db)
	ctx := context.Background()
	const batch = 50
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := make([]*user.User, batch)
		for j := range rows {
			rows[j] = benchUserFixtureWithID(uint64(i*batch + j + 20000))
		}
		if err := tbl.BatchCreate(ctx, &rows); err != nil {
			b.Fatal(err)
		}
	}
}

type benchSimpleModel struct {
	ID   uint64 `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

func (benchSimpleModel) TableName() string { return "bench_simple" }

func Benchmark_Create_noForceT(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(benchSQLiteDSN()), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}
	if err := db.AutoMigrate(&benchSimpleModel{}); err != nil {
		b.Fatal(err)
	}
	tbl := dbx.NewT[benchSimpleModel](func() dbx.Conn { return gormdriver.NewConn(db) })
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := &benchSimpleModel{ID: uint64(i + 1), Name: "x"}
		if err := tbl.Create(ctx, m); err != nil {
			b.Fatal(err)
		}
	}
}

func TestAllocsBaseline_ToORM(t *testing.T) {
	src := benchUserFixture()
	allocs := testing.AllocsPerRun(100, func() {
		benchConvSink = src.ToORM()
	})
	t.Logf("ToORM allocs/run: %.0f", allocs)
	if allocs == 0 {
		t.Error("expected at least 1 alloc for heap-escaped TUser")
	}
}

func Benchmark_GetOne_noForceT(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(benchSQLiteDSN()), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}
	if err := db.AutoMigrate(&benchSimpleModel{}); err != nil {
		b.Fatal(err)
	}
	if err := db.Create(&benchSimpleModel{ID: 1, Name: "seed"}).Error; err != nil {
		b.Fatal(err)
	}
	tbl := dbx.NewT[benchSimpleModel](func() dbx.Conn { return gormdriver.NewConn(db) })
	ctx := context.Background()
	dst := &benchSimpleModel{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := tbl.GetOne(ctx, dst, uint64(1)); err != nil {
			b.Fatal(err)
		}
	}
}

*/
