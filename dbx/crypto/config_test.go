package crypto

import (
	"errors"
	"reflect"
	"testing"
)

type tagRow struct {
	Phone  string `json:"phone" gorm:"column:phone;encrypt:searchable"`
	IDCard string `json:"id_card" gorm:"column:id_card;encrypt:storage"`
	Name   string `json:"name" gorm:"column:name;encrypt:true"` // v1 tag — ignored
}

func TestDiscoverFields(t *testing.T) {
	fields, err := DiscoverFields(&tagRow{})
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 {
		t.Fatalf("fields = %#v", fields)
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := Config{
		KeyRing: KeyRing{ActiveID: "1", Keys: map[string][]byte{"1": make([]byte, 32)}},
		Fields:  []FieldSpec{{Column: "phone", Struct: "Phone", Mode: ModeSearchable}},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := cfg
	bad.Fields = []FieldSpec{{Mode: ModeStorage}}
	if err := bad.Validate(); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("want ErrInvalidConfig, got %v", err)
	}
	badMode := cfg
	badMode.Fields = []FieldSpec{{Column: "phone", Struct: "Phone", Mode: Mode(99)}}
	if err := badMode.Validate(); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad mode: %v", err)
	}
	duplicate := cfg
	duplicate.Fields = []FieldSpec{
		{Column: "phone", Struct: "Phone", Mode: ModeSearchable},
		{Column: "phone", Struct: "Name", Mode: ModeStorage},
	}
	if err := duplicate.Validate(); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("duplicate column: %v", err)
	}
	badRing := cfg
	badRing.KeyRing = KeyRing{}
	if err := badRing.Validate(); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad ring: %v", err)
	}
}

func TestDiscoverFieldsBranches(t *testing.T) {
	if _, err := DiscoverFields(nil); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("nil model: %v", err)
	}
	if _, err := DiscoverFields(1); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("scalar model: %v", err)
	}
	type discoverRow struct {
		Token string `gorm:"encrypt:storage"`
		Code  string `gorm:"encrypt: searchable "`
		Old   string `gorm:"encrypt: true"`
		Skip  string
		hide  string `gorm:"encrypt:searchable"`
	}
	fields, err := DiscoverFields(&discoverRow{})
	if err != nil {
		t.Fatal(err)
	}
	want := []FieldSpec{
		{Column: "token", Struct: "Token", Mode: ModeStorage},
		{Column: "code", Struct: "Code", Mode: ModeSearchable},
	}
	if !reflect.DeepEqual(fields, want) {
		t.Fatalf("fields = %#v, want %#v", fields, want)
	}
	if mode, ok := parseEncryptMode("column:x;encrypt:unknown"); ok || mode != 0 {
		t.Fatalf("unknown encrypt mode = %v, %t", mode, ok)
	}
	if mode, ok := parseEncryptMode("column:x"); ok || mode != 0 {
		t.Fatalf("missing encrypt mode = %v, %t", mode, ok)
	}
	if got := camelToSnake("APIKey"); got != "a_p_i_key" {
		t.Fatalf("camelToSnake = %q", got)
	}
}
