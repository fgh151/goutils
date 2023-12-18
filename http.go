package sdk

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
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

func AccountMiddleware(whiteList []string) gin.HandlerFunc {

	wl := append([]string{"/metrics", "/healthz", "/readyz"}, whiteList...)

	return func(c *gin.Context) {

		for _, s := range wl {
			if ok, _ := regexp.MatchString(s, c.Request.URL.Path); ok {
				c.Next()
				return
			}
		}

		doNext := true

		key := c.Request.Header.Get("ApiKey")
		hash := c.Request.Header.Get("Hash")
		time := c.Request.Header.Get("Time")

		req, err := http.NewRequest(http.MethodGet, os.Getenv("DNS_ACCOUNT")+"/apiaccount/check", nil)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Cant check account"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			doNext = false
			return
		}

		q := req.URL.Query()
		q.Add("ApiKey", key)
		q.Add("Hash", hash)
		q.Add("Time", time)
		q.Add("Origin", c.Request.Header.Get("Origin"))
		req.URL.RawQuery = q.Encode()

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Cant check api account"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			doNext = false
			return
		}

		if res.StatusCode != http.StatusOK {
			c.JSON(http.StatusUnauthorized, res.Body)
			c.Writer.WriteHeaderNow()
			c.Abort()
			doNext = false
			return
		}

		if doNext {
			c.Next()
		}
	}
}

func RbacMiddleware(role string) gin.HandlerFunc {
	return func(c *gin.Context) {

		log.Println(c.RemoteIP())

		header := c.Request.Header.Get("Authorization")

		if header == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Messed Authorization header"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}

		splitToken := strings.Split(header, "Bearer ")

		if len(splitToken) < 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Messed Bearer token"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}

		token := strings.TrimSpace(splitToken[1])

		req, err := http.NewRequest(http.MethodGet, os.Getenv("DNS_USERS")+"/user/can/"+token+"/"+role, nil)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Cant check user permissions"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}

		res, err := http.DefaultClient.Do(req)

		if res.StatusCode != http.StatusOK {
			c.JSON(http.StatusUnauthorized, res.Body)
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}

		c.Next()
	}
}

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := uuid.New().String()
		c.Set("traceId", u)

		c.Next()
	}
}
