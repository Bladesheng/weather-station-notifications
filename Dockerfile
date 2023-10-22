FROM golang:1.21.3-bookworm AS builder
RUN apt-get update && apt-get upgrade -y

WORKDIR /app

# get dependencies first separately
COPY go.* .
RUN go mod download

# the rest of the source code needed for building
COPY . .

RUN go build
