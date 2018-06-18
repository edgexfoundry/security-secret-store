#!/bin/bash
#  ----------------------------------------------------------------------------------
#  pki-setup.sh	version 1.0 created February 15, 2018
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
function usage() {
        echo "Usage  : pki-setup.sh </path/to/config-script>"
        echo "Example: pki-setup.sh pki-setup-config.env"
        echo " "
}

function initEnv() {
         # --------------------------------------------
         DEBUG=$1  # Debug switch: "debug-off" or "debug-on"
         # --------------------------------------------
         OPENSSL_BIN=$(which openssl)
         if [[ $? == 1 ]]; then
            echo "ERROR: openssl binary not found or not in the path"
            exit
         fi
         echo ">> openssl binary found: ${OPENSSL_BIN}"
         echo ">>" $(openssl version)
         echo ""
         # --------------------------------------------
         source $1  # </path/to/config-script>
         # --------------------------------------------
}

function initCA() {
         # ---------------------------------------------
         # CA (self signing) name, keystore alias and file (.pem .priv.key)
         CA_NAME=${ORG}$1
         # CA root certificate subject
         # CA_C
         # CA_ST
         # CA_L
         # CA_O
         CA_CN=${ORG}" "$1
         CA_HOME=${SETUP_PKI}/${CA_NAME}
         CERT_EXTENSIONS="ssl_server.ext"
         
         if [[ ${CREATE_CA} == "true" ]]; then
                # Create configuration directory (CA_HOME)
                echo ">> Creating a new Root CA Certificate"
                rm -rf ${CA_HOME}
                mkdir -p ${CA_HOME}
                chmod -R 755 ${SETUP_PKI}
                # Generate an initial serial number for signed/generated certificates
                echo `date +%s` > ${CA_HOME}/${CA_NAME}.srl
                # Create a certificate extensions file:
                #  The generated server certificate will be for TLS server usage only!
                cat > ${CA_HOME}/${CERT_EXTENSIONS} <<EOF
nsCertType=server
EOF
        else
                echo ">> Reusing existing Root CA Certificate"
                # Certificate serial number is incremented during CA signing process
        fi
         # ---------------------------------------------
}

function initVaultServer() {
         # --------------------------------------------
         # Vault FQDN, files (.pem .req .priv.key)
         if [[ ${DOMAIN} == "local" ]]; then
                FQDN=$1
         else
                FQDN=$1.${DOMAIN}
         fi
         # Vault TLS certificate subject
         # TLS_C
         # TLS_ST
         # TLS_L
         # TLS_O
         TLS_CN=${FQDN}
         # ---------------------------------------------
}

function genRootCA() {
        # Generate Root CA private key (RSA/4096) and trusted Root CA certificate (1825 days = 5 years)
        ${OPENSSL_BIN} req -x509 -sha256 -nodes -days 1825 -newkey rsa:4096 \
                   -subj "/C=${CA_C}/ST=${CA_ST}/L=${CA_L}/O=${CA_O}/CN=${CA_CN}/emailAddress=${CA_NAME}@${DOMAIN}"  \
                   -keyout ${CA_HOME}/${CA_NAME}.priv.key -out ${CA_HOME}/${CA_NAME}.pem

        # Generate Root CA private key (ECDSA/secp384r1) and trusted Root CA certificate
        #${OPENSSL_BIN} req -x509 -sha256 -nodes -days 1825 -newkey ec:secp384r1 \
        #           -subj "/C=${CA_C}/ST=${CA_ST}/L=${CA_L}/O=${CA_O}/CN=${CA_CN}/emailAddress=${CA_NAME}@${DOMAIN}"  \
        #           -keyout ${CA_HOME}/${CA_NAME}.priv.key -out ${CA_HOME}/${CA_NAME}.pem

        if [[ ${DEBUG} == "debug-on" ]]; then
                echo "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
                echo " CA Signing Certificate"
                echo "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
                ${OPENSSL_BIN} x509 -in ${CA_HOME}/${CA_NAME}.pem -noout -text
                echo "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
        fi
}

function genVaultCertReq() {
        # Generate TLS server private key (RSA/4096)
        ${OPENSSL_BIN} genrsa -out ${CA_HOME}/${FQDN}.priv.key 4096

        # Generate TLS server private key (ECDSA/secp384r1)
        #${OPENSSL_BIN} ecparam -genkey -name secp384r1 -out ${CA_HOME}/${FQDN}.priv.key

        # Generate server certificate signing request for TLS server certificate (CN must equal the FQDN)
        ${OPENSSL_BIN} req -sha256 -new -key ${CA_HOME}/${FQDN}.priv.key \
                   -out ${CA_HOME}/${FQDN}.req \
                   -subj "/C=${TLS_C}/ST=${TLS_ST}/L=${TLS_L}/O=${TLS_O}/CN=${TLS_CN}/emailAddress=admin@${DOMAIN}"

        if [[ ${DEBUG} == "debug-on" ]]; then
                echo "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
                echo " TLS Server Certificate Request for: ${FQDN}"
                echo "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
                ${OPENSSL_BIN} req -in ${CA_HOME}/${FQDN}.req -noout -text
                echo "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
        fi
}

function signVaultCertReq() {
        # CA signs TLS server certificate request with its trusted root certificate,
        # and creates the final TLS server certificate (1825 days = 5 years)
        ${OPENSSL_BIN} x509 -sha256 -req -in ${CA_HOME}/${FQDN}.req \
                    -CA ${CA_HOME}/${CA_NAME}.pem -CAkey ${CA_HOME}/${CA_NAME}.priv.key \
                    -CAserial ${CA_HOME}/${CA_NAME}.srl \
                    -extfile ${CA_HOME}/${CERT_EXTENSIONS} \
                    -days 1825 -outform PEM -out ${CA_HOME}/${FQDN}.pem

        if [[ ${DEBUG} == "debug-on" ]]; then
                echo "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
                echo " TLS Server Certificate for: ${FQDN}"
                echo "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
                ${OPENSSL_BIN} x509 -in ${CA_HOME}/${FQDN}.pem -noout -text
                echo "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
        fi
}

function storeToPKCS12() {
        # Create PKCS12 container for TLS server certificate and SK 
        # SK: Secret Key aka Private key
        # PK: Public Key
        #
        ${OPENSSL_BIN} pkcs12 -export -in ${CA_HOME}/${FQDN}.pem \
                      -inkey ${CA_HOME}/${FQDN}.priv.key -out ${CA_HOME}/${FQDN}.p12 \
                      -name ${FQDN} -password pass:${PKCS12_PASSWORD}
}

function houseKeeping() {
        chmod 644 ${SETUP_PKI}/${CA_NAME}/*
}
# ----------------- Common Functions ------------------------

#===================================== MAIN ====================================

# pki-setup.sh arguments check
if [[ $# != 1 ]]; then
        echo "SYNTAX ERROR"
        usage
        exit
fi

if [[ ! -f $1 ]]; then
        echo "FILE $1 NOT FOUND"
        exit
fi

# OpenSSL workout...
echo ">> Create binary paths and folder paths"
initEnv $1

echo ">> Prepare Root CA Certificate"
initCA TrustCA

echo ">> Prepare Vault Server TLS certificate"
initVaultServer ${HOST}

if [[ ${CREATE_CA} == "true" ]]; then
        echo ">> Generate Root CA private key and trusted root certificate"
        genRootCA debug-off
fi

echo ">> Generate TLS server private key"
echo ">> Generate server certificate signing request for TLS server certificate (CN must equal the FQDN)"
genVaultCertReq debug-off

echo ">> Root CA signs TLS server certificate request with its trusted root certificate,"
echo ">> and creates the final TLS server certificate"
signVaultCertReq debug-off

echo ">> Create PKCS12 containers for TLS server certificate and SK"
storeToPKCS12

houseKeeping
echo ">> PKI setup script completed."

exit

#EOF
