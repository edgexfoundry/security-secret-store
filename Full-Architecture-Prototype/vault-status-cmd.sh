#!/bin/bash
#  ----------------------------------------------------------------------------------
#  vault-status-cmd.sh version 1.0 created February 12, 2018
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

source vault-cmd-global.env

for vault_container in myvault-s1 myvault-s2; do
        echo "+++ ${vault_container} status:"
	echo ""
	docker exec -it ${vault_container} vault status ${_TLS}
	echo ""
done

exit

#EOF
