package crypto

import (
	"errors"
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
}
