

.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./db/migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: 
	@echo 'Running up migrations...'
	migrate -path ./db/migrations -database postgresql://root:secret@localhost:5431/logs?sslmode=disable up

