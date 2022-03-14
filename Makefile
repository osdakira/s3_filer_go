.PHONY: build clean

build:
	go fmt s3_filer/*.go
	export GO111MODULE=on
	env GOOS=darwin go build -ldflags="-s -w" -o bin/s3_filer s3_filer/*.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

run:
	docker-compose run --rm app make build
	bin/s3_filer
