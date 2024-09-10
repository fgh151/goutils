### Install
```shell
go get github.com/runetid/go-sdk
```

### Config

```go
type ApplicationConfig struct {
	PublicRoutes     []string // Public routes
	DbMigrationsPath string // path to migrations on github 
	DbSchema         string 
}
```

### Migrations

For migrations use [migrate](https://github.com/golang-migrate/migrate)

Store migrations in `scripts/migrations` folder

Add certs to container, example:

```dockerfile
FROM debian:buster-slim
RUN apt-get update && apt-get install -y ca-certificates openssl
ARG cert_location=/usr/local/share/ca-certificates
# Get certificate from "github.com"
RUN openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > ${cert_location}/github.crt
# Update certificates
RUN update-ca-certificates

COPY --from=builder /app/main /main
ENTRYPOINT ["/main", "-v"]
```

### Context variables
- ```traceId``` - уникальный идентификатор трассировки в формате UUID ```string```
- ```databaseConn``` - экземпляр подключения к базе данных ```*gorm.DB```
- ```event_id``` - идентификатор мероприятия ```int64```
- ```role``` - роль пользователя ```string```
- ```token``` - токен пользователя ```string```
- ```user``` - экземпляр пользователя ```*models.User```
- ```event``` - экземпляр мероприятия ```*models.Event```

### Available env variables

- ```ENVIRONMENT``` - Окружение запуска ```DEV|PROD|TEST```
- ```HTTP_ADDR``` - Порт запуска HTTP
- ```GH_LOGIN``` - Логин пользователя GitHub для запуска миграций
- ```DB_HOST``` - IP адрес базы данных
- ```DB_USER``` - Пользователь базы данных
- ```DB_PASSWORD``` - Пароль базы данных
- ```DB_NAME``` - Имя базы данных
- ```DB_PORT``` - Порт базы данных
- ```DNS_ACCOUNT``` - DNS адрес микросервиса аккаунтов
- ```DNS_USERS``` - DNS адрес микросервиса пользователей
- ```DNS_EVENT``` - DNS адрес микросервиса мероприятий