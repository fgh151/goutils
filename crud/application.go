package crud

import (
	"context"
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	postgres2 "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/rgglez/gormcache"
	"github.com/runetid/go-sdk"
	"github.com/runetid/go-sdk/log"
	"strings"

	//"github.com/runetid/go-sdk/log"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	log2 "log"
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

type ApplicationConfig struct {
	PublicRoutes     []string
	DbMigrationsPath string
	DbSchema         string
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
	go a.runInternalServer()
	go a.Router.Run(os.Getenv("HTTP_ADDR"))
	isReady.Store(true)
	<-done
}

type ModelWithList interface {
	List(db *gorm.DB, request ListRequest, ctx *context.Context, params ...FilterParams) (interface{}, int64, error)
	GetFilterParams(c *gin.Context) []FilterParams
}

type ModelWithGet interface {
	Get(db *gorm.DB, key string, ctx *context.Context) (interface{}, error)
}

type ModelWithUpdate interface {
	Update(db *gorm.DB, key string, ctx *context.Context) (interface{}, error)
	DecodeCreate(c *gin.Context) (interface{}, error)
}

type ModelWithCreate interface {
	Create(db *gorm.DB, ctx *context.Context) (interface{}, error)
	DecodeCreate(c *gin.Context) (interface{}, error)
}

type ModelWithDelete interface {
	Delete(db *gorm.DB, key string, ctx *context.Context) (bool, error)
	Get(db *gorm.DB, key string, ctx *context.Context) (interface{}, error)
}

type CrudModel interface {
	List(db *gorm.DB, request ListRequest, ctx *context.Context, params ...FilterParams) (interface{}, int64, error)
	GetFilterParams(c *gin.Context) []FilterParams
	Create(db *gorm.DB, ctx *context.Context) (interface{}, error)
	Update(db *gorm.DB, key string, ctx *context.Context) (interface{}, error)
	DecodeCreate(c *gin.Context) (interface{}, error)
	Delete(db *gorm.DB, key string, ctx *context.Context) (bool, error)
	Get(db *gorm.DB, key string, ctx *context.Context) (interface{}, error)
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

func (a Application) AppendListEndpoint(prefix string, entity ModelWithList, middlewares ...gin.HandlerFunc) {
	a.Router.GET(prefix+"/list", func(c *gin.Context) {

		tx := a.Db.WithContext(c)

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

		ctx := context.WithoutCancel(c)

		m, cnt, err = entity.List(tx, request, &ctx, entity.GetFilterParams(c)...)

		if m == nil {
			m = make([]string, 0)
		}

		c.JSON(200, gin.H{"data": m, "error": err, "total": cnt})
		return
	})
}

func (a Application) AppendCreateEndpoint(prefix string, entity ModelWithCreate, middlewares ...gin.HandlerFunc) {
	a.Router.POST(prefix, func(c *gin.Context) {
		tx := a.Db.WithContext(c)

		for _, middleware := range middlewares {
			middleware(c)
		}

		if len(c.Errors) > 0 {
			return
		}

		decode, err := entity.DecodeCreate(c)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		ctx := context.WithoutCancel(c)

		m, err := decode.(ModelWithCreate).Create(tx, &ctx)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": m})
		return
	})
}

func (a Application) AppendUpdateEndpoint(prefix string, entity ModelWithUpdate, middlewares ...gin.HandlerFunc) {
	a.Router.PUT(prefix, func(c *gin.Context) {
		tx := a.Db.WithContext(c)

		for _, middleware := range middlewares {
			middleware(c)
		}

		if len(c.Errors) > 0 {
			return
		}

		decode, err := entity.DecodeCreate(c)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		ctx := context.WithoutCancel(c)

		m, err := decode.(CrudModel).Update(tx, c.Param("id"), &ctx)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": m})
		return
	})
}

func (a Application) AppendDeleteEndpoint(prefix string, entity ModelWithDelete, middlewares ...gin.HandlerFunc) {
	a.Router.DELETE(prefix, func(c *gin.Context) {
		tx := a.Db.WithContext(c)
		for _, middleware := range middlewares {
			middleware(c)
		}

		if len(c.Errors) > 0 {
			return
		}

		ctx := context.WithoutCancel(c)

		model, err := entity.Get(tx, c.Param("id"), &ctx)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Удаление невозможно " + err.Error()})
			return
		}

		entity = model.(ModelWithDelete)

		del, err := entity.Delete(a.Db, c.Param("id"), &ctx)

		if !del {
			c.JSON(http.StatusNotFound, gin.H{"message": "Удаление невозможно " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
		return
	})
}

func (a Application) AppendGetEndpoint(prefix string, entity ModelWithGet, middlewares ...gin.HandlerFunc) {
	a.Router.GET(prefix, func(c *gin.Context) {
		tx := a.Db.WithContext(c)
		for _, middleware := range middlewares {
			middleware(c)
		}

		if len(c.Errors) > 0 {
			return
		}

		ctx := context.WithoutCancel(c)

		model, err := entity.Get(tx, c.Param("id"), &ctx)

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
	return NewCrudApplicationWithConfig(ApplicationConfig{
		PublicRoutes:     publicRoutes,
		DbMigrationsPath: "",
		DbSchema:         "public",
	})
}

func NewCrudApplicationWithConfig(config ApplicationConfig) (*Application, error) {

	logger := log.NewAppLogger()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: log.NewGormLogger(&logger)})

	isTesting := strings.ToUpper(os.Getenv("ENVIRONMENT")) == "TEST"

	if err != nil {
		if !isTesting {
			panic("failed to connect database: " + err.Error())
		}
	}

	cacheSrv, hasCache := os.LookupEnv("CACHE_SRV")
	if hasCache {
		mdb := memcache.New(cacheSrv)
		cache := gormcache.NewGormCache("my_cache", gormcache.NewMemcacheClient(mdb), gormcache.CacheConfig{
			TTL:    600 * time.Second,
			Prefix: "cache:",
		})

		err = db.Use(cache)
		if err == nil {
			db.Session(&gorm.Session{Context: context.WithValue(context.Background(), gormcache.UseCacheKey, true)})
		}
	}

	if config.DbMigrationsPath != "" {
		m, merr := migrate.New(
			fmt.Sprintf("github://%s:%s@%s", os.Getenv("GH_LOGIN"), os.Getenv("GH_TOKEN"), config.DbMigrationsPath),
			fmt.Sprintf("postgres://%s:%s@%s:%s/%s?search_path=%s&sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"), config.DbSchema),
		)

		if merr == nil {
			e := m.Up()
			log2.Println(e)
		} else {
			if !errors.Is(merr, postgres2.ErrNoSchema) {
				log2.Fatal(merr)
			}
		}
	}

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set("traceId", log.GetTraceId(c))
	})
	r.Use(log.GinLoggerMiddleware(&logger, log.GinLoggerMiddlewareParams{}))
	r.Use(sdk.UserMiddleware())
	r.Use(sdk.TraceMiddleware())
	r.Use(sdk.CorsMiddleware())
	r.Use(sdk.JsonMiddleware())
	r.Use(sdk.DbMiddleware(db))
	r.Use(sdk.AccountMiddleware(config.PublicRoutes))

	//if logger.Inner == false {
	//	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
	//		Formatter: func(param gin.LogFormatterParams) string {
	//			return fmt.Sprintf("%s - [%s] %s %s %s %d %s \"%s\" %s %s\n ",
	//				param.ClientIP,
	//				param.TimeStamp.Format(time.RFC1123),
	//				param.Method,
	//				param.Path,
	//				param.Request.Proto,
	//				param.StatusCode,
	//				param.Latency,
	//				param.Request.UserAgent(),
	//				param.ErrorMessage,
	//				param.Keys["traceId"],
	//			)
	//		},
	//		Output:    logger.Writer(),
	//		SkipPaths: []string{},
	//	}))
	//	r.Use(gin.Recovery())
	//}

	return &Application{
		Router: r,
		Db:     db,
		Logger: &logger,
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
	Value    interface{}
	Operator string
}
