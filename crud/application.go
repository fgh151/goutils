package crud

import (
	"context"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
	"github.com/rgglez/gormcache"
	"github.com/runetid/go-sdk"
	"github.com/runetid/go-sdk/log"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type Application struct {
	Router *gin.Engine
	Db     *gorm.DB
	Logger *log.AppLogger
}

func (a Application) Run() {
	isReady := &atomic.Value{}
	isReady.Store(false)
	sdk.AppendMetrics(a.Router)

	a.Router.GET("/healthz", sdk.HealthzWithDb(a.Db))
	a.Router.GET("/readyz", gin.WrapF(sdk.Readyz(isReady)))

	a.Router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Page not found"})
	})

	done := make(chan bool)
	go a.Router.Run(os.Getenv("HTTP_ADDR"))
	isReady.Store(true)
	<-done
}

type CrudModel interface {
	List(db *gorm.DB, request ListRequest, params ...FilterParams) (interface{}, int64, error)
	GetFilterParams(c *gin.Context) []FilterParams
	Create(db *gorm.DB) (interface{}, error)
	Update(db *gorm.DB, key string) (interface{}, error)
	DecodeCreate(c *gin.Context) (interface{}, error)
	Delete(db *gorm.DB, key string) bool
	Get(db *gorm.DB, key string) (interface{}, error)
}

type BaseCrudModel struct {
}

func (u BaseCrudModel) GetFilterParams(c *gin.Context) []FilterParams {
	var p []FilterParams
	return p
}

func (u BaseCrudModel) DecodeCreate(c *gin.Context) (interface{}, error) {
	return c.Bind(u), nil
}

func (a Application) AppendListEndpoint(prefix string, entity CrudModel, middlewares ...gin.HandlerFunc) {
	a.Router.GET(prefix+"/list", func(c *gin.Context) {

		for _, middleware := range middlewares {
			middleware(c)
		}

		if len(c.Errors) > 0 {
			return
		}

		var request ListRequest
		err := c.Bind(&request)
		if err != nil {
			c.JSON(500, gin.H{"message": "Wrong limit or offset params " + err.Error()})
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}

		t, e := c.GetQueryMap("filter")

		if e == true {
			request.Filter = t
		}

		s, e := c.GetQueryMap("sort")

		if e == true {
			request.Sort = s
		}

		var m interface{}
		var cnt int64

		m, cnt, err = entity.List(a.Db, request, entity.GetFilterParams(c)...)

		if m == nil {
			m = make([]string, 0)
		}

		c.JSON(200, gin.H{"data": m, "error": err, "total": cnt})
		return
	})
}

func (a Application) AppendCreateEndpoint(prefix string, entity CrudModel, middlewares ...gin.HandlerFunc) {
	a.Router.POST(prefix, func(c *gin.Context) {

		for _, middleware := range middlewares {
			middleware(c)
		}

		if len(c.Errors) > 0 {
			return
		}

		decode, err := entity.DecodeCreate(c)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err})
			return
		}

		m, err := decode.(CrudModel).Create(a.Db)
		if err != nil {
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": m, "error": err})
		return
	})
}

func (a Application) AppendUpdateEndpoint(prefix string, entity CrudModel, middlewares ...gin.HandlerFunc) {
	a.Router.PUT(prefix, func(c *gin.Context) {

		for _, middleware := range middlewares {
			middleware(c)
		}

		if len(c.Errors) > 0 {
			return
		}

		decode, _ := entity.DecodeCreate(c)
		m, err := decode.(CrudModel).Update(a.Db, c.Param("id"))

		c.JSON(200, gin.H{"data": m, "error": err})
		return
	})
}

func (a Application) AppendDeleteEndpoint(prefix string, entity CrudModel, middlewares ...gin.HandlerFunc) {
	a.Router.DELETE(prefix, func(c *gin.Context) {

		for _, middleware := range middlewares {
			middleware(c)
		}

		if len(c.Errors) > 0 {
			return
		}

		if entity.Delete(a.Db, c.Param("id")) {

			c.JSON(http.StatusOK, gin.H{"message": "ok"})
			return
		}

		c.JSON(http.StatusConflict, gin.H{"message": "cant delete"})
		return
	})
}

func (a Application) AppendGetEndpoint(prefix string, entity CrudModel, middlewares ...gin.HandlerFunc) {
	a.Router.GET(prefix, func(c *gin.Context) {

		for _, middleware := range middlewares {
			middleware(c)
		}

		if len(c.Errors) > 0 {
			return
		}

		model, err := entity.Get(a.Db, c.Param("id"))

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"data": model, "error": err})
			return
		}

		c.JSON(200, gin.H{"data": model, "error": err})
		return
	})
}

func (a Application) AppendSwagger(prefix string) {
	a.Router.GET(prefix+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (a Application) Schedule(ctx context.Context, p time.Duration, f func(time time.Time)) {
	go Schedule(ctx, p, f)
}

func NewCrudApplication(publicRoutes []string) (*Application, error) {

	logger := log.NewAppLogger()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			NoLowerCase: true,
		},
	})

	mdb := memcache.New(os.Getenv("CACHE_SRV"))
	cache := gormcache.NewGormCache("my_cache", gormcache.NewMemcacheClient(mdb), gormcache.CacheConfig{
		TTL:    600 * time.Second,
		Prefix: "cache:",
	})

	err = db.Use(cache)
	if err == nil {
		db.Session(&gorm.Session{Context: context.WithValue(context.Background(), gormcache.UseCacheKey, true)})
	}

	r := gin.Default()
	r.Use(sdk.TraceMiddleware())
	r.Use(sdk.CorsMiddleware())
	r.Use(sdk.JsonMiddleware())
	r.Use(sdk.DbMiddleware(db))
	r.Use(sdk.AccountMiddleware(publicRoutes))

	if logger.Inner == false {
		r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
			Formatter: func(param gin.LogFormatterParams) string {
				return fmt.Sprintf("%s - [%s] %s %s %s %d %s \"%s\" %s %s\n ",
					param.ClientIP,
					param.TimeStamp.Format(time.RFC1123),
					param.Method,
					param.Path,
					param.Request.Proto,
					param.StatusCode,
					param.Latency,
					param.Request.UserAgent(),
					param.ErrorMessage,
					param.Keys["traceId"],
				)
			},
			Output:    logger.Writer(),
			SkipPaths: []string{},
		}))
		r.Use(gin.Recovery())
	}

	return &Application{
		Router: r,
		Db:     db,
		Logger: logger,
	}, err
}

type ListRequest struct {
	Limit  int               `form:"limit" binding:"required,number,min=1,max=100"`
	Offset int               `form:"offset" binding:"number"`
	Filter map[string]string `form:"filter"`
	//Pagination map[string]string `form:"pagination"`
	Sort map[string]string `form:"sort"`
}

type FilterParams struct {
	Key      string
	Value    string
	Operator string
}
