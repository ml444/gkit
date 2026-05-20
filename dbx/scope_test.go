package dbx

import (
	"testing"
)

func Test_isNonEmptySlice(t *testing.T) {
	type args struct {
		slice interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"1", args{slice: []string{"a", "b"}}, true},
		{"2", args{slice: []uint64{}}, false},
		{"3", args{slice: nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNonEmptySlice(tt.args.slice); got != tt.want {
				t.Errorf("isNonEmptySlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Benchmark_isNonEmptySlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isNonEmptySlice([]string{"a", "b"})
	}
}

func TestScope_Incr(t *testing.T) {
	s := NewScope(testGetDB(), &testUser{})
	username := "TestScope_Incr"
	// create user
	u := &testUser{Name: username, Age: 10}
	err := s.Create(u)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID == 0 {
		t.Fatal("user id should not be 0")
	}
	// incr age
	err = s.Eq("id", u.ID).UpdateColumnWithIncr("age", 1)
	if err != nil {
		t.Fatal(err)
	}
	// check age
	user1 := &testUser{}
	err = s.Where("name", username).First(user1)
	if err != nil {
		t.Fatal(err)
	}
	if user1.Age != 11 {
		t.Errorf("age should be 11, got %d", user1.Age)
	}
	t.Logf("==before user: %v", u)
	t.Logf("==after user: %v", user1)

	// incr age -2
	err = NewScope(testGetDB(), &testUser{}).Eq("id", u.ID).UpdateColumnWithIncr("age", -2)
	if err != nil {
		t.Fatal(err)
	}
	// check age
	user2 := &testUser{}
	err = s.Where("name", username).First(user2)
	if err != nil {
		t.Fatal(err)
	}
	if user2.Age != 9 {
		t.Errorf("age should be 11, got %d", user2.Age)
	}
	t.Logf("==before user: %v", user1)
	t.Logf("==after user: %v", user2)
}
