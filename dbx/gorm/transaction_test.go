package gorm_test

import (
	"context"
	"errors"
	"testing"

	gormdriver "github.com/ml444/gkit/dbx/gorm"
	"gorm.io/gorm"
)

func TestTxGoSuccessAndError(t *testing.T) {
	conn := gormdriver.NewConn(openGorm(t))
	called := false
	if err := gormdriver.TxGo(context.Background(), conn, func(tx *gorm.DB) error {
		called = true
		return tx.Create(&gormRow{Name: "ok"}).Error
	}); err != nil || !called {
		t.Fatalf("TxGo success = %v, called=%t", err, called)
	}
	want := errors.New("rollback")
	if err := gormdriver.TxGo(context.Background(), conn, func(*gorm.DB) error { return want }); !errors.Is(err, want) {
		t.Fatalf("TxGo error = %v", err)
	}
}
