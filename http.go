package sdk

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CorsMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func JsonMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	}
}

func DbMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("databaseConn", db)
		c.Next()
	}
}

func ApiMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		//log.Println(c.RemoteIP())
		//
		//key := c.Request.Header.Get("ApiKey")
		//hash := c.Request.Header.Get("Hash")
		//time := c.Request.Header.Get("Time")
		//
		//var account models.Account
		//
		//tx := db.Debug().Where("key = ? AND blocked = ?", key, false).Find(&account)
		//
		//if tx.Error != nil || tx.RowsAffected < 1 {
		//	c.JSON(http.StatusUnauthorized, gin.H{"Message": "Unauthorized"})
		//	c.Writer.WriteHeaderNow()
		//	c.Abort()
		//	return
		//}
		//
		//if account.CheckHash(hash, time) {
		//	c.Set("apiAccount", &account)
		//} else {
		//	c.JSON(http.StatusUnauthorized, gin.H{"Message": "Wrong api key or hash or timestamp"})
		//	c.Writer.WriteHeaderNow()
		//	c.Abort()
		//	return
		//}
		//
		//ip := c.Request.Header.Get("X-Real-IP")
		//
		//if ip == "" {
		//	ip = c.RemoteIP()
		//}
		//
		//if !account.CheckIp(db, ip) {
		//	c.JSON(http.StatusUnauthorized, gin.H{"Message": "Wrong server ip"})
		//	c.Writer.WriteHeaderNow()
		//	c.Abort()
		//	return
		//}

		c.Next()
	}
}
