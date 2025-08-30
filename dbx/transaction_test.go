package dbx

import (
	"context"
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTxGo(t *testing.T) {
	type args struct {
		db       *gorm.DB
		executes []TxHandler
	}
	normalExecute := func(tx *gorm.DB) error {
		fmt.Println("normal execute")
		err := tx.Model(&testOrmModel{}).Create(&testOrmModel{Name: "test"}).Error
		if err != nil {
			return err
		}
		return nil
	}
	panicExecute := func(tx *gorm.DB) error {
		fmt.Println("panic execute")
		panic("test panic")
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"ok", args{db: testGetDB(), executes: []TxHandler{normalExecute}}, false},
		{"panic", args{db: testGetDB(), executes: []TxHandler{panicExecute}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TxGo(context.Background(), tt.args.db, tt.args.executes...); (err != nil) != tt.wantErr {
				t.Errorf("TxGo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScopeTxGo(t *testing.T) {
	type args struct {
		db        *gorm.DB
		callbacks []TxCallback
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScopeTxGo(context.Background(), tt.args.db, tt.args.callbacks...); (err != nil) != tt.wantErr {
				t.Errorf("ScopeTxGo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type testUser struct {
	ID   uint
	Name string
	Age  uint
}

type testKind struct {
	ID   uint
	Kind string
}

func testDB() *gorm.DB {
	if tx == nil {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		tx = db
		db.AutoMigrate(&testUser{}, &testKind{})
	}
	return tx
}

func TestTxCreateMultiModels(t *testing.T) {
	type args struct {
		db *gorm.DB
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"ok", args{db: testDB()}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if err := TxCreateMultiModels(ctx, tt.args.db, &testUser{1, "far", 18}, &testKind{1, "foo"}); (err != nil) != tt.wantErr {
				t.Errorf("ScopeTxGo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
