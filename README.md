# s3_filer_go

# Development

```
$ docker-compose run --rm app
Creating s3_filer_go_app_run ... done
root@36e0ffe445ad:/usr/src/app# make build
export GO111MODULE=on
env GOOS=linux go build -ldflags="-s -w" -o bin/s3_filer s3_filer/*.go
```
