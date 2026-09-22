package dbx

import (
	"github.com/ml444/gkit/dbx/crypto"
)

// T2Option configures a T2 repository. Returning an error aborts NewT2.
type T2Option func(*T)

// WithCrypto attaches a v2 field-encryption Processor.
// If cfg.Fields is empty, fields are discovered from gorm encrypt:storage|searchable tags.
func WithTCrypto(cfg crypto.Config) T2Option {
	return func(t *T) {
		fields := cfg.Fields
		if len(fields) == 0 {
			discovered, err := crypto.DiscoverFields(t.ormModel)
			if err != nil {
				panic(err.Error())
			}
			fields = discovered
		}
		cfg.Fields = fields
		if len(cfg.Fields) == 0 {
			// No fields: leave processor nil (fast path). Still validate keyring only if fields expected?
			// Spec: empty fields => no-op. Skip NewProcessor when no fields.
			return 
		}
		p, err := crypto.NewProcessor(cfg, t.ormModel)
		if err != nil {
			panic(err.Error())
		}
		t.crypto = p
		t.disableDecrypt = cfg.DisableDecrypt
	}
}

// WithDecrypt controls whether read paths decrypt encrypted fields.
func WithTDecrypt(enabled bool) T2Option {
	return func(t *T) {
		t.disableDecrypt = !enabled
	}
}

func WithTCreateBatchSize(size int) T2Option {
	return func(t *T) {
		t.BatchCreateSize = size
	}
}

func WithTPrimaryKey(pk string) T2Option {
	return func(t *T) {
		t.PrimaryKey = pk
	}
}

func WithTGenerateIDFunc(idFunc GenerateIDFunc) T2Option {
	return func(t *T) {
		t.IdGenerator = idFunc
	}
}

func WithTNotFoundErrCode(code int32) T2Option {
	return func(t *T) {
		t.NotFoundErrCode = code
	}
}

func WithTIgnoreNotFoundErr() T2Option {
	return func(t *T) {
		t.IgnoreNotFoundErr = true
	}
}
