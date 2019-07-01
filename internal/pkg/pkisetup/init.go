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

package pkisetup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/* CertConfig holds information required to create PKI environment */
type CertConfig struct {
	// Main Setup Configuration
	configFile string
	pkiCaDir   string
	dumpConfig bool // Dump the JSON config to console
	newCA      bool

	// Root CA Certificate
	caName     string
	caKeyFile  string
	caCertFile string
	caCountry  string
	caState    string
	caLocality string
	caOrg      string

	// TLS Server Certificate
	tlsHost     string
	tlsDomain   string
	tlsFQDN     string
	tlsAltFQDN  string
	tlsKeyFile  string
	tlsCertFile string
	tlsCountry  string
	tlsState    string
	tlsLocality string
	tlsOrg      string

	// Key Generation
	dumpKeys   bool // Dump the keys to console: debug only!
	rsaScheme  bool
	rsaKeySize int
	ecScheme   bool
	ecCurve    string
}

/* CreateEnv creates enviroment for the PKI certs */
func CreateEnv(x509config *X509Config) (CertConfig, error) {

	cf := CertConfig{}

	// Abs returns an absolute representation of path.
	// If the path is not absolute it will be joined with the current working directory
	// to turn it into an absolute path.
	wDir, err := filepath.Abs(x509config.WorkingDir)
	if err != nil {
		FatalIfErr(err, "ERROR: failed to build the working directory absolute path from JSON config")
	}
	log.Printf("Working directory: %s", wDir)

	// pkiCaDir: Concatenate working dir absolute path with PKI setup dir, using separator "/"
	cf.pkiCaDir = strings.Join([]string{wDir, x509config.PKISetupDir, x509config.RootCA.CAName}, "/")

	// Convert create_new_ca JSON string "true|false" to boolean
	cf.newCA, err = strconv.ParseBool(x509config.CreateNewRootCA)
	// Convert dump_config JSON string "true|false" to boolean
	cf.dumpConfig, err = strconv.ParseBool(x509config.DumpConfig)
	// Convert dump_keys JSON string "true|flase| to boolean
	cf.dumpKeys, err = strconv.ParseBool(x509config.KeyScheme.DumpKeys)
	// Convert rsa JSON string "true|false" to boolean
	cf.rsaScheme, err = strconv.ParseBool(x509config.KeyScheme.RSA)
	// Convert rsa_key_size JSON string to integer
	cf.rsaKeySize, err = strconv.Atoi(x509config.KeyScheme.RSAKeySize)
	// Convert ec JSON string "true|false" to boolean
	cf.ecScheme, err = strconv.ParseBool(x509config.KeyScheme.EC)
	// EC chosen curve
	cf.ecCurve = x509config.KeyScheme.ECCurve
	// Init: CA name and PEM key/cert filenames
	cf.caName = x509config.RootCA.CAName
	cf.caKeyFile = filepath.Join(cf.pkiCaDir, cf.caName+skFileExt)
	cf.caCertFile = filepath.Join(cf.pkiCaDir, cf.caName+certFileExt)
	// Init: TLS host.domain and PEM key/cert filenames
	cf.tlsHost = x509config.TLSServer.TLSHost
	cf.tlsDomain = x509config.TLSServer.TLSDomain
	if cf.tlsDomain == "local" {
		cf.tlsFQDN = cf.tlsHost
		cf.tlsAltFQDN = cf.tlsHost + "." + cf.tlsDomain
	} else {
		cf.tlsFQDN = cf.tlsHost + "." + cf.tlsDomain
		cf.tlsAltFQDN = ""
	}
	cf.tlsKeyFile = filepath.Join(cf.pkiCaDir, cf.tlsHost+skFileExt)
	cf.tlsCertFile = filepath.Join(cf.pkiCaDir, cf.tlsHost+certFileExt)
	// CA subjects
	cf.caCountry = x509config.RootCA.CACountry
	cf.caState = x509config.RootCA.CAState
	cf.caLocality = x509config.RootCA.CALocality
	cf.caOrg = x509config.RootCA.CAOrg
	// TLS subjects
	cf.tlsCountry = x509config.TLSServer.TLSCountry
	cf.tlsState = x509config.TLSServer.TLSSate
	cf.tlsLocality = x509config.TLSServer.TLSLocality
	cf.tlsOrg = x509config.TLSServer.TLSOrg

	// Print the JSON parameters to console
	if cf.dumpConfig {
		// inside createEnv, local x509config is already an address
		// as it was passed to createEnv w/ the & prefix
		_ = dumpJSONConfig(x509config)
	}

	// Creating a new fresh PKI setup dir, if new CA is requested
	if cf.newCA {
		// Remove eventual previous PKI setup directory
		// Create a new empty PKI setup directory
		log.Println("New CA creation requested by configuration")
		log.Println("Cleaning up CA PKI setup directory")

		err = os.RemoveAll(cf.pkiCaDir) // Remove pkiCaDir
		if err != nil {
			return cf, fmt.Errorf("Failed to remove the existing CA PKI configuration directory: %s (%s)", cf.pkiCaDir, err)
		}

		log.Printf("Creating CA PKI setup directory: %s", cf.pkiCaDir)
		err = os.MkdirAll(cf.pkiCaDir, 0750) // Create pkiCaDir
		if err != nil {
			return cf, fmt.Errorf("Failed to create the CA PKI configuration directory: %s (%s)", cf.pkiCaDir, err)
		}
	} else { // Using an existing PKI setup directory, if new CA is *NOT* requested
		log.Println("No new CA creation requested by configuration")

		// Is the CA there? (if nil then OK... but could be something else than a directory)
		stat, err := os.Stat(cf.pkiCaDir)
		if err != nil {
			if os.IsNotExist(err) {
				return cf, fmt.Errorf("CA PKI setup directory does not exist: %s", cf.pkiCaDir)
			}
			return cf, fmt.Errorf("CA PKI setup directory cannot be reached: %s (%s)", cf.pkiCaDir, err)
		}
		if stat.IsDir() {
			log.Printf("Existing CA PKI setup directory: %s", cf.pkiCaDir)
		} else {
			return cf, fmt.Errorf("Existing CA PKI setup directory is not a directory: %s", cf.pkiCaDir)
		}
	}
	return cf, nil
}
