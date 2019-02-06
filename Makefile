#  SPDX-License-Identifier: Apache-2.0'

.PHONY: prepare build clean docker run

GO=CGO_ENABLED=0 GOOS=linux go
DOCKERS=docker_vault docker_vault_worker
PKISETUP=pkisetup
VAULTWORKER=edgex-vault-worker
.PHONY: $(DOCKERS)

VERSION=$(shell cat ./VERSION)
GIT_SHA=$(shell git rev-parse HEAD)

prepare:
	glide install
	
clean:
	cd core && rm -f $(VAULTWORKER)
	cd pkisetup && rm -f $(PKISETUP)

build: build_pki build_worker

build_pki:
	cd pkisetup && $(GO) build -a -ldflags="-s -w" -o $(PKISETUP) .

build_worker:
	cd core && $(GO) build -a -o $(VAULTWORKER) .

docker: $(DOCKERS)

docker_vault: build_pki
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

