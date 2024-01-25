gen:
	go generate ./...

build:
	docker build --platform=linux/amd64 --target release -t registry.gitlab.com/fullpipe/registry/bore-server .

push:
	docker push registry.gitlab.com/fullpipe/registry/bore-server:latest
