FROM golang:1.27.2-alpine3.24@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS builder
ADD . /go/dns-drain/
WORKDIR /go/dns-drain/cmd/dns-drainctl
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/dns-drainctl

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
WORKDIR /app
COPY --from=builder /go/bin/dns-drainctl .

RUN adduser -S -G users dns-drain
USER dns-drain

ENTRYPOINT ["/app/dns-drainctl"]
