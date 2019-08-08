FROM golang:1.12-alpine

# Install build dependencies
RUN apk add git make

# Force the go compiler to use modules
ENV GO111MODULE=on

# Copy go.{mod,sum} to download and cache dependencies layer
COPY go.mod /usr/src/lokoctl/go.mod
COPY go.sum /usr/src/lokoctl/go.sum
COPY Makefile /usr/src/lokoctl/Makefile

WORKDIR /usr/src/lokoctl

# Only download dependencies
RUN go mod download

# Copy remaining source code and build
COPY . .
RUN make
