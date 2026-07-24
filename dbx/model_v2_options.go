package dbx

import (
	"github.com/ml444/gkit/dbx/crypto"
)

// T2Option configures a T2 repository. Returning an error aborts NewT2.
type T2Option func(*T2) error

// WithCrypto attaches a v2 field-encryption Processor.
// If cfg.Fields is empty, fields are discovered from gorm encrypt:storage|searchable tags.
func WithCrypto(cfg crypto.Config) T2Option {
	return func(t *T2) error {
		fields := cfg.Fields
		if len(fields) == 0 {
			discovered, err := crypto.DiscoverFields(t.ormModel)
			if err != nil {
				return err
			}
			fields = discovered
		}
		cfg.Fields = fields
		if len(cfg.Fields) == 0 {
			// No fields: leave processor nil (fast path). Still validate keyring only if fields expected?
			// Spec: empty fields => no-op. Skip NewProcessor when no fields.
			return nil
		}
		p, err := crypto.NewProcessor(cfg, t.ormModel)
		if err != nil {
			return err
		}
		t.crypto = p
		t.disableDecrypt = cfg.DisableDecrypt
		return nil
	}
}

// WithDecrypt controls whether read paths decrypt encrypted fields.
func WithDecrypt(enabled bool) T2Option {
	return func(t *T2) error {
		t.disableDecrypt = !enabled
		return nil
	}
}

func WithT2CreateBatchSize(size int) T2Option {
	return func(t *T2) error {
		t.BatchCreateSize = size
		return nil
	}
}

func WithT2PrimaryKey(pk string) T2Option {
	return func(t *T2) error {
		t.PrimaryKey = pk
		return nil
	}
}

func WithT2GenerateIDFunc(idFunc GenerateIDFunc) T2Option {
	return func(t *T2) error {
		t.IdGenerator = idFunc
		return nil
	}
}

func WithT2NotFoundErrCode(code int32) T2Option {
	return func(t *T2) error {
		t.NotFoundErrCode = code
		return nil
	}
}

func WithT2IgnoreNotFoundErr() T2Option {
	return func(t *T2) error {
		t.IgnoreNotFoundErr = true
		return nil
	}
}
