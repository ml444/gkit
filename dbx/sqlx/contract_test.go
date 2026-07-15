package sqlx_test

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/dbx/internal/driverstest"
	sqlxdriver "github.com/ml444/gkit/dbx/sqlx"
)

var contractDBSeq uint64

func TestDriverContract(t *testing.T) {
	driverstest.Run(t, func(t *testing.T) dbx.Conn {
		t.Helper()
		dsn := fmt.Sprintf("file:sqlx-contract-%d?mode=memory&cache=shared", atomic.AddUint64(&contractDBSeq, 1))
		raw, err := sqlx.Connect("sqlite3", dsn)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = raw.Close() })
		for _, ddl := range []string{
			"CREATE TABLE dbx_contract_model (id INTEGER PRIMARY KEY, name TEXT NOT NULL, age INTEGER NOT NULL)",
			"CREATE TABLE dbx_contract_kind (id INTEGER PRIMARY KEY, kind TEXT NOT NULL)",
		} {
			if _, err := raw.Exec(ddl); err != nil {
				t.Fatal(err)
			}
		}
		return sqlxdriver.NewConn(raw)
	})
}
