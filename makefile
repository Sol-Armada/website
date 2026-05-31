.PHONY: clean build build-web build-server dev-web dev-server dev

version := $(shell git describe --tags --abbrev=0)
hash := $(shell git rev-parse --short HEAD)

clean:
	@rm -rf bin/

build-web-production:
	cd web && yarn build

build-web-beta:
	cd web && yarn build-beta

build-server: clean
	cd api && go build -ldflags "-X main.version=${version} -X main.hash=${hash}" -o ../bin/api ./cmd/server

build-production: build-web-production build-server

build-beta: build-web-beta build-server

build: build-production

dev-web:
	cd web && yarn dev

dev-server:
	cd api && go run ./cmd/server

dev: build-web-production
	@echo "Starting development environment..."
	@echo "Frontend: http://localhost:5173"
	@echo "Backend: http://localhost:8080"

