#  SPDX-License-Identifier: Apache-2.0'

.PHONY: prepare build clean docker run test

GO=CGO_ENABLED=0 GO111MODULE=on GOOS=linux go
DOCKERS=docker_pki_init docker_vault docker_vault_worker
PKIINIT=pki-init
PKISETUP=pkisetup
VAULTWORKER=edgex-vault-worker
.PHONY: $(DOCKERS)

VERSION=$(shell cat ./VERSION)
GIT_SHA=$(shell git rev-parse HEAD)

prepare:
	
clean:
	cd cmd/vaultworker && rm -f $(VAULTWORKER)
	cd cmd/pkisetup && rm -f $(PKISETUP)
	cd cmd/pkiinit && rm -f $(PKIINIT)

build: build_pki_init build_pki_setup build_worker

build_pki_init:
	cd cmd/pkiinit && $(GO) build -a -o $(PKIINIT) .

build_pki_setup:
	cd cmd/pkisetup && $(GO) build -a -ldflags="-s -w" -o $(PKISETUP) .

build_worker:
	cd cmd/vaultworker && $(GO) build -a -o $(VAULTWORKER) .

docker: $(DOCKERS)

docker_pki_init: build_pki_init build_pki_setup
	docker build \
        --no-cache=true --rm=true \
		-f Dockerfile.pki-init \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-edgex-pki-init:$(GIT_SHA) \
		-t edgexfoundry/docker-edgex-pki-init:$(VERSION)-dev \
		-t edgexfoundry/docker-edgex-pki-init:latest \
		.

docker_vault: build_pki_setup
	docker build \
        --no-cache=true --rm=true \
		-f Dockerfile.vault \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-edgex-vault:$(GIT_SHA) \
		-t edgexfoundry/docker-edgex-vault:$(VERSION)-dev \
		-t edgexfoundry/docker-edgex-vault:latest \
		.

docker_vault_worker: build_worker
	docker build \
        --no-cache=true --rm=true \
		-f Dockerfile.vault-worker \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-edgex-vault-worker-go:$(GIT_SHA) \
		-t edgexfoundry/docker-edgex-vault-worker-go:$(VERSION)-dev \
		-t edgexfoundry/docker-edgex-vault-worker-go:latest \
		.

test: 
		$(GO) test ./... -cover
		$(GO) vet ./...
		
# address this later in a separate issue and PR
# gofmt -l . 
# [ "`gofmt -l .`" = "" ]