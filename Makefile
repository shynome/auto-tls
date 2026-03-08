VERSION := $(shell git describe --tags --always --dirty)

build:
	CGO_ENABLED=0 go build -ldflags="-X 'main.Version=${VERSION}' -s -w" -o auto-tls .
docker: build
	docker build . -t shynome/auto-tls:${VERSION}
push: docker
	docker push shynome/auto-tls:${VERSION}
