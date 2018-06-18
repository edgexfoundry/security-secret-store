#!/bin/bash

docker build --no-cache=true --rm=true -t edgexfoundry/docker-edgex-vault-init-unseal -f ./Dockerfile.vault-init-unseal .

exit
#EOF
