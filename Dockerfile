FROM golang:1.27.1-alpine3.24@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS builder
ADD . /go/dns-drain/
WORKDIR /go/dns-drain/cmd/dns-drainctl
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/dns-drainctl

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
WORKDIR /app
COPY --from=builder /go/bin/dns-drainctl .

RUN adduser -S -G users dns-drain
USER dns-drain

ENTRYPOINT ["/app/dns-drainctl"]
