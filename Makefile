################################################################################
##                             VERSION PARAMS                                 ##
################################################################################

## Docker Build Versions
DOCKER_BUILD_IMAGE = golang:1.13
DOCKER_BASE_IMAGE = alpine:3.10

## Tool Versions
TERRAFORM_VERSION=0.11.14
KOPS_VERSION=1.15.0
HELM_VERSION=v2.14.2
KUBECTL_VERSION=v1.14.0

################################################################################

GO ?= $(shell command -v go 2> /dev/null)
GOPATH ?= $(shell go env GOPATH)
MATTERMOST_CLOUD_IMAGE ?= mattermost/mattermost-cloud:test
MACHINE = $(shell uname -m)
GOFLAGS ?= $(GOFLAGS:)
BUILD_TIME := $(shell date -u +%Y%m%d.%H%M%S)
BUILD_HASH := $(shell git rev-parse HEAD)
AWS_SRC_PATH := $(GOPATH)/src/github.com/aws/aws-sdk-go/service

export GO111MODULE=on

## Checks the code style, tests, builds and bundles.
all: check-style dist

## Runs govet and gofmt against all packages.
.PHONY: check-style
check-style: govet lint
	@echo Checking for style guide compliance

## Runs lint against all packages.
.PHONY: lint
lint:
	@echo Running lint
	env GO111MODULE=off $(GO) get -u golang.org/x/lint/golint
	golint -set_exit_status ./...
	@echo lint success

## Runs govet against all packages.
.PHONY: vet
govet:
	@echo Running govet
	$(GO) vet ./...
	@echo Govet success

## Builds and thats all :)
.PHONY: dist
dist:	build


# Generate mocks from the interfaces.
.PHONY: mocks
mocks:
	@env GO111MODULE=off $(GO) get -u github.com/vektra/mockery/.../
	
	# AWS
	@env GO111MODULE=off $(GO) get -u github.com/aws/aws-sdk-go/...
	# $(GOPATH)/bin/mockery -dir $(AWS_SRC_PATH)/ec2/ec2iface -all -output ./internal/tools/aws/mocks
	# $(GOPATH)/bin/mockery -dir $(AWS_SRC_PATH)/rds/rdsiface -all -output ./internal/tools/aws/mocks
	# $(GOPATH)/bin/mockery -dir $(AWS_SRC_PATH)/s3/s3iface -all -output ./internal/tools/aws/mocks
	# $(GOPATH)/bin/mockery -dir $(AWS_SRC_PATH)/acm/acmiface -all -output ./internal/tools/aws/mocks
	# $(GOPATH)/bin/mockery -dir $(AWS_SRC_PATH)/iam/iamiface -all -output ./internal/tools/aws/mocks
	# $(GOPATH)/bin/mockery -dir $(AWS_SRC_PATH)/route53/route53iface -all -output ./internal/tools/aws/mocks
	$(GOPATH)/bin/mockery -dir $(AWS_SRC_PATH)/secretsmanager/secretsmanageriface -all -output ./internal/tools/aws/mocks
	
	# INTERNAL
	$(GOPATH)/bin/mockery -dir ./internal/provisioner -all -output ./internal/provisioner/mocks
	$(GOPATH)/bin/mockery -dir ./internal/supervisor -all -output ./internal/supervisor/mocks
	$(GOPATH)/bin/mockery -dir ./internal/testlib -all -output ./internal/testlib/mocks
	$(GOPATH)/bin/mockery -dir ./internal/webhook -all -output ./internal/webhook/mocks
	$(GOPATH)/bin/mockery -dir ./internal/store -all -output ./internal/store/mocks
	$(GOPATH)/bin/mockery -dir ./internal/api -all -output ./internal/api/mocks
	
	# MODEL
	$(GOPATH)/bin/mockery -dir ./model -all -output ./model/mocks

.PHONY: build
build: ## Build the mattermost-cloud
	@echo Building Mattermost-Cloud
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) -a -installsuffix cgo -o build/_output/bin/cloud  ./cmd/cloud

build-image:  ## Build the docker image for mattermost-cloud
	@echo Building Mattermost-cloud Docker Image
	docker build \
	--build-arg DOCKER_BUILD_IMAGE=$(DOCKER_BUILD_IMAGE) \
	--build-arg DOCKER_BASE_IMAGE=$(DOCKER_BASE_IMAGE) \
	. -f build/Dockerfile -t $(MATTERMOST_CLOUD_IMAGE) \
	--no-cache

get-terraform: ## Download terraform only if it's not available. Used in the docker build
	@if [ ! -f build/terraform ]; then \
		curl -Lo build/terraform.zip https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && cd build && unzip terraform.zip &&\
		chmod +x terraform && rm terraform.zip;\
	fi

get-kops: ## Download kops only if it's not available. Used in the docker build
	@if [ ! -f build/kops ]; then \
		curl -Lo build/kops https://github.com/kubernetes/kops/releases/download/${KOPS_VERSION}/kops-linux-amd64 &&\
		chmod +x build/kops;\
	fi

get-helm: ## Download helm only if it's not available. Used in the docker build
	@if [ ! -f build/helm ]; then \
		curl -Lo build/helm.tar.gz https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz &&\
		cd build && tar -zxvf helm.tar.gz &&\
		cp linux-amd64/helm . && chmod +x helm && rm helm.tar.gz && rm -rf linux-amd64;\
	fi

get-kubectl: ## Download kubectl only if it's not available. Used in the docker build
	@if [ ! -f build/kubectl ]; then \
		curl -Lo build/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl &&\
		chmod +x build/kubectl;\
	fi


.PHONY: install
install: build
	go install ./...
