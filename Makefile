
REGISTRIES?=registry.qtt6.cn
REPOSITORY?=paas-dev
APP=storesd
V=$(shell cat VERSION)


build:
	GOOS=linux GOARCH=amd64 go build -o deploy/bin/storesd main.go config.go


image: build
	@docker build -f deploy/Dockerfile deploy -t $(REGISTRIES)/$(REPOSITORY)/$(APP):$(V)
	@docker push $(REGISTRIES)/$(REPOSITORY)/$(APP):$(V)
	@echo "$(REGISTRIES)/$(REPOSITORY)/$(APP):$(V)"
