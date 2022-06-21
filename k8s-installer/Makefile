GOPATH ?= $(shell go env GOPATH)
IMG_PREFIX ?= k8s-installer
TAG ?= latest
REPO_URL ?= $(shell git remote get-url origin)
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
COMMIT_REF ?= $(shell git rev-parse --verify HEAD)
LOW_VERSION_GIT_REPO_URL ?= $(shell git ls-remote --get-url origin)
GOOS ?= $(shell go env GOHOSTOS)
GOARCH ?= $(shell go env GOARCH)
ORG ?= caas4
TESTHOST ?= '172.20.149.55:8099'

apps = 'server' 'client'

VERSION ?= $(shell git rev-parse --short HEAD)-$(shell date -u '+%Y%m%d%I%M%S')
BUILDX ?= false

.PHONY: build
build:
	gofmt -w test pkg cmd tools internal
	for app in $(apps) ;\
	do \
		CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./cmd/$$app -a -ldflags "-w -s -X k8s-installer/pkg/server/version.Version=${VERSION}" ./cmd/$$app;\
	done

.PHONY: swagger
swagger:
	go run tools/doc-gen/main.go --output=swagger-ui/swagger.json

.PHONY: swagger-server
swagger-server:
	go run swagger-ui/swagger.go

.PHONY: image-server
image-server:
	docker build --build-arg VERSION=$(VERSION) --build-arg REPO_URL=$(REPO_URL) --build-arg BRANCH=$(BRANCH) --build-arg COMMIT_REF=$(COMMIT_REF) -f build/server/Dockerfile -t $(ORG)/$(IMG_PREFIX)-server:latest .

.PHONY: image-client
image-client:
	docker build --build-arg VERSION=$(VERSION) --build-arg REPO_URL=$(REPO_URL) --build-arg BRANCH=$(BRANCH) --build-arg COMMIT_REF=$(COMMIT_REF) -f build/client/Dockerfile -t $(ORG)/$(IMG_PREFIX)-client:latest .

.PHONY: image-swagger-deps
image-swagger-deps:
	sudo docker run -d --rm --name k8s-installer-server-swagger caas4/k8s-installer-server sleep 1000
	sudo docker cp k8s-installer-server-swagger:/app/swagger.json swagger-ui/
	sudo docker exec k8s-installer-server-swagger mkdir -p /app/hack/deps
	sudo docker exec k8s-installer-server-swagger /app/server genDeps
	sudo docker cp k8s-installer-server-swagger:/app/hack/deps/k8s-installer-deps.json hack/deps/
	sudo docker kill k8s-installer-server-swagger

.PHONY: image
image: image-server image-client image-swagger-deps

.PHONY: push-image
push-image:
	sudo docker push $(ORG)/$(IMG_PREFIX)-client:latest
	sudo docker push $(ORG)/$(IMG_PREFIX)-server:latest

.PHONY: api-test
api-test:
	echo $(TESTHOST) > /tmp/caas4_host
	CGO_ENABLED=1 go test ./test/cmd/server

.PHONY: test
test:
	mkdir -p ~/.k8s-installer/
	cp etc/* ~/.k8s-installer/
	echo $(TESTHOST) > /tmp/caas4_host
	CGO_ENABLED=1 go test ./...

.PHONY: deps
deps: build
	cmd/server/server genDeps

.PHONY: all
all: build swagger deps
