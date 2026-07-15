package gorm_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/ml444/gkit/dbx"
	gormdriver "github.com/ml444/gkit/dbx/gorm"
	dbxmock "github.com/ml444/gkit/dbx/mock"
	"gorm.io/driver/sqlite"
	stdgorm "gorm.io/gorm"
)

var gormTestSeq uint64

type gormRow struct {
	ID    int64 `gorm:"primaryKey"`
	Name  string
	Score int64
}

func (gormRow) TableName() string { return "gorm_rows" }

func openGorm(t *testing.T) *stdgorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:gorm-x-%d?mode=memory&cache=shared", atomic.AddUint64(&gormTestSeq, 1))
	db, err := stdgorm.Open(sqlite.Open(dsn), &stdgorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&gormRow{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestConnTransactionManager(t *testing.T) {
	db := openGorm(t)
	conn := gormdriver.NewConn(db)
	c := conn.(interface {
		DB() *stdgorm.DB
		Driver(context.Context) dbx.Driver
		Begin(context.Context, ...dbx.TxOption) (dbx.Driver, error)
		Commit(dbx.Driver) error
		Rollback(dbx.Driver) error
	})
	if c.DB() != db || c.Driver(context.Background()) == nil {
		t.Fatal("connection did not expose the underlying database")
	}
	tx, err := c.Begin(context.Background(), dbx.WithTxOptions(sql.TxOptions{ReadOnly: true}))
	if err != nil || c.Commit(tx) != nil {
		t.Fatalf("begin/commit = %v", err)
	}
	tx, err = c.Begin(context.Background())
	if err != nil || c.Rollback(tx) != nil {
		t.Fatalf("begin/rollback = %v", err)
	}
	wrong := dbxmock.New()
	if err := c.Commit(wrong); !errors.Is(err, stdgorm.ErrInvalidTransaction) {
		t.Fatalf("Commit wrong driver = %v", err)
	}
	if err := c.Rollback(wrong); !errors.Is(err, stdgorm.ErrInvalidTransaction) {
		t.Fatalf("Rollback wrong driver = %v", err)
	}
}
