.PHONY: clean build build-web build-server

version := $(shell git describe --tags --abbrev=0)
hash := $(shell git rev-parse --short HEAD)

clean:
	@rm -f bin/website

build-web-production:
	@yarn build

build-web-beta:
	@yarn build-beta

build-server: clean
	cd server && go build -ldflags "-X main.version=${version} -X main.hash=${hash}" -o ../bin/website ./

build-production: build-web-production build-server

build-beta: build-web-beta build-server

build: build-production
