package utils

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
)

func MD5(src io.Reader) ([]byte, error) {
	hash := md5.New()

	chunkSize := 65536

	for buf, reader := make([]byte, chunkSize), bufio.NewReader(src); ; {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		hash.Write(buf[:n])
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
