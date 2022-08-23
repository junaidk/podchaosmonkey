DOCKER_REGISTRY := podchaosmonkey
CMD      := podchaosmonkey

# Docker Tag from Git
DOCKER_IMAGE_TAG ?= ${GIT_TAG}
ifeq ("${DOCKER_IMAGE_TAG}","")
	DOCKER_IMAGE_TAG = "1.0"
endif

BIN_OUTDIR ?= ./build/bin
BIN_NAME   ?= podchaosmonkey

# Go Build Flags
GOBUILDFLAGS :=
GOBUILDFLAGS += -o ${BIN_OUTDIR}/${BIN_NAME}

.PHONY: build
build: CMD = main.go
build:
	go build ${GOBUILDFLAGS} ${CMD}

DOCKER_ARGS:=
#DOCKER_ARGS+= --force-rm
DOCKER_ARGS+= -f ./Dockerfile
DOCKER_ARGS+= -t ${DOCKER_REGISTRY}/${CMD}:${DOCKER_IMAGE_TAG}


# Build docker image
.PHONY: image
image:
	docker build ${DOCKER_ARGS} .

# deploys to configured kubernetes instance
.PHONY: deploy
deploy:
	kubectl delete -f k8s/ 2>/dev/null; true
	kubectl create -f k8s/

# deploys test workload to configured kubernetes instance
.PHONY: deploy-workload
deploy-workload:
	kubectl delete -f k8s/test-workload 2>/dev/null; true
	kubectl delete ns workloads 2>/dev/null; true
	kubectl create ns workloads
	kubectl create -f k8s/test-workload

.PHONY: test
test:
	go test ./...
