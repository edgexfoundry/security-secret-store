#!/bin/bash
#  ----------------------------------------------------------------------------------
#  vault-login-cmd.sh version 1.0 created February 13, 2018
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

# ATTENTION: a vault login is only valid when processed against a vault leader, not a standby.

source vault-cmd-global.env

if [[ $# != 1 ]]; then
        echo "+++ SYNTAX ERROR"
        usage "<vault-server-container-id (s1|s2)>"
        exit
fi

if [[ ($1 != "s1") && ($1 != "s2") ]]; then
        echo "+++ SYNTAX ERROR: Invalid Vault Server Id (s1|s2)"
        usage "<vault-server-container-id (s1|s2)>"
        exit
fi

vault_container="myvault-$1"

# Login to Vault master (not possible on Vault slave i.e. standby)
echo "+++ ${vault_container} login"
echo ""
docker exec -it ${vault_container} vault login ${_TLS} $_ROOT_TOKEN 
echo ""

exit

#EOF
