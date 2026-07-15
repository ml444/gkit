package main

import (
	"errors"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	user "github.com/ml444/gkit/cmd/protoc-gen-go-gorm/tests/user"
)

func initDb() (*gorm.DB, error) {
	DBURI := os.Getenv("SERVICE_DB_URI")
	if DBURI == "" {
		return nil, errors.New("SERVICE_DB_URI is required for integration test")
	}
	db, err := gorm.Open(mysql.Open(DBURI), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(
		&user.TUser{},
		&user.TRecord{},
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	db, err := initDb()
	if err != nil {
		panic(err)
	}

	u := user.User{
		IsValidated: true,
		Name:        "test1",
		Age:         new(uint32),
		CreatedAt:   1234567890,
		UpdatedAt:   14567890,
		DeletedAt:   1234567890,
		Detail1: &user.UserInfo{
			LoginCount:  11,
			LastLoginIp: "198.168.0.1",
			LastLoginAt: 1234567890,
			GroupIds:    []uint64{123, 456},
		},
		DetailBlob1: &user.User_DetailBlob{
			LoginCount:  11,
			LastLoginIp: "198.168.0.1",
			LastLoginAt: 1234567890,
			GroupIds:    []int64{123, 456},
		},
		Avatar: []byte("yaedwihf;wjpgtjrogjbrkeek"),
		Tags:   []string{"foo", "bar"},
		GroupTags: map[string]uint64{
			"foo": 123,
			"Bar": 456,
		},
		ClientLoginInfo: map[int32]*user.UserInfo{
			1: {LoginCount: 123},
		},
		IgnoreData: map[uint64]*user.UserInfo{
			2: {LoginCount: 234, LastLoginIp: "172.123.1.11"},
		},
		State: 2,
	}
	tu := u.ToORM().(*user.TUser)
	err = db.Create(tu).Error
	if err != nil {
		panic(err)
	}

	fmt.Println(tu.Name)

	var loaded user.TUser
	err = db.Where("id", tu.Id).First(&loaded).Error
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", loaded.ToSource())

	var uu []*user.TUser
	err = db.Clauses(loaded.UseIndex2IdxName()).Where("name LIKE 'test%'").Where("deleted_at > ?", 0).Find(&uu).Error
	if err != nil {
		panic(err)
	}
	if len(uu) == 0 {
		panic("expected at least one row")
	}
}
