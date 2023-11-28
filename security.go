package sdk

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"math/rand"
)

func GetPasswordHash(password string) string {
	hashes := md5.New()
	hashes.Write([]byte(password))
	return hex.EncodeToString(hashes.Sum(nil))
}

func RandString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func AccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()
	}
}
