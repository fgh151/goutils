package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/runetid/go-sdk"
	"github.com/runetid/go-sdk/crud"
	_ "github.com/runetid/go-sdk/example/docs"
	"gorm.io/gorm"
	"log"
	"net/http"
)

// ApiAccount model info
// @Description User account information
// @Description with user id and username
type ApiAccount struct {
	ID            int    `gorm:"primaryKey;column:id" json:"id"`
	Key           string `gorm:"column:key;index" json:"key"`
	Secret        string `gorm:"column:secret" json:"secret"`
	EventID       int    `gorm:"column:event_id" json:"event_id"`
	Role          string `gorm:"column:role" json:"role"`
	Blocked       bool   `gorm:"column:blocked;index" json:"blocked"`
	BlockedReason string `gorm:"column:blocked_reason" json:"blocked_reason"`
	Comment       string `gorm:"column:comment" json:"comment"`

	Domains []AccountDomain `gorm:"foreignKey:account_id"`

	crud.BaseCrudModel `swaggerignore:"true"`
}

func (u ApiAccount) TableName() string {
	return "api_account"
}

func (u ApiAccount) List(db *gorm.DB, request crud.ListRequest, params ...crud.FilterParams) (interface{}, int64, error) {
	var models []ApiAccount
	err := db.Debug().Limit(request.Limit).Offset(request.Offset).Find(&models).Error
	var count int64
	db.Model(ApiAccount{}).Count(&count)
	return models, count, err
}

func (u ApiAccount) Create(db *gorm.DB) (interface{}, error) {

	err := db.Debug().Create(&u).Error
	return u, err
}

func (u ApiAccount) Update(db *gorm.DB) (interface{}, error) {

	err := db.Debug().Save(&u).Error
	return u, err
}

func (u ApiAccount) DecodeCreate(c *gin.Context) (interface{}, error) {
	err := c.Bind(&u)
	return u, err
}

func (u ApiAccount) Delete(db *gorm.DB, key string) bool {
	return db.Debug().Delete(&ApiAccount{}, key).RowsAffected > 0
}

func (u ApiAccount) Get(db *gorm.DB, key string) (interface{}, error) {
	var model ApiAccount
	tx := db.Debug().Where("id = ?", key).First(&model)
	err := tx.Error

	if tx.RowsAffected < 1 {
		err = errors.New("not found")
	}

	return model, err
}

func (u ApiAccount) CheckHash(hash string, time string) bool {
	data := u.Key + time + u.Secret
	h := md5.Sum([]byte(data))
	result := hex.EncodeToString(h[:16])
	return hash == result
}

func (u ApiAccount) CheckIp(db *gorm.DB, ip string) bool {
	var exists bool
	err := db.Debug().Model(AccountDomain{}).
		Select("count(*) > 0").
		Where("domain = ? AND account_id = ?", ip, u.ID).
		Find(&exists).Error

	return exists && err == nil
}

type AccountDomain struct {
	ID        int    `gorm:"primaryKey;column:id" json:"id"`
	AccountId int    `gorm:"column:account_id;index" json:"account_id"`
	Domain    string `gorm:"column:domain;index" json:"domain"`
	Comment   string `gorm:"column:comment" json:"comment"`
}

func (u *AccountDomain) TableName() string {
	return "api_account_domain"
}

func (u AccountDomain) List(db *gorm.DB, request crud.ListRequest, params ...crud.FilterParams) (interface{}, int64, error) {
	var models []ApiAccount

	query := db.Debug().Limit(request.Limit).Offset(request.Offset)

	for _, param := range params {
		query.Where(param.Key+" "+param.Operator+" ?", param.Value)
	}

	err := query.Find(&models).Error
	var count int64
	db.Model(ApiAccount{}).Count(&count)
	return models, count, err
}

func (u AccountDomain) GetFilterParams(c *gin.Context) []crud.FilterParams {
	var p []crud.FilterParams
	p = append(p, crud.FilterParams{
		Key:      "account_id",
		Value:    c.Param("id"),
		Operator: "=",
	})
	return p
}

func (u AccountDomain) Create(db *gorm.DB) (interface{}, error) {
	err := db.Debug().Create(&u).Error
	return u, err
}

func (u AccountDomain) Update(db *gorm.DB) (interface{}, error) {
	err := db.Debug().Save(&u).Error
	return u, err
}

func (u AccountDomain) DecodeCreate(c *gin.Context) (interface{}, error) {
	err := c.Bind(&u)
	u.AccountId, err = sdk.String2Int(c.Param("id"))
	return u, err
}

func (u AccountDomain) Delete(db *gorm.DB, key string) bool {
	return db.Debug().Delete(&AccountDomain{}, key).RowsAffected > 0
}

func (u AccountDomain) Get(db *gorm.DB, key string) (interface{}, error) {
	var model AccountDomain
	tx := db.Debug().Where("id = ?", key).First(&model)
	err := tx.Error

	if tx.RowsAffected < 1 {
		err = errors.New("not found")
	}

	return model, err
}

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {

	app, err := crud.NewCrudApplication([]string{"/apiaccount/check", "/apiaccount/swagger/*"})

	if err != nil {
		log.Panic("Cant init app")
	}

	app.Db.AutoMigrate(ApiAccount{}, AccountDomain{})

	app.AppendListEndpoint("/apiaccount", ApiAccount{})
	app.AppendCreateEndpoint("/apiaccount", ApiAccount{})
	app.AppendDeleteEndpoint("/apiaccount/:id", ApiAccount{})
	app.AppendGetEndpoint("/apiaccount/:id", ApiAccount{})
	app.AppendUpdateEndpoint("/apiaccount/:id", ApiAccount{})

	app.AppendListEndpoint("/apiaccount/:id/domain", AccountDomain{})
	app.AppendDeleteEndpoint("/apiaccount/domain/:id", AccountDomain{})
	app.AppendCreateEndpoint("/apiaccount/:id/domain", AccountDomain{})
	app.AppendUpdateEndpoint("/apiaccount/:id/domain", AccountDomain{})

	app.AppendSwagger("/apiaccount/")

	app.Router.GET("/apiaccount/check", func(c *gin.Context) {
		key := c.Request.Header.Get("ApiKey")
		hash := c.Request.Header.Get("Hash")
		time := c.Request.Header.Get("Time")

		var account ApiAccount

		tx := app.Db.Debug().Where("key = ? AND blocked = ?", key, false).Find(&account)

		if tx.Error != nil || tx.RowsAffected < 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"Message": "Unauthorized"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}

		if account.CheckHash(hash, time) {
			c.Set("apiAccount", &account)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"Message": "Wrong api key or hash or timestamp"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}

		ip := c.Request.Header.Get("X-Real-IP")

		if ip == "" {
			ip = c.RemoteIP()
		}

		if !account.CheckIp(app.Db, ip) {
			c.JSON(http.StatusUnauthorized, gin.H{"Message": "Wrong server ip"})
			c.Writer.WriteHeaderNow()
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, gin.H{"Message": "ok"})
		c.Writer.WriteHeaderNow()
		c.Abort()
		return
	})

	app.Run()
}

// list godoc
// @Summary      List accounts
// @Description  List accounts
// @Tags         Api
// @Accept       json
// @Produce      json
// @Param        limit   path      int  true  "records limit"
// @Param        offset   path      int  true  "records offset"
// @Success      200  {array}  ApiAccount
// @Router       /apiaccount/list [get]
func list() {

}
