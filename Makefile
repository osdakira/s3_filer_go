.PHONY: build clean

build:
	export GO111MODULE=on
	env GOOS=darwin go build -ldflags="-s -w" -o bin/s3_filer s3_filer/*.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock
