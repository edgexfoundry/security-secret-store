#!/bin/bash
#  ----------------------------------------------------------------------------------
#  consul-agent-config-cmd.sh version 1.0 created February 16, 2018
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

function usage() {
        echo "+++ Usage: $0 $@"
}

if [[ $# != 1 ]]; then
        echo "+++ SYNTAX ERROR"
        usage "<consul-server-container-id (s1|s2|s3)>"
        exit
fi

case $1 in
    s1|s2|s3|c1|c2) 
        true;;
    *)
        echo "+++ SYNTAX ERROR: Invalid Consul Agent Id (s1|s2|s3|c1|c2)"
        usage "<consul-server-container-id (s1|s2|s3|c1|c2)>"
        exit;;
esac

consul_host="consul-$1"

echo "+++ ${consul_container}: Consul Agent Configuration:"
echo ""
curl -v http://${consul_host}.edgex.net:8500/v1/agent/self | jq

exit

#EOF
