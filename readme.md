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