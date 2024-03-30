package dbx

import (
	"testing"

	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/tests/user"
)

func TestCURD(t *testing.T) {
	t.Log("TestCURD")
	tx := testDb()
	err := tx.AutoMigrate(&user.User{})
	if err != nil {
		t.Fatal(err)
	}
	scope := dbx.NewScope(tx, &user.User{})
	u := &user.User{Name: "test", Age: 18}
	err = scope.Create(u)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(u)

}
