#!/bin/bash
#  ----------------------------------------------------------------------------------
#  vault-kong.sh	version 1.0 created June 15, 2018
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

#
# ----------------- Common Functions ------------------------
#
function houseKeeping() {
    rm -f ${_TMP}
    rm -f ${_PAYLOAD_KONG}
}
# ----------------- Common Functions ------------------------

#===================================== MAIN ====================================

# Variables and parameters
_VAULT_DIR="/vault"
_VAULT_CONFIG_DIR="${_VAULT_DIR}/config"
_VAULT_PKI_DIR="${_VAULT_DIR}/pki"
_VAULT_FILE_DIR="${_VAULT_DIR}/file"

_PAYLOAD_INIT="${_VAULT_FILE_DIR}/payload-init.json"
_PAYLOAD_UNSEAL="${_VAULT_FILE_DIR}/payload-unseal.json"
_PAYLOAD_KONG="${_VAULT_FILE_DIR}/payload-kong.json"
_RESP_INIT="${_VAULT_FILE_DIR}/resp-init.json"
_RESP_UNSEAL="${_VAULT_FILE_DIR}/resp-unseal.json"
_VAULT_CONFIG="${_VAULT_CONFIG_DIR}/local.json"
_TMP="${_VAULT_FILE_DIR}/_tmp.vault"
_EXIT="0"

_CA="EdgeXFoundryCA"
_CA_DIR="${_VAULT_PKI_DIR}/${_CA}"
_CA_PEM="${_CA_DIR}/${_CA}.pem"
_TLS=" --cacert ${_CA_PEM}"

_KONG_SVC="edgex-kong"
_KONG_PEM="${_CA_DIR}/${_KONG_SVC}.pem"
_KONG_SK="${_CA_DIR}/${_KONG_SVC}.priv.key"
_REDIRECT=" --location" # If HTTP temporary redirect (HTTP STATUS 307) follow it
_HTTP_SCHEME="https"
_VAULT_SVC="edgex-vault"
_EDGEX_DOMAIN=""
_VAULT_PORT="8200"
_VAULT_API_PATH_KONG="/v1/secret/edgex/pki/tls/${_KONG_SVC}"

houseKeeping # temp files and payloads

# Generate Kong PKI/TLS materials if they haven't been already...
if [[ (! -f ${_KONG_PEM}) || (! -f ${_KONG_SK}) ]]; then
    echo ">> (3) Create PKI materials for Kong TLS server certificate"
    /vault/pkisetup --config /vault/pkisetup-kong.json
    chown vault:vault ${_CA_DIR}/${_KONG_SVC}.*
else
    echo ">> (3) PKI materials for Kong TLS server certificate already created"
    openssl x509 -noout -subject -in ${_KONG_PEM}
    openssl x509 -noout -issuer -in ${_KONG_PEM}
fi

echo ""    
echo ">> (4) Fetch the Vault Root Token"
_ROOT_TOKEN=$(cat ${_RESP_INIT} | jq -r '.root_token')

echo ">> (5) Test if the Kong Key/Value already exists"
curl -sw 'HTTP-STATUS: %{http_code}\n' ${_TLS} ${_REDIRECT} \
    --header "X-Vault-Token: ${_ROOT_TOKEN}" \
    --request GET \
    ${_HTTP_SCHEME}://${_VAULT_SVC}:${_VAULT_PORT}${_VAULT_API_PATH_KONG} > ${_TMP}

# Check http status code returned by get request
result=$(tail -1 ${_TMP} | grep "HTTP-STATUS:" | cut -d' ' -f2)

case ${result} in
    # HTTP-STATUS: 200 -> key found
    "200")
        echo "==> Key/Value already in Vault, done!"
    ;;
    # HTTP-STATUS: 404 -> Key not found
    "404")
        echo ">> (6) Create the Kong JSON with TLS certificate and private key (base64 encoded)"
        jq -n --arg cert "$(cat ${_KONG_PEM}|base64)" \
            --arg sk "$(cat ${_KONG_SK}|base64)" \
            '{cert:$cert,sk:$sk}' > ${_PAYLOAD_KONG}

        echo ">> (7) Load the Kong JSON PKI materials in Vault"
        curl -sw 'HTTP-STATUS: %{http_code}\n' ${_TLS} ${_REDIRECT} \
            --header "X-Vault-Token: ${_ROOT_TOKEN}" \
            --header "Content-Type: application/json" \
            --request POST \
            --data @${_PAYLOAD_KONG} \
            ${_HTTP_SCHEME}://${_VAULT_SVC}:${_VAULT_PORT}${_VAULT_API_PATH_KONG} > ${_TMP}

        # Check http status code returned by post request
        result=$(tail -1 ${_TMP} | grep "HTTP-STATUS:" | cut -d' ' -f2)

        if [[ ${result} == "204" ]]; then
            echo "==> Key/Value successfully written in Vault, done!"
        else
            echo "==> Error while writing Key/Value in Vault!"
            _EXIT="1"
        fi
    ;;
    *)
        echo "==> Unattended Error while reading Key/Value"
        _EXIT="1"
    ;;
esac

# Handle the error exit use case
if [[ ${_EXIT} == "1" ]]; then
    echo "==> Vault request response:"
    cat ${_TMP}
    echo ">>"
fi

echo ""

houseKeeping # temp files and payloads

exit ${_EXIT}

#EOF
