package crypto

import (
	"errors"
	"testing"
)

type userRow struct {
	Phone string `json:"phone"`
	Name  string `json:"name"`
}

func TestProcessorStructCopyEncryptDecrypt(t *testing.T) {
	cfg := Config{
		KeyRing: testRing(t),
		Fields: []FieldSpec{
			{Column: "phone", Struct: "Phone", Mode: ModeSearchable},
		},
		Mutate: MutateCopy,
	}
	p, err := NewProcessor(cfg, &userRow{})
	if err != nil {
		t.Fatal(err)
	}
	in := &userRow{Phone: "138", Name: "n"}
	outAny, err := p.EncryptValue(in)
	if err != nil {
		t.Fatal(err)
	}
	out := outAny.(*userRow)
	if in.Phone != "138" {
		t.Fatal("MutateCopy must not change input")
	}
	if out.Phone == "138" || out.Name != "n" {
		t.Fatalf("out = %#v", out)
	}
	plainAny, err := p.DecryptValue(out)
	if err != nil {
		t.Fatal(err)
	}
	plain := plainAny.(*userRow)
	if plain.Phone != "138" {
		t.Fatalf("plain = %#v", plain)
	}
}

func TestProcessorMapAndQuery(t *testing.T) {
	cfg := Config{
		KeyRing: testRing(t),
		Fields: []FieldSpec{
			{Column: "phone", Struct: "Phone", Mode: ModeSearchable},
			{Column: "id_card", Struct: "IDCard", Mode: ModeStorage},
		},
		Mutate: MutateCopy,
	}
	p, err := NewProcessor(cfg, &struct {
		Phone  string `json:"phone"`
		IDCard string `json:"id_card"`
	}{})
	if err != nil {
		t.Fatal(err)
	}
	q, err := p.EncryptQueryMap(map[string]any{"phone": "138"})
	if err != nil {
		t.Fatal(err)
	}
	if q["phone"] == "138" {
		t.Fatal("phone should be encrypted")
	}
	if _, err := p.EncryptQueryMap(map[string]any{"id_card": "x"}); !errors.Is(err, ErrNotSearchable) {
		t.Fatalf("want ErrNotSearchable, got %v", err)
	}
}

func TestProcessorInPlace(t *testing.T) {
	cfg := Config{
		KeyRing: testRing(t),
		Fields:  []FieldSpec{{Column: "phone", Struct: "Phone", Mode: ModeSearchable}},
		Mutate:  MutateInPlace,
	}
	p, _ := NewProcessor(cfg, &userRow{})
	in := &userRow{Phone: "138"}
	out, err := p.EncryptValue(in)
	if err != nil {
		t.Fatal(err)
	}
	if out != in || in.Phone == "138" {
		t.Fatalf("inplace failed: %#v", in)
	}
}

func TestProcessorLegacyDecrypt(t *testing.T) {
	cfg := Config{
		KeyRing: testRing(t),
		Fields:  []FieldSpec{{Column: "phone", Struct: "Phone", Mode: ModeSearchable}},
		LegacyDecrypt: func(mode Mode, ciphertext string) (string, error) {
			if ciphertext == "legacy:138" {
				return "138", nil
			}
			return "", ErrDecryptFailed
		},
	}
	p, _ := NewProcessor(cfg, &userRow{})
	row := &userRow{Phone: "legacy:138"}
	out, err := p.DecryptValue(row)
	if err != nil {
		t.Fatal(err)
	}
	if out.(*userRow).Phone != "138" {
		t.Fatalf("%#v", out)
	}
}
