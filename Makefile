.PHONY: frontend backend build docker run tidy migrate deploy release

frontend:
	cd web/manager && yarn install && yarn build

backend:
	go build -o bin/portal ./cmd/portal
	go build -o bin/ws ./cmd/ws
	go build -o bin/cron ./cmd/cron

migrate: backend
	./bin/portal migrate

build: frontend backend

deploy:
	./deploy.sh

release: build
	./deploy.sh

tidy:
	go mod tidy

run:
	./bin/portal

docker:
	docker build -t portal-go:latest .

docker-run:
	docker run --rm -p 3000:3000 -p 9999:9999 -v $(PWD)/config:/app/config -v $(PWD)/upload:/app/upload portal-go:latest
