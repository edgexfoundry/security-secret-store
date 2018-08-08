#!/bin/bash
#  ----------------------------------------------------------------------------------
#  vault-setup.sh    version 1.0 created July 18, 2018
#
#  @author:  Tony Espy, Canonical
#  @email:   espy@canonical.com
#
#  Copyright (c) 2018, Canonical Ltd
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

# Variables and parameters
echo ">> Setup _VAULT_DIR and fix permissions"
_VAULT_DIR=${_VAULT_DIR:-/vault}
_VAULT_SCRIPT_DIR=${_VAULT_SCRIPT_DIR:-/vault}
_PKI_SETUP_VAULT_ENV=${_PKI_SETUP_VAULT_ENV:-pki-setup-config-vault.env}

${_VAULT_SCRIPT_DIR}/pki-setup.sh ${_PKI_SETUP_VAULT_ENV}

# Don't chown in snap, as snaps don't support daemons using
# setuid/gid to drop from root to a specified user/group.
if [ -z "$SNAP" ]; then
    chown -R vault:vault ${_VAULT_DIR}
    chown -R vault:vault ${_VAULT_DIR}/pki
fi

chmod 750 ${_VAULT_DIR}/pki
chmod 640 ${_VAULT_DIR}/pki/*/*
