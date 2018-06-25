#!/bin/bash
#  ----------------------------------------------------------------------------------
#  _exec_bash.sh version 1.0 created February 16, 2018
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

if [[ $# -lt 1 ]]; then
  echo 'Missing parameters: container'
  echo 'Usage is : _exec_bash.sh container'
  exit
fi
#
# If Vault is configured with TLS then vault commands need to verify the TLS server certificate,
# which was signed by a Root CA (self signed included), use VAULT_CAPATH to refer the Root CA 
# certificate (in PEM format).
#
set -v
docker exec -it -e PS1='\u@\h:\w \$ ' -e VAULT_CAPATH='/vault/pki/EdgeXTrustCA.pem' $1 sh
set +v

exit
#EOF
