.PHONY: clean build build-web build-server

version := $(shell git describe --tags --abbrev=0)
hash := $(shell git rev-parse --short HEAD)

clean:
	@rm -f bin/website

build-web:
	@yarn build

build-server: clean
	cd server && go build -ldflags "-X main.version=${version} -X main.hash=${hash}" -o ../bin/website ./

build: build-web build-server
	@true
