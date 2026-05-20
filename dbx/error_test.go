package dbx

import (
	"errors"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func TestIsDuplicateErr(t *testing.T) {
	mysqlErr := &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry"}
	if !IsDuplicateErr(mysqlErr) {
		t.Fatal("expected mysql duplicate to match")
	}
	if IsDuplicateErr(errors.New("other error")) {
		t.Fatal("unexpected duplicate match")
	}
	if !IsDuplicateErr(errors.New("duplicate key value violates unique constraint")) {
		t.Fatal("expected postgres-style message to match")
	}
}
