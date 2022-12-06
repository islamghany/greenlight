FROM  golang:1.18-alpine as builder

WORKDIR /app

COPY . .


RUN CGO_ENABLED=0 go build -o main main.go

# RUN chmod +x /app/mailer



FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

CMD ["/app/main"]