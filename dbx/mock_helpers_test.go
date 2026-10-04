package dbx_test

import (
	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/dbx/crypto"
)

func testKeyRing() crypto.KeyRing {
	return crypto.KeyRing{
		ActiveID: "1",
		Keys: map[string][]byte{
			"1": []byte("0123456789abcdef0123456789abcdef"),
			"0": []byte("abcdef0123456789abcdef0123456789"),
		},
	}
}

func encryptRowCrypto(mutate ...crypto.MutatePolicy) dbx.TOption {
	policy := crypto.MutateInPlace
	if len(mutate) > 0 {
		policy = mutate[0]
	}
	return dbx.WithTCrypto(crypto.Config{
		KeyRing: testKeyRing(),
		Fields:  []crypto.FieldSpec{{Column: "name", Struct: "Name", Mode: crypto.ModeSearchable}},
		Mutate:  policy,
	})
}

func nameSearchableCrypto(mutate ...crypto.MutatePolicy) dbx.TOption {
	policy := crypto.MutateInPlace
	if len(mutate) > 0 {
		policy = mutate[0]
	}
	return dbx.WithTCrypto(crypto.Config{
		KeyRing: testKeyRing(),
		Fields:  []crypto.FieldSpec{{Column: "name", Struct: "Name", Mode: crypto.ModeSearchable}},
		Mutate:  policy,
	})
}

func dualFieldCrypto() dbx.TOption {
	return dbx.WithTCrypto(crypto.Config{
		KeyRing: testKeyRing(),
		Fields: []crypto.FieldSpec{
			{Column: "name", Struct: "Name", Mode: crypto.ModeSearchable},
			{Column: "secret", Struct: "Secret", Mode: crypto.ModeStorage},
		},
		Mutate: crypto.MutateInPlace,
	})
}
