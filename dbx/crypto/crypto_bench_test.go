package crypto_test

import (
	"testing"

	"github.com/ml444/gkit/dbx/crypto"
)

type EncryptRow struct {
	ID    uint64 `json:"id" gorm:"primaryKey"`
	Phone string `json:"phone" gorm:"column:phone;encrypt:searchable"`
	Card  string `json:"id_card" gorm:"column:id_card;encrypt:storage"`
	Name  string `json:"name" gorm:"column:name"`
}

func (EncryptRow) TableName() string { return "v2_encrypt_rows" }

func testKeyRing() crypto.KeyRing {
	return crypto.KeyRing{
		ActiveID: "1",
		Keys: map[string][]byte{
			"1": []byte("0123456789abcdef0123456789abcdef"),
			"0": []byte("abcdef0123456789abcdef0123456789"),
		},
	}
}

func BenchmarkProcessorCopy(b *testing.B) {
	p, err := crypto.NewProcessor(crypto.Config{
		KeyRing: testKeyRing(),
		Fields:  []crypto.FieldSpec{{Column: "phone", Struct: "Phone", Mode: crypto.ModeSearchable}},
		Mutate:  crypto.MutateCopy,
	}, &EncryptRow{})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row := &EncryptRow{Phone: "138"}
		enc, err := p.EncryptValue(row)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := p.DecryptValue(enc); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProcessorInPlace(b *testing.B) {
	p, err := crypto.NewProcessor(crypto.Config{
		KeyRing: testKeyRing(),
		Fields:  []crypto.FieldSpec{{Column: "phone", Struct: "Phone", Mode: crypto.ModeSearchable}},
		Mutate:  crypto.MutateInPlace,
	}, &EncryptRow{})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row := &EncryptRow{Phone: "138"}
		enc, err := p.EncryptValue(row)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := p.DecryptValue(enc); err != nil {
			b.Fatal(err)
		}
	}
}
