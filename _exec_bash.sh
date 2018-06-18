#!/bin/bash
#  ----------------------------------------------------------------------------------
#  @edgex/developer-scripts
#  _exec_bash.sh	version 1.0   created May 24, 2018
#
#  Alain Pulluelo, VP Security & Mobile Innovation (alain.pulluelo@forgerock.com)
#
#  ForgeRock Office of the CTO
#
#  201 Mission St, Suite 2900
#  San Francisco, CA 94105, USA
#  +1(415)-559-1100
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
_container=$1
set -v
docker exec -it -e PS1='\u@\h:\w \$ ' -e VAULT_CAPATH='/vault/pki/EdgeXFoundryTrustCA/EdgeXFoundryTrustCA.pem' ${_container} sh
#docker exec -it -e PS1='\u@\h:\w \$ ' ${_container} sh
set +v

exit
#EOF
