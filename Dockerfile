# syntax = docker.io/docker/dockerfile:experimental

# BUILDER
FROM golang as builder

WORKDIR /app
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY --from=mwader/static-ffmpeg:5.1.2 /ffmpeg /usr/local/bin/
COPY --from=quay.io/goswagger/swagger /usr/bin/swagger /usr/bin/swagger
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# BUILD
FROM builder as build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# RUN make gen

RUN go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/app

# release
FROM golang:alpine as release

RUN apk update
RUN apk upgrade
RUN apk add --no-cache ffmpeg

# COPY --from=mwader/static-ffmpeg:5.1.2 /ffmpeg /usr/local/bin/
COPY --from=build /go/bin/app /go/bin/app
ENTRYPOINT ["/go/bin/app"]
