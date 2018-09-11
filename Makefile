#  SPDX-License-Identifier: Apache-2.0'

.PHONY: build clean docker run

GO=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go
DOCKERS=docker_vault docker_vault_worker
PKISETUP=pkisetup
VAULTWORKER=edgex-vault-worker
.PHONY: $(DOCKERS)

VERSION=$(shell cat ./VERSION)
GIT_SHA=$(shell git rev-parse HEAD)

clean:
	cd core && rm -f $(VAULTWORKER)
	cd pkisetup.src && rm -f $(PKISETUP)

build:
	cd pkisetup.src && $(GO) build -a -ldflags="-s -w" -o $(PKISETUP) .
	cd core && $(GO) build -a -o $(VAULTWORKER) .

docker: $(DOCKERS)

docker_vault: build
	docker build \
        --no-cache=true --rm=true \
		-f Dockerfile.vault \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-edgex-vault:$(GIT_SHA) \
		-t edgexfoundry/docker-edgex-vault:$(VERSION)-dev \
		-t edgexfoundry/docker-edgex-vault:latest \
		.

docker_vault_worker: 
	docker build \
        --no-cache=true --rm=true \
		-f Dockerfile.vault-worker \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-edgex-vault-worker:$(GIT_SHA) \
		-t edgexfoundry/docker-edgex-vault-worker:$(VERSION)-dev \
		-t edgexfoundry/docker-edgex-vault-worker:latest \
		.

