build:
	CGO_ENABLED=0 go build -o bin/app cmd/app/main.go

build-image: build
	docker build -t app:latest .

generate-data:
	GOGC=off go run cmd/util/util.go

up: build-image
	docker-compose up -d

down:
	docker-compose down