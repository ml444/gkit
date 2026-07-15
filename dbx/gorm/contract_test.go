package gorm_test

import (
	"fmt"
	"sync/atomic"
	"testing"

	"gorm.io/driver/sqlite"
	stdgorm "gorm.io/gorm"

	"github.com/ml444/gkit/dbx"
	gormdriver "github.com/ml444/gkit/dbx/gorm"
	"github.com/ml444/gkit/dbx/internal/driverstest"
)

var contractDBSeq uint64

func TestDriverContract(t *testing.T) {
	driverstest.Run(t, func(t *testing.T) dbx.Conn {
		t.Helper()
		dsn := fmt.Sprintf("file:gorm-contract-%d?mode=memory&cache=shared", atomic.AddUint64(&contractDBSeq, 1))
		db, err := stdgorm.Open(sqlite.Open(dsn), &stdgorm.Config{})
		if err != nil {
			t.Fatal(err)
		}
		if err := db.AutoMigrate(driverstest.MigrationModels()...); err != nil {
			t.Fatal(err)
		}
		return gormdriver.NewConn(db)
	})
}
