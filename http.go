package sdk

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/runetid/go-sdk/models"
	"gorm.io/gorm"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strconv"
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

func intersect(n1, n2 *net.IPNet) bool {
	return n2.Contains(n1.IP) || n1.Contains(n2.IP)
}

func AccountMiddleware(whiteList []string) gin.HandlerFunc {

	wl := append([]string{"/metrics", "/healthz", "/readyz"}, whiteList...)

	return func(c *gin.Context) {

		addrs, err := net.InterfaceAddrs()
		_, remoteCIDR, err := net.ParseCIDR(c.ClientIP() + "/24")
		if err == nil {

			for _, a := range addrs {
				_, localCIDR, pe := net.ParseCIDR(a.String() + "/24")
				if pe == nil && intersect(localCIDR, remoteCIDR) {
					c.Next()
					return
				}
			}
		}

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
			log.Println(err.Error() + " " + c.Request.Header.Get("referer"))
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
			log.Println(err.Error() + " " + c.Request.Header.Get("referer"))
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

		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		var response models.ApiAccountResponse
		if err := json.Unmarshal(body, &response); err != nil { // Parse []byte to go struct pointer
			log.Println("Can not unmarshal api account response " + string(body))
		} else {
			c.Set("event_id", response.Data.EventId)
			c.Set("role", response.Data.Role)
		}

		if doNext {
			c.Next()
		}
	}
}

func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exist := c.Get("role")

		if exist == false || role != "admin" {
			log.Println("Admin only method :" + c.Request.Header.Get("referer"))
			c.JSON(http.StatusUnauthorized, gin.H{"message": "admin only method"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}
	}
}

func RbacMiddleware(role string) gin.HandlerFunc {
	return func(c *gin.Context) {

		log.Println(c.RemoteIP())

		header := c.Request.Header.Get("Authorization")

		if header == "" {
			log.Println("RBAC middleware: Messed Authorization header " + c.Request.Header.Get("referer"))
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Messed Authorization header"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}

		splitToken := strings.Split(header, "Bearer ")

		if len(splitToken) < 2 {
			log.Println("RBAC Missed Bearer token " + c.Request.Header.Get("referer"))
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Missed Bearer token"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}

		token := strings.TrimSpace(splitToken[1])

		req, err := http.NewRequest(http.MethodGet, os.Getenv("DNS_USERS")+"/user/can/"+token+"/"+role, nil)

		if err != nil {
			log.Println("RBAC Missed cant create request to user microservice " + c.Request.Header.Get("referer"))
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Cant check user permissions"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}

		res, err := http.DefaultClient.Do(req)

		if res.StatusCode != http.StatusOK {
			log.Println("RBAC Missed cant fetch user microservice " + c.Request.Header.Get("referer"))
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

func UserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		token := c.Request.Header.Get("Authorization")
		splitToken := strings.Split(token, "Bearer ")

		if len(splitToken) > 1 {

			token = splitToken[1]

			if token != "" {
				c.Set("token", token)

				u, err := RawFetchModel(http.MethodGet, os.Getenv("DNS_USER")+"/user/byToken/"+token, nil, c.Value("traceId").(string), models.User{})

				if err == nil {
					c.Set("user", u)
				}
			}
		}

		c.Next()
	}
}

func EventMiddle(c *gin.Context) {
	key := c.Request.Header.Get("ApiKey")

	if key != "" {
		ars, err := RawFetchModel[models.ApiAccountResponse](http.MethodGet, os.Getenv("DNS_ACCOUNT")+"/apiaccount/check/"+key, nil, c.Value("traceId").(string), models.ApiAccountResponse{})
		if err == nil {
			event, err := RawFetchModel[models.Event](http.MethodGet, os.Getenv("DNS_EVENT")+"/event/"+strconv.FormatInt(ars.Data.EventId, 10), nil, c.Value("traceId").(string), models.Event{})
			if err != nil {
				c.Set("event", event)
			}
		}
	}

	c.Next()
}

func RawFetchModel[T any](method string, url string, body io.Reader, traceId string, model T) (T, error) {
	resp, err := RawFetch(method, url, body, traceId)
	if err == nil {
		b, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(b, &model)
		if err == nil {
			return model, nil
		}
	}

	err = errors.New("not found")

	return model, err
}

func RawFetch(method string, url string, body io.Reader, traceId string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	req.Header.Set("X-Trace-Id", traceId)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
func FetchInternal(url string, traceId string) (interface{}, error) {
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Trace-Id", traceId)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

type ErrorMessage struct {
	Message string `json:"message" example:"Модель не найдена"`
}

func ErrorHandler(c *gin.Context, code int, text string) {
	c.JSON(code, ErrorMessage{Message: text})
	c.Writer.WriteHeaderNow()
	c.Abort()
	return
}
