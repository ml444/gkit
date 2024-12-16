package cryptox

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5(s string) string {
	m := md5.New()
	_, _ = m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}
