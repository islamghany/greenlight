

.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./db/migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:  
	@echo 'Running up migrations...'
	migrate -path ./db/migrations -database postgresql://root:secret@localhost:5431/logs?sslmode=disable up

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/clean
db/migrations/clean: 
	@echo 'Fix and force version...'
	migrate -path ./db/migrations -database postgresql://root:secret@localhost:5431/logs?sslmode=disable force ${version}

## proto/generate generate the types from proto file
.PHONY: proto/generate
proto/generate:
	@echo 'Running user proto file'
	protoc --go_out=. --go_opt=paths=source_relative \
  	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
 	userspb/user.proto