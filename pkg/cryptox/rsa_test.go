package cryptox

import (
	"bytes"
	"reflect"
	"testing"
)

var testPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCVe+SGi5uDSz/4
XIeihF6cKY1MgbqL+u4GAUMsPq+vFlKyZT2Ibaqs4JT6iKVt7r8JjniLb2PZwEwR
5My2RQng9VQm0p7iFR8GgP+kr2NuP9dBzrq1kKHuOH37iE5ZY39fXJh0VnV/Rfif
clmq/5kMdw7Y7Hj6TpN6TbZkG1tgnzZVTowBZ7kRVp1UqApqpBvoYzwi5EFZgEDi
LrwdhQhWscQ22rFsXZadPNTHG/ST9cXG3Euj23gLpZkrt1EPxTffNGNoIMmrCqYr
3KTO3UW9pToDs/WDmO0MO2nquc60D9IQZHBXKxaDkTl7d+ifCnJDGGCZscddaIGh
NLwBZV1jAgMBAAECggEAEyPneYDTu0Z93OfAKEOJt7YsVQEBaV4KbzNnLfmi1ijm
EtiavebI5VTmToQDpqVcybT342ayYtXYB6yDt8z0PF09VrE+TdWFgPgwg54/fYTo
I5F8X7YyvcV3ACeOXKy8SPIaxT6y0cacVJI4QAh1SN7PxF/XB7na8VyU/5FvLFpU
ah9vF2jo5VjsF5caBAiHHSOx8E9z88GHCYiCLPri2Sq3wpnrXIcRs8FONr9JKaiQ
9e6qQxRam9ZD0S43PVy7E6jg2WsAsqowCe5ssOqSEON9apF1qqNRJL5ts98C7kJI
+ymsgjsxRjqKtsOUUVLJA1aAR0KcGUqfy6y3fMYouQKBgQDFGL0L832YqAeNjTRk
VeUoWpES0otIDeT65X2dkTyv2LdUylTlVlJhmcdSSO+LfWeNftsM6DSyph0vwma6
GZ5lneWy0dzbrQrHZi7OLEb1Ez0c7WjyWM6xDytNw300izYHVyYrv4KjCr+2JWud
LP7b+cC6JZK1Zem427bhVj5uqQKBgQDCKHQdiVyA5bjgka1ZgRM85zJ8K74NCtNO
UEuKKXflavhopEDp1QUk+N9ySsD7jeB+ZtviGypfb6pBMIJ4nrBVcFXb0lqAWD4s
ts7X5Ag2uHvnRVa0ZRrPZZPyp47iwhTFHKVQyGrbbVp2Aq20f+CIBUvDoClmsBoX
0kYRK3HvKwKBgGJdylvgldpOYhafVnqM8+WD7ct7ENBRPuqJBnxRM/x/KGBE6sHa
pxrW6MeEZyky2S+hFCoI6eQPS5m+aA6RIqCMgUsRuixY3HxP3yQ+rNs7UtDRHAN3
lxB/BZm16xMCN2DKed5zoftFLhD19BNplXir2SgOAH5P8qmz3j3wERChAoGBAJp0
IgMJLexgUxVa7iMgmlQ9u5yqE5M+hGBtYdp0KKv5z8k2uWkLC/+gd+js7N5wvCDx
5IPXhnrLUw5u76vS2YXuSm8HxPUKvdNGTf/SqHIXioGtWE9DivNn5C0J/JIJQQqZ
Qi2kcdVDBc6RTOwlOlIanG3wMF8/QlKm9RRdklJnAoGAAn6Ho8MgaTPj2OxEVbFH
v7l2QgGTwDEX8cL6nbUYUOnG8ndKwV7r/rpLI3SwUfnMhj7FUhoh4iX00L13NtB3
JWnolfgo1JrHd3S4K2mTOfiRCerUHkpM4pjFD4KD1rK7LMOX+s76uD82PbQBZO7t
80dgMTjkhBVQ9TGADd39wog=
-----END PRIVATE KEY-----`

var testPrivateKeyPKCS8 = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCVe+SGi5uDSz/4
XIeihF6cKY1MgbqL+u4GAUMsPq+vFlKyZT2Ibaqs4JT6iKVt7r8JjniLb2PZwEwR
5My2RQng9VQm0p7iFR8GgP+kr2NuP9dBzrq1kKHuOH37iE5ZY39fXJh0VnV/Rfif
clmq/5kMdw7Y7Hj6TpN6TbZkG1tgnzZVTowBZ7kRVp1UqApqpBvoYzwi5EFZgEDi
LrwdhQhWscQ22rFsXZadPNTHG/ST9cXG3Euj23gLpZkrt1EPxTffNGNoIMmrCqYr
3KTO3UW9pToDs/WDmO0MO2nquc60D9IQZHBXKxaDkTl7d+ifCnJDGGCZscddaIGh
NLwBZV1jAgMBAAECggEAEyPneYDTu0Z93OfAKEOJt7YsVQEBaV4KbzNnLfmi1ijm
EtiavebI5VTmToQDpqVcybT342ayYtXYB6yDt8z0PF09VrE+TdWFgPgwg54/fYTo
I5F8X7YyvcV3ACeOXKy8SPIaxT6y0cacVJI4QAh1SN7PxF/XB7na8VyU/5FvLFpU
ah9vF2jo5VjsF5caBAiHHSOx8E9z88GHCYiCLPri2Sq3wpnrXIcRs8FONr9JKaiQ
9e6qQxRam9ZD0S43PVy7E6jg2WsAsqowCe5ssOqSEON9apF1qqNRJL5ts98C7kJI
+ymsgjsxRjqKtsOUUVLJA1aAR0KcGUqfy6y3fMYouQKBgQDFGL0L832YqAeNjTRk
VeUoWpES0otIDeT65X2dkTyv2LdUylTlVlJhmcdSSO+LfWeNftsM6DSyph0vwma6
GZ5lneWy0dzbrQrHZi7OLEb1Ez0c7WjyWM6xDytNw300izYHVyYrv4KjCr+2JWud
LP7b+cC6JZK1Zem427bhVj5uqQKBgQDCKHQdiVyA5bjgka1ZgRM85zJ8K74NCtNO
UEuKKXflavhopEDp1QUk+N9ySsD7jeB+ZtviGypfb6pBMIJ4nrBVcFXb0lqAWD4s
ts7X5Ag2uHvnRVa0ZRrPZZPyp47iwhTFHKVQyGrbbVp2Aq20f+CIBUvDoClmsBoX
0kYRK3HvKwKBgGJdylvgldpOYhafVnqM8+WD7ct7ENBRPuqJBnxRM/x/KGBE6sHa
pxrW6MeEZyky2S+hFCoI6eQPS5m+aA6RIqCMgUsRuixY3HxP3yQ+rNs7UtDRHAN3
lxB/BZm16xMCN2DKed5zoftFLhD19BNplXir2SgOAH5P8qmz3j3wERChAoGBAJp0
IgMJLexgUxVa7iMgmlQ9u5yqE5M+hGBtYdp0KKv5z8k2uWkLC/+gd+js7N5wvCDx
5IPXhnrLUw5u76vS2YXuSm8HxPUKvdNGTf/SqHIXioGtWE9DivNn5C0J/JIJQQqZ
Qi2kcdVDBc6RTOwlOlIanG3wMF8/QlKm9RRdklJnAoGAAn6Ho8MgaTPj2OxEVbFH
v7l2QgGTwDEX8cL6nbUYUOnG8ndKwV7r/rpLI3SwUfnMhj7FUhoh4iX00L13NtB3
JWnolfgo1JrHd3S4K2mTOfiRCerUHkpM4pjFD4KD1rK7LMOX+s76uD82PbQBZO7t
80dgMTjkhBVQ9TGADd39wog=
-----END PRIVATE KEY-----
`

func TestGenRSAKey(t *testing.T) {
	type args struct {
		bits int
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			err := GenRSAKey(out, tt.args.bits)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenRSAKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("GenRSAKey() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func TestRSA(t *testing.T) {
	var (
		testString = "test"
		testBytes  = []byte("test")
	)
	x, err := NewRSA([]byte(testPrivateKeyPKCS8))
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	// test string
	s, err := x.Encrypt(testString)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	t.Logf("ciphertext: %s\n", s.(string))
	t.Logf("ciphertext: %v\n", s)
	plainText, err := x.Decrypt(s)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	if plainText != testString {
		t.Errorf("plainText: %v, want %v", plainText, testString)
	}

	// test bytes
	b, err := x.Encrypt(testBytes)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	plainText2, err := x.Decrypt(b)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	if !reflect.DeepEqual(plainText2, testBytes) {
		t.Errorf("plainText: %v, want %v", plainText2, testBytes)
	}

	// test PKCS1v15
	bb, err := x.EncryptPKCS1v15(testBytes)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	plainText3, err := x.DecryptPKCS1v15(bb)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	if !reflect.DeepEqual(plainText3, testBytes) {
		t.Errorf("plainText: %v, want %v", plainText2, testBytes)
	}
}
