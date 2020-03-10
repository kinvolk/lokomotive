FROM golang

COPY . /usr/src/lokomotive

WORKDIR /usr/src/lokomotive

RUN make MOD=vendor install-slim
