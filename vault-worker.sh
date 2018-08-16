#!/bin/bash
#  ----------------------------------------------------------------------------------
#  vault-worker.sh	version 1.0 created June 14, 2018
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

# Variables and parameters
_VAULT_SCRIPT_DIR=${_VAULT_SCRIPT_DIR:-/vault}

while true
do
   # Init/Unseal processes
   ${_VAULT_SCRIPT_DIR}/vault-init-unseal.sh

   # If Vault init/unseal was OK... eventually prepare materials for Kong
   if [[ $? == 0 ]]; then
       ${_VAULT_SCRIPT_DIR}/vault-kong.sh
   fi

   sleep ${WATCHDOG_DELAY}
done

exit

#EOF
