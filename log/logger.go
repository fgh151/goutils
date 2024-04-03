package log

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"log"
	"time"
)

func GetTraceId(c *gin.Context) string {
	request := c.Request.Header.Get("X-Trace-Id")

	if request == uuid.New().String() {

	}

	return request
}

func NewAppLogger() AppLogger {

	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	log.SetOutput(logger.Writer())

	return AppLogger{logger: logger}

}

type AppLogger struct {
	logger  *logrus.Logger
	TraceId string
}

func (l AppLogger) Info(v ...any) {
	l.logger.Info(v...)
}

func (l AppLogger) Warn(v ...any) {
	l.logger.Warn(v...)
}

func (l AppLogger) Error(v ...any) {
	l.logger.Error(v...)
}

func (l AppLogger) WithContext(ctx context.Context) *logrus.Entry {

	return l.logger.WithFields(logrus.Fields{
		"traceId": ctx.Value("traceId"),
	})
}

func (l AppLogger) GetLogger() *logrus.Logger {
	return l.logger
}

// GinLoggerMiddlewareParams defines the configuration options for the logger middleware
type GinLoggerMiddlewareParams struct {
	// a list of paths to skip logging for
	SkipPaths []string
}

// GinLoggerMiddleware returns a middleware handler function that can be used with the gin
// router for logging requests.
func GinLoggerMiddleware(appLogger *AppLogger, params GinLoggerMiddlewareParams) gin.HandlerFunc {
	// build a map of the skipped paths
	var skipMap map[string]struct{}

	if length := len(params.SkipPaths); length > 0 {
		skipMap = make(map[string]struct{}, length)

		for _, path := range params.SkipPaths {
			skipMap[path] = struct{}{}
		}
	}

	// return the handler function
	return func(c *gin.Context) {

		// we use the match path because it will match the value
		// defined on the router
		matchPath := c.FullPath()

		if _, ok := skipMap[matchPath]; !ok {
			// get the context of the request
			//ctx := c.Request.Context()

			// get basic information about the request
			//requestID := c.GetHeader("X-Request-ID")
			start := time.Now().UTC()
			path := c.Request.URL.Path
			//hostname, err := os.Hostname()

			//if err != nil {
			//	hostname = "unknown"
			//}

			// builld the request appLogger
			//appLogger := logrus.WithContext(ctx).WithFields(logrus.Fields{
			//	"request_id": requestID,
			//	"hostname":   hostname,
			//})

			// process request
			c.Next()

			// response metrics
			timestamp := time.Now().UTC()
			latency := timestamp.Sub(start)
			statusCode := c.Writer.Status()
			//dataLength := c.Writer.Size()

			//appLogger = appLogger.WithFields(logrus.Fields{
			//	"status_code": statusCode,
			//	"latency":     latency.Milliseconds(),
			//	"data_length": dataLength,
			//})
			//
			//if len(c.Errors) > 0 {
			//	appLogger = appLogger.WithField("errors", c.Errors)
			//}

			msg := fmt.Sprintf("[%s] %d %s (%dms)", timestamp.Format(time.RFC3339), statusCode, path, latency.Milliseconds())

			if statusCode >= 500 {
				appLogger.logger.WithFields(logrus.Fields{
					"method":  c.Request.Method,
					"path":    c.Request.URL.Path,
					"query":   c.Request.URL.Query(),
					"traceId": c.Value("traceId"),
				}).Error(msg)
			} else if statusCode >= 400 {
				appLogger.logger.WithFields(logrus.Fields{
					"method":  c.Request.Method,
					"path":    c.Request.URL.Path,
					"query":   c.Request.URL.Query(),
					"traceId": c.Value("traceId"),
				}).Warn(msg)
			} else {
				appLogger.logger.WithFields(logrus.Fields{
					"method":  c.Request.Method,
					"path":    c.Request.URL.Path,
					"query":   c.Request.URL.Query(),
					"traceId": c.Value("traceId"),
				}).Info(msg)
			}
		}
	}
}

type GormLogger struct {
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
	Debug                 bool
	logger                *logrus.Logger
}

func NewGormLogger(logger *AppLogger) *GormLogger {
	return &GormLogger{
		SkipErrRecordNotFound: true,
		Debug:                 true,
		logger:                logger.GetLogger(),
	}
}

func (l *GormLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, s string, args ...interface{}) {

	l.logger.Info(s)

	l.logger.Infof(s, args)
}

func (l *GormLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	l.logger.Warn(s)
	l.logger.Warnf(s, args)
}

func (l *GormLogger) Error(ctx context.Context, s string, args ...interface{}) {
	l.logger.Error(s)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := logrus.Fields{}
	if l.SourceField != "" {
		fields[l.SourceField] = utils.FileWithLineNum()
	}

	fields["sql"] = sql
	fields["traceId"] = ctx.Value("traceId")

	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		l.logger.WithField("test", "1").Error(sql)
		//fields[l.logger.ErrorKey] = err
		l.logger.WithFields(fields).Errorf("%s [%s]", sql, elapsed)
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		l.logger.WithField("test", "2").Warn(sql)
		l.logger.WithFields(fields).Warnf("%s [%s]", sql, elapsed)
		return
	}

	if l.Debug {

		//entry := logrus.NewEntry(l.logger)
		//entry.WithField("test", "3")
		//entry.Info(sql)

		//l.logger.Infof(sql)

		l.logger.WithFields(fields).Info(sql)

		l.logger.WithFields(fields).Debugf("%s [%s]", sql, elapsed)
	}
}
