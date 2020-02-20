FROM golang

COPY . /usr/src/lokoctl

WORKDIR /usr/src/lokoctl

RUN make MOD=vendor install-slim
