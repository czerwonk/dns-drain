FROM golang:1.26.6-alpine3.24@sha256:3889b425f035be855a72fb4755265311293b6d414521f0a519d819df32222d83 AS builder
ADD . /go/dns-drain/
WORKDIR /go/dns-drain/cmd/dns-drainctl
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/dns-drainctl

FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
WORKDIR /app
COPY --from=builder /go/bin/dns-drainctl .

RUN adduser -S -G users dns-drain
USER dns-drain

ENTRYPOINT ["/app/dns-drainctl"]
