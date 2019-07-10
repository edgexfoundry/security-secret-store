/*
   Copyright 2018 ForgeRock AS.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

  @author: Alain Pulluelo, ForgeRock (created: July 27, 2018)
  @author: Tingyu Zeng, DELL (updated: May 21, 2019)
  @version: 1.0.0
*/

package main

import (
	"flag"
	pki "github.com/edgexfoundry/security-secret-store/internal/pkg/pkisetup"
	"log"
	"os"
	"strconv"
)

func main() {

	var configFile string
	// Handling the command flags
	log.SetFlags(0)
	flag.StringVar(&configFile, "config", "", "use a JSON file as configuration: /path/to/file.json")
	flag.Parse()
	// Missing --config flag
	if configFile == "" {
		log.Println("ERROR: missing mandatory parameter: --config | -config")
		log.Println(pki.CmdUsageMsg)
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Read the Json config file and unmarshall content into struct type X509Config
	log.Printf("Config file      : %s \n", configFile)
	x509config, err := pki.ReadConfig(&configFile)
	if err != nil {
		pki.FatalIfErr(err, "Opening configuration file")
	}

	// Create and initialize the fs environment and global vars for the PKI materials
	cf, err := pki.CreateEnv(&x509config)
	if err != nil {
		pki.FatalIfErr(err, "Environment initialization")
	}

	newCA, err := strconv.ParseBool(x509config.CreateNewRootCA)

	// Optionaly generate the Root CA PKI materials (RSA or EC)
	if newCA {
		if _, _, err = pki.GenCA(&cf); err != nil {
			pki.FatalIfErr(err, "Root CA generation")
		}
	}

	// Generate the TLS server PKI materials (RSA or EC)
	if _, _, err = pki.GenCert(&cf); err != nil {
		pki.FatalIfErr(err, "TLS server generation")
	}
}
