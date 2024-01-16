package main

import (
	"errors"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"tests/user"
)

func initDb() (*gorm.DB, error) {
	var err error
	// pwd := os.Getenv("DB_PWD")
	// DBURI := fmt.Sprintf("root:%s@tcp(192.168.64.9:3306)/test_orm?charset=utf8mb4&parseTime=True&loc=Local", pwd)
	DBURI := os.Getenv("SERVICE_DB_URI")
	if DBURI == "" {
		println("not found db uri")
		return nil, errors.New("not found dbURI")
	}
	db, err := gorm.Open(mysql.Open(DBURI), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(
		&user.ModelUser{},
		&user.ModelRecord{},
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
	println(db.Error)
	u := user.ModelUser{
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
		ClientLoginInfo: user.User_ClientLoginInfo{
			1: &user.UserInfo{LoginCount: 123},
		},
		IgnoreData: user.User_IgnoreData{
			2: &user.UserInfo{LoginCount: 234, LastLoginIp: "172.123.1.11"},
		},
		State: 2,
		// Phone: new(string),
	}
	err = db.Create(&u).Error
	if err != nil {
		println(err.Error())
		return
	}

	fmt.Println(u.Name)

	var m user.ModelUser
	err = db.Where("id", u.Id).First(&m).Error
	if err != nil {
		println(err.Error())
		return
	}
	fmt.Printf("%+v \n", m.ToSource())

	var uu []*user.ModelUser
	err = db.Clauses(m.UseIndex2IdxName()).Where("name LIKE 'test%'").Where("deleted_at > ?", 0).Find(&uu).Error
	if err != nil {
		println(err.Error())
		return
	}
	if len(uu) == 0 {
		panic("error")
	}
}
