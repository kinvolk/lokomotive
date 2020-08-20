FROM golang as builder

WORKDIR /usr/src/lokomotive

COPY . .

RUN make MOD=vendor build-webhook

# Admission webhook

FROM scratch

COPY --from=builder /usr/src/lokomotive/admission-webhook-server /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/admission-webhook-server"]
