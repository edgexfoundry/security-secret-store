#!/bin/bash
#  ----------------------------------------------------------------------------------
#  vault-init-unseal.sh	version 1.0 created June 14, 2018
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
    rm -f ${_PAYLOAD_INIT}
    rm -f ${_PAYLOAD_UNSEAL}
}


function vaulInitialization() {
    echo ">> (1) Vault Initialization Process"

    # Create the Vault init request payload with only 1 key and obviously threshold 1
    cat > ${_PAYLOAD_INIT} <<EOF
    {
    "secret_shares": 1,
    "secret_threshold": 1
    }
EOF
    # ---------------------------------------------------------------------------
    # Vault Initialization API
    # ---------------------------------------------------------------------------
    curl -sw 'HTTP-STATUS: %{http_code}\n' ${_TLS} ${_REDIRECT} \
        --request PUT \
        --data @${_PAYLOAD_INIT} \
        ${_HTTP_SCHEME}://${_VAULT_SVC}:${_VAULT_PORT}/v1/sys/init > ${_TMP}
    # ---------------------------------------------------------------------------

    # Check http status code returned by init request
    result=$(tail -1 ${_TMP} | grep "HTTP-STATUS:" | cut -d' ' -f2)

    case ${result} in
        # If Vault initialization OK
        #
        # Example:
        # {"keys":["8e70bcf6ba046b59857cba1ec6495c58b53f6c60e37f871e53b4b391ff43ec59"],"keys_base64":["jnC89roEa1mFfLoexklcWLU/bGDjf4ceU7Szkf9D7Fk="],"root_token":"6e2e099f-e5a4-028b-6b84-f11d1fb1ad9d"}
        # HTTP-STATUS: 200
        "200")
            # let's grab the init key
            _INIT_KEY=$(head -1 ${_TMP} | jq -r '.keys | .[0]')
            # let's grab the root token
            _ROOT_TOKEN=$(head -1 ${_TMP} | jq -r '.root_token')
            # save the key and the root token JSON (strip HTTP-STATUS)
            head -1 ${_TMP} | jq '.' > ${_RESP_INIT}
            chown vault:vault ${_RESP_INIT}
            echo ">> Vault successfully initialized"
        ;;
        # If Vault already initialized
        #
        # Example:
        # {"errors":["Vault is already initialized"]}
        # HTTP STATUS: 400
        "400")
            # let's grab the error message in case...
            result=$(head -1 ${_TMP} | jq -r '.errors | .[0]')
            echo ">> ${result}"
            # let's go unseal... but before we need:
            # 1) Check previous init response JSON still exists
            if [[ -f ${_RESP_INIT} ]]; then
                # 2) Grab the init key from previous init
                _INIT_KEY=$(cat ${_RESP_INIT} | jq -r '.keys | .[0]')
                # 3) Grab the root token from previous init
                _ROOT_TOKEN=$(cat ${_RESP_INIT} | jq -r '.root_token')
            else
                echo ">> Vault initialization error!"
                echo ">> Previous init response file not found: ${_RESP_INIT}"
                _EXIT="1"
            fi
        ;;
        #
        # Unattended error...
        #
        *)
            echo ">> Vault initialization unattended error"
            _EXIT="1"
        ;;
    esac

    # Handle the error exit use case for Vault init process
    if [[ ${_EXIT} == "1" ]]; then
        echo "==> Vault init request response:"
        cat ${_TMP}
        echo ">>"
    fi

    return ${_EXIT}
}


function vaultUnsealing() {
    echo ">> (2) Vault Unseal Process"

    # https://www.vaultproject.io/api/system/seal-status.html

    # Create the Vault unseal request payload with the unseal key
    cat > ${_PAYLOAD_UNSEAL} <<EOF
    {
    "key": "${_INIT_KEY}"
    }
EOF
    # ---------------------------------------------------------------------------
    # Vault Unsealing API
    # ---------------------------------------------------------------------------
    curl -sw 'HTTP-STATUS: %{http_code}\n' ${_TLS} ${_REDIRECT} \
        --request PUT \
        --data @${_PAYLOAD_UNSEAL} \
        ${_HTTP_SCHEME}://${_VAULT_SVC}:${_VAULT_PORT}/v1/sys/unseal > ${_TMP}
    # ---------------------------------------------------------------------------

    # Check http status code returned by unseal request
    result=$(tail -1 ${_TMP} | grep "HTTP-STATUS:" | cut -d' ' -f2)

    # If Vault unsealing OK
    #
    # Example:
    # {"type":"shamir","sealed":false,"t":1,"n":1,"progress":0,"nonce":"","version":"0.10.2","cluster_name":"vault-cluster-1df0f671","cluster_id":"10f1a1eb-ad7a-511f-06af-aab9c370e412"}
    # HTTP-STATUS: 200
    #
    # Remark: unsealing Vault when already unsealed generates same output response with code 200

    if [[ ${result} == "200" ]]; then
        # let's grab the sealed state boolean
        result=$(head -1 ${_TMP} | jq -r '.sealed')
        if [[ ${result} == "false" ]]; then
            # save the unseal JSON response (strip HTTP-STATUS)
            head -1 ${_TMP} | jq '.' > ${_RESP_UNSEAL}
            chown vault:vault ${_RESP_UNSEAL}
            echo ">> Vault successfully unsealed"
        else
            echo ">> Vault unseal ok but incoherent sealed status!"
            _EXIT="1"
        fi
    else
        echo ">> Vault unseal error!"
        _EXIT="1"
    fi

    # Handle the error exit use case for Vault unseal process
    if [[ ${_EXIT} == "1" ]]; then
        echo "==> Vault unseal request response:"
        cat ${_TMP}
        echo ">>"
    fi

    return ${_EXIT}
}


function vaultRegistered() {
    sleep 3 # Allow Consul DNS table refresh
    echo ">> Check Vault is registered as a service in Consul"
    echo ">> DNS requests to Consul service on port 8600/tcp"
    echo -n "vault.service.consul: "
    dig +short +tcp -p8600 @edgex-core-consul vault.service.consul
    echo -n "active.vault.service.consul: "
    dig +short +tcp -p8600 @edgex-core-consul  active.vault.service.consul
    echo ""

    return ${_EXIT}
}
# ----------------- Common Functions ------------------------


#===================================== MAIN INIT ===============================

# Variables and parameters
_VAULT_DIR="/vault"
_VAULT_CONFIG_DIR="${_VAULT_DIR}/config"
_VAULT_PKI_DIR="${_VAULT_DIR}/pki"
_VAULT_FILE_DIR="${_VAULT_DIR}/file"

_PAYLOAD_INIT="${_VAULT_FILE_DIR}/payload-init.json"
_PAYLOAD_UNSEAL="${_VAULT_FILE_DIR}/payload-unseal.json"
_RESP_INIT="${_VAULT_FILE_DIR}/resp-init.json"
_RESP_UNSEAL="${_VAULT_FILE_DIR}/resp-unseal.json"
_VAULT_CONFIG="${_VAULT_CONFIG_DIR}/local.json"
_TMP="${_VAULT_FILE_DIR}/_tmp.vault"
_EXIT="0"

_CA="EdgeXFoundryTrustCA"
_CA_DIR="${_VAULT_PKI_DIR}/${_CA}"
_CA_PEM="${_CA_DIR}/${_CA}.pem"
_TLS=" --cacert ${_CA_PEM}"
_REDIRECT=" --location" # If HTTP temporary redirect (HTTP STATUS 307) follow it
_HTTP_SCHEME="https"
_VAULT_SVC="edgex-vault"
_EDGEX_DOMAIN=""
_VAULT_PORT="8200"

#===================================== MAIN PROC ===============================

# 0) Cleanup up previous temp files and payloads
houseKeeping

# 1) Init Vault
vaulInitialization

# Handle the error exit use case for Vault init process
if [[ $? == "0" ]]; then
    # 2) Unseal Vault
    vaultUnsealing 
    # Handle the error exit use case for Vault unseal process
    if [[ $? == "0" ]]; then
        # 4) Check Vault was successfully registered as a Consul service
        vaultRegistered 
    fi
fi

# 5) Cleanup up temp files and payloads
houseKeeping 

exit ${_EXIT}
#EOF