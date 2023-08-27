build:
	CGO_ENABLED=0 go build -o bin/app cmd/app/main.go

build-image: build
	docker build -t app:latest .

generate-data:
	GOGC=off go run cmd/util/util.go

generate-company-data:
	go run cmd/company-util/main.go

up: build generate-company-data
	docker-compose up --build -d

down:
	docker-compose down