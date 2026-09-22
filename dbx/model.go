package dbx

// type cipherKind int

// const (
// 	cipherKindEncrypt cipherKind = 0
// 	cipherKindDecrypt cipherKind = 1
// )

// // CipherKind identifies encrypt/decrypt operations for exported test APIs.
// type CipherKind = cipherKind

// const (
// 	CipherKindEncrypt = cipherKindEncrypt
// 	CipherKindDecrypt = cipherKindDecrypt
// )

// type ICipher interface {
// 	Encrypt(plaintext any) (any, error)
// 	Decrypt(ciphertext any) (any, error)
// }

type IModel interface {
	ToORM() ITModel
}
type ITModel interface {
	ToSource() IModel
	ForceTModel() bool
	CopyToSource(dst IModel) error
	CopyToSourceIgnoreEmpty(dst IModel) error
}

type GenerateIDFunc func() uint64
