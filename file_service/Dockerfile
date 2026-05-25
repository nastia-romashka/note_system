FROM golang:1.25 AS builder

WORKDIR /src/app

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY app/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/file_service ./cmd/main

FROM alpine:3.22

WORKDIR /service

COPY --from=builder /out/file_service ./file_service
COPY config.yml ./config.yml

EXPOSE 8085

CMD ["./file_service"]
