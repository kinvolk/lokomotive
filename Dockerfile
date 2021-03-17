ARG GOLANG_VERSION=1.15.10
FROM golang:${GOLANG_VERSION}

COPY . /usr/src/lokomotive

WORKDIR /usr/src/lokomotive

RUN make MOD=vendor install-slim
