FROM golang:1.27.1-alpine3.24@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS builder
ADD . /go/dns-drain/
WORKDIR /go/dns-drain/cmd/dns-drainctl
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/dns-drainctl

FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
WORKDIR /app
COPY --from=builder /go/bin/dns-drainctl .

RUN adduser -S -G users dns-drain
USER dns-drain

ENTRYPOINT ["/app/dns-drainctl"]
