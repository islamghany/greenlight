FROM golang:1.18.4-alpine3.16 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o main main.go



FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

COPY app.env .

COPY db/migrations ./db/migrations


CMD ["/app/main"]