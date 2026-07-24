package crypto

import "errors"

var (
	ErrUnknownKey    = errors.New("dbx/crypto: unknown key id")
	ErrDecryptFailed = errors.New("dbx/crypto: decrypt failed")
	ErrNotSearchable = errors.New("dbx/crypto: field is not searchable")
	ErrInvalidConfig = errors.New("dbx/crypto: invalid config")
)
