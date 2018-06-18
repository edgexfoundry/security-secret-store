#!/bin/bash

docker build --no-cache=true --rm=true -t edgexfoundry/docker-edgex-vault -f ./Dockerfile.vault .

exit
#EOF
