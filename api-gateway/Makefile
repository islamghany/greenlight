
run:
	@echo 'Running API Gateway'
	docker run --name greenlight-gateway -p 3050:80 greenlight-api-gateway:latest

build:
	@echo 'Building API Gateway image'
	docker image build --no-cache -t greenlight-api-gateway:latest .

rebuild:
	@echo 'Deleting greenlight-gateway container'
	docker container rm greenlight-gateway
	@echo 'Deleting API Gateway image'
	docker rmi greenlight-api-gateway:latest
	make build
	make run