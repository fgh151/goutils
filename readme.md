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

for migrations use [migrate](https://github.com/golang-migrate/migrate)

store migrations in `scripts/migrations` folder