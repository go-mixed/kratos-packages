package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
)

func MD5(src io.Reader) ([]byte, error) {
	hash := md5.New()

	_, err := io.Copy(hash, src)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func MD5String(src string) string {
	out, _ := MD5(strings.NewReader(src))
	return hex.EncodeToString(out)
}

func Sha256(src string) string {
	m := sha256.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}
