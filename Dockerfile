FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY *.go ./
RUN go build -o fetcher

FROM alpine:latest
WORKDIR /root/

RUN apk update && \
    apk add --no-cache ca-certificates && \
    update-ca-certificates

COPY --from=builder /app/fetcher /usr/local/bin/fetcher

RUN chmod +x /usr/local/bin/fetcher

CMD ["fetcher"]