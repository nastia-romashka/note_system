FROM golang:1.25 AS builder

WORKDIR /src/app

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY app/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/search_service ./cmd/main

FROM alpine:3.22

WORKDIR /service

COPY --from=builder /out/search_service ./search_service
COPY config.yml ./config.yml

EXPOSE 8086

CMD ["./search_service"]
