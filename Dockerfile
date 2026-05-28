# syntax=docker/dockerfile:1

FROM golang:1.25 AS builder

WORKDIR /src

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/flysoft-flight-service ./cmd/app

FROM alpine:3.22

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder /out/flysoft-flight-service /usr/local/bin/flysoft-flight-service

USER app

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/flysoft-flight-service"]
