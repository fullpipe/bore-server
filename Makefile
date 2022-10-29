gen:
	go generate ./...

build:
	docker build --target release -t rg.fr-par.scw.cloud/???/bore-server .

push:
	docker push rg.fr-par.scw.cloud/???/bore-server:latest
