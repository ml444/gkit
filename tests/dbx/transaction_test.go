package dbx

import (
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/tests/user"
)

func testDb() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func TestTxGo(t *testing.T) {
	type args struct {
		db  *gorm.DB
		fns []dbx.TxHandler
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_db",
			args: args{
				db: testDb(),
				fns: []dbx.TxHandler{
					func(tx *gorm.DB) error {
						err := tx.AutoMigrate(&user.User{})
						if err != nil {
							return err
						}
						u := &user.User{Name: "test", Age: 18}
						defer t.Log(u)
						return tx.Model(&user.User{}).Create(&u).Error
					},
					func(tx *gorm.DB) error {
						return tx.Model(&user.User{}).Where("id", 1).Updates(&user.User{Name: "test_update", Age: 20}).Error
					},
					func(tx *gorm.DB) error {
						var u user.User
						err := tx.Model(&user.User{}).First(&u, 1).Error
						if err != nil {
							return err
						}
						if u.Name != "test_update" {
							return errors.New("name not equal")
						}
						if u.Age != 20 {
							return errors.New("age not equal")
						}
						return nil
					},
					func(tx *gorm.DB) error {
						return tx.Migrator().DropTable(&user.User{})
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test_tx",
			args: args{
				db:  testDb().Begin(),
				fns: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := dbx.TxGo(tt.args.db, tt.args.fns...); (err != nil) != tt.wantErr {
				t.Errorf("TxGo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestScopeTxGo(t *testing.T) {
	type args struct {
		db  *gorm.DB
		fns []dbx.TxCallback
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_db",
			args: args{
				db: testDb(),
				fns: []dbx.TxCallback{
					func() (model interface{}, execute func(scope *dbx.Scope) error) {
						return &user.User{}, func(scope *dbx.Scope) error {
							err := scope.DB.AutoMigrate(&user.User{})
							if err != nil {
								return err
							}
							u := &user.User{Name: "test", Age: 18}
							defer t.Log(u)
							err = scope.Create(&u)
							if err != nil {
								return err
							}
							if u.Id != 1 {
								return errors.New("id not equal")
							}
							return nil
						}
					},
					func() (model interface{}, execute func(scope *dbx.Scope) error) {
						return &user.User{}, func(scope *dbx.Scope) error {
							return scope.Where("id", 1).Update(&user.User{Name: "test_update", Age: 20})
						}
					},
					func() (model interface{}, execute func(scope *dbx.Scope) error) {
						return &user.User{}, func(scope *dbx.Scope) error {
							var u user.User
							err := scope.First(&u, 1)
							if err != nil {
								return err
							}
							if u.Name != "test_update" {
								return errors.New("name not equal")
							}
							if u.Age != 20 {
								return errors.New("age not equal")
							}
							return nil
						}
					},
					func() (model interface{}, execute func(scope *dbx.Scope) error) {
						return &user.User{}, func(scope *dbx.Scope) error {
							return scope.DB.Migrator().DropTable(&user.User{})
						}
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test_tx",
			args: args{
				db:  testDb().Begin(),
				fns: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := dbx.ScopeTxGo(tt.args.db, tt.args.fns...); (err != nil) != tt.wantErr {
				t.Errorf("TxGo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
