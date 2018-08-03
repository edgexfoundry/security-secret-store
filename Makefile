#  SPDX-License-Identifier: Apache-2.0'

.PHONY: build clean run

DOCKERS=docker_vault docker_vault_worker
.PHONY: $(DOCKERS)

VERSION=$(shell cat ./VERSION)

GIT_SHA=$(shell git rev-parse HEAD)

clean:
	rm -f $(MICROSERVICES)

build: $(DOCKERS)

docker_vault:
	docker build \
    --no-cache=true --rm=true \
		-f Dockerfile.vault \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-edgex-vault:$(GIT_SHA) \
		-t edgexfoundry/docker-edgex-vault:$(VERSION)-dev \
		.

docker_vault_worker:
	docker build \
    --no-cache=true --rm=true \
		-f Dockerfile.vault-worker \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-edgex-vault-worker:$(GIT_SHA) \
		-t edgexfoundry/docker-edgex-vault-worker:$(VERSION)-dev \
		.

