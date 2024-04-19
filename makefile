clean:
	@rm -f bin/website

build-web:
	@yarn build

build-server:
	@go build -o bin/website ./server

build: clean build-web build-server
	@true
