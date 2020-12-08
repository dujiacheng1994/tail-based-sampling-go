package utils

import (
	"crypto/md5"
	"encoding/hex"
)

var DataSourcePort string

func MD5(key string) string{
	byteKey := []byte(key)
	md5Ctx := md5.New()
	md5Ctx.Write(byteKey)
	cipherStr := md5Ctx.Sum(nil)

	return hex.EncodeToString(cipherStr)
}