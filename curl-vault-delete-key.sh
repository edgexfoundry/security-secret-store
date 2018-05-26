#!/bin/bash
#  ----------------------------------------------------------------------------------
#  curl-vault-delete-key.sh version 1.0 created February 14, 2018
#
#  @author:  Alain Pulluelo, ForgeRock
#  @email:   alain.pulluelo@forgerock.com
#  @address: 201 Mission St, Suite 2900
#            San Francisco, CA 94105, USA
#  @phone:   +1(415)-559-1100
#
#  Copyright (c) 2018, ForgeRock
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#  ----------------------------------------------------------------------------------

# https://www.vaultproject.io/api/secret/kv/

source curl-vault-global.env

if [[ $# != 2 ]]; then
        echo "+++ SYNTAX ERROR"
        usage "<vault-server-id (s1|s2)> <key-name>"
        exit
fi

if [[ ($1 != "s1") && ($1 != "s2") ]]; then
        echo "+++ SYNTAX ERROR: Invalid Vault Server Id (s1|s2)"
        usage "<vault-server-id (s1|s2)> <key-name>"
        exit
fi

setServerPort $1
_KEY=$2

echo ">> Delete a secret: ${_KEY} @${_SERVER}:${_PORT}"

curl -sw 'HTTP STATUS: %{http_code}\n' ${_TLS} ${_REDIRECT} \
  --header "X-Vault-Token: ${_ROOT_TOKEN}" \
  --request DELETE \
  ${_HTTP_SCHEME}://${_SERVER}:${_PORT}/v1/secret/${_KEY} 

exit

#EOF
