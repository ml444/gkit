package dbx

import (
	"testing"

	"github.com/ml444/gkit/dbx/crypto"
)

// Benchmarks compare v1 CheckAndCrypto vs v2 Processor (Copy / InPlace).

func BenchmarkV1CheckAndCrypto(b *testing.B) {
	repo := NewT[encryptRow](func() Conn { return stubTxConn{d: stubDriver{}} },
		SetSpecifyFieldCipherMap(map[string]FieldCipher{
			"name": {StructField: "Name", Cipher: prefixCipher{}},
		}),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row := &encryptRow{Name: "alice"}
		if err := repo.CheckAndCrypto(row, CipherKindEncrypt, false); err != nil {
			b.Fatal(err)
		}
		if err := repo.CheckAndCrypto(row, CipherKindDecrypt, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkV2ProcessorCopy(b *testing.B) {
	p, err := crypto.NewProcessor(crypto.Config{
		KeyRing: testKeyRing(),
		Fields:  []crypto.FieldSpec{{Column: "phone", Struct: "Phone", Mode: crypto.ModeSearchable}},
		Mutate:  crypto.MutateCopy,
	}, &v2EncryptRow{})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row := &v2EncryptRow{Phone: "138"}
		enc, err := p.EncryptValue(row)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := p.DecryptValue(enc); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkV2ProcessorInPlace(b *testing.B) {
	p, err := crypto.NewProcessor(crypto.Config{
		KeyRing: testKeyRing(),
		Fields:  []crypto.FieldSpec{{Column: "phone", Struct: "Phone", Mode: crypto.ModeSearchable}},
		Mutate:  crypto.MutateInPlace,
	}, &v2EncryptRow{})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row := &v2EncryptRow{Phone: "138"}
		enc, err := p.EncryptValue(row)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := p.DecryptValue(enc); err != nil {
			b.Fatal(err)
		}
	}
}
