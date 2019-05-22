/*******************************************************************************
 * Copyright 2018 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @author: Tingyu Zeng, Dell
 * @version: 1.0.0
 *******************************************************************************/
package vaultworker

import (
	"fmt"
	"os"
)

var usageStr = `
Usage: %s [options]
Server Options:
	--consul=true/false				Indicates if retrieving config from Consul
	--insureskipverify=true/false			Indicates if skipping the server side SSL cert verifcation, similar to -k of curl
	--init=true/false				Indicates if security service should be initialized	
	--configfile=<file.toml>			Use a different config file (default: res/configuration.toml)
	--wait=<time in seconds>		Indicates how long the program will pause between the vault initialization until it succeeds
	--debug=true/false				Output sensitive debug informations for security service
	Common Options:
	-h, --help					Show this message
`

// HelpCallback	print out the flag options for the server
func HelpCallback() {
	msg := fmt.Sprintf(usageStr, os.Args[0])
	fmt.Printf("%s\n", msg)
	os.Exit(0)
}
