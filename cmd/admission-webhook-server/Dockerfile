ARG GOLANG_VERSION=1.15.10
FROM golang:${GOLANG_VERSION) as builder

WORKDIR /usr/src/lokomotive

COPY . .

RUN make MOD=vendor build-webhook

# Admission webhook

FROM scratch

COPY --from=builder /usr/src/lokomotive/admission-webhook-server /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/admission-webhook-server"]
