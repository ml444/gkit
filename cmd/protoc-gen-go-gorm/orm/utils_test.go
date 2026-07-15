package orm

import "testing"

func TestIgnoreEmptyCondition(t *testing.T) {
	tests := []struct {
		oldTyp, newTyp, name, want string
	}{
		{"string", "string", "Name", `x.Name != ""`},
		{"bool", "bool", "Ok", "x.Ok"},
		{"*uint32", "*uint32", "Age", "x.Age != nil"},
		{"[]uint64", "UserInfo_GroupIds", "GroupIds", "len(x.GroupIds) > 0"},
		{"[]string", "User_Tags", "Tags", "len(x.Tags) > 0"},
		{"map[string]uint64", "User_GroupTags", "GroupTags", "len(x.GroupTags) > 0"},
		{"uint64", "uint64", "ID", "x.ID != 0"},
	}
	for _, tt := range tests {
		if got := IgnoreEmptyCondition(tt.oldTyp, tt.newTyp, tt.name); got != tt.want {
			t.Errorf("IgnoreEmptyCondition(%q, %q, %q) = %q, want %q", tt.oldTyp, tt.newTyp, tt.name, got, tt.want)
		}
	}
}
