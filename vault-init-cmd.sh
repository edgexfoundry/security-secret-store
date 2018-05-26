#!/bin/bash
#  ----------------------------------------------------------------------------------
#  vault-init-cmd.sh version 1.0 created February 13, 2018
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

# IMPORTANT: only init the master (leader), the slave (standby) only needs unsealing with resulting leader's keys

source vault-cmd-global.env

rm -f ${_CURRENT_INIT_ASSETS} ${_TMP}

vault_container="myvault-s1"
echo "+++ ${vault_container} init process..."
docker exec -it ${vault_container} vault operator init -format=json ${_TLS} > ${_TMP}

# Removing ANSI color codes
cat ${_TMP} | perl -pe 's/\x1b\[[0-9;]*[mG]//g' > ${_CURRENT_INIT_ASSETS}

chmod 600 ${_CURRENT_INIT_ASSETS}
cat ${_CURRENT_INIT_ASSETS} | jq

rm -f ${_TMP}

exit

#EOF
