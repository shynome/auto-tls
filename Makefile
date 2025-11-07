build:
	CGO_ENABLED=0 go build -ldflags="-X 'main.Version=$$(git describe --tags --always --dirty)' -s -w" -o auto-tls .
docker: build
	docker build . -t shynome/auto-tls:$$(git describe --tags --always --dirty)
push: docker
	docker push shynome/auto-tls:$$(git describe --tags --always --dirty)
