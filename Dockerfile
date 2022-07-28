#Build Stage
FROM golang:1.18.4-alpine3.16 AS builder
WORKDIR /app
COPY . .
RUN go build -o main ./cmd/api/
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.12.2/migrate.linux-amd64.tar.gz | tar xvz

#Run Stage
FROM alpine
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate.linux-amd64 ./migrate
COPY .envrc .
COPY start.sh .
COPY wait-for.sh .
COPY migrations ./migrations

EXPOSE 4000
CMD ["/app/main"]
ENTRYPOINT [ "/app/start.sh" ]