#!/bin/bash
#  ----------------------------------------------------------------------------------
#  vault-switch-login-cmd.sh version 1.0 created February 13, 2018
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

# REMARK: a vault login is only valid when processed against a vault leader, not a standby.

source vault-cmd-global.env

# Step down of Vault leader (S1), becoming standby, this will elect the standby Vault (S2) as new leader
vault_container=myvault-s1
echo "+++ ${vault_container} step-down"
docker exec -it ${vault_container} vault operator step-down ${_TLS}
echo ""

sleep 6
status myvault-s1
status myvault-s2
echo ""

# Login to new Vault leader (not possible on Vault slave i.e. standby)
vault_container=myvault-s2
echo "+++ ${vault_container} login"
docker exec -it ${vault_container} vault login ${_TLS} ${_ROOT_TOKEN}
echo ""

# Step down of Vault leader (S2), becoming standby, this will elect the standby Vault (S1) as new leader
vault_container=myvault-s2
echo "+++ ${vault_container} step-down"
docker exec -it ${vault_container} vault operator step-down ${_TLS}
echo ""

sleep 6
status myvault-s1
status myvault-s2

exit

#EOF
