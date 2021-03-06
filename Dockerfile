FROM golang:1.18rc1-bullseye

WORKDIR /usr/src/app

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
    file \
 && apt-get -y clean \
 && rm -rf /var/lib/apt/lists/*

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum /usr/src/app
RUN go mod download && go mod verify

COPY . .
