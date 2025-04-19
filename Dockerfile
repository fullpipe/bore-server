# syntax = docker.io/docker/dockerfile:experimental

# BUILDER
FROM golang as builder

WORKDIR /app

COPY --from=mwader/static-ffmpeg:5.1.2 /ffmpeg /usr/local/bin/

# BUILD
FROM builder as build

ENV CGO_ENABLED=0
ENV GOOS=linux

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

COPY --from=build /go/bin/app /go/bin/app
ENTRYPOINT ["/go/bin/app"]
