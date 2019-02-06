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
  @version: 1.0.0
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// KeyScheme parameters (RSA vs EC)
// RSA: 1024, 2048, 4096
// EC: 224, 256, 384, 521
type KeyScheme struct {
	DumpKeys   string `json:"dump_keys"`
	RSA        string `json:"rsa"`
	RSAKeySize string `json:"rsa_key_size"`
	EC         string `json:"ec"`
	ECCurve    string `json:"ec_curve"`
}

// RootCA parameters from JSON: x509_root_ca_parameters
type RootCA struct {
	CAName     string `json:"ca_name"`
	CACountry  string `json:"ca_c"`
	CAState    string `json:"ca_st"`
	CALocality string `json:"ca_l"`
	CAOrg      string `json:"ca_o"`
}

// TLSServer parameters from JSON config: x509_tls_server_parameters
type TLSServer struct {
	TLSHost     string `json:"tls_host"`
	TLSDomain   string `json:"tls_domain"`
	TLSCountry  string `json:"tls_c"`
	TLSSate     string `json:"tls_st"`
	TLSLocality string `json:"tls_l"`
	TLSOrg      string `json:"tls_o"`
}

// X509Config JSON config file main structure
type X509Config struct {
	CreateNewRootCA string    `json:"create_new_rootca"`
	WorkingDir      string    `json:"working_dir"`
	PKISetupDir     string    `json:"pki_setup_dir"`
	DumpConfig      string    `json:"dump_config"`
	KeyScheme       KeyScheme `json:"key_scheme"`
	RootCA          RootCA    `json:"x509_root_ca_parameters"`
	TLSServer       TLSServer `json:"x509_tls_server_parameters"`
}

const cmdUsageMsg = "Usage of ./pkisetup:"
const skFileExt = ".priv.key"
const certFileExt = ".pem"

var (
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
)

// ----------------------------------------------------------
func main() {

	// Handling the command flags
	log.SetFlags(0)
	flag.StringVar(&configFile, "config", "", "use a JSON file as configuration: /path/to/file.json")
	flag.Parse()
	// Missing --config flag
	if configFile == "" {
		log.Println("ERROR: missing mandatory parameter: --config | -config")
		log.Println(cmdUsageMsg)
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Read the Json config file and unmarshall content into struct type X509Config
	log.Printf("Config file      : %s \n", configFile)
	x509config, err := readConfig(&configFile)
	if err != nil {
		fatalIfErr(err, "Opening configuration file")
	}

	// Create and initialize the fs environment and global vars for the PKI materials
	err = createEnv(&x509config)
	if err != nil {
		fatalIfErr(err, "Environment initialization")
	}

	// Optionaly generate the Root CA PKI materials (RSA or EC)
	if newCA {
		// caCert, caSK, err := genCA()
		if _, _, err = genCA(); err != nil {
			fatalIfErr(err, "Root CA generation")
		}
	}

	// Generate the TLS server PKI materials (RSA or EC)
	// tlsCert, tlsSK, err := genCert()
	if _, _, err = genCert(); err != nil {
		fatalIfErr(err, "TLS server generation")
	}
}

// ----------------------------------------------------------
func readConfig(configFilePtr *string) (X509Config, error) {

	var jsonX509Config X509Config

	// Open JSON config file
	jsonFile, err := os.Open(*configFilePtr)
	if err != nil {
		return jsonX509Config, err
	}

	// Read JSON config file into byteArray
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return jsonX509Config, err
	}

	// Initialize the final X509 Configuration array
	// Unmarshal byteArray with the jsonFile's content into jsonX509Config
	json.Unmarshal(byteValue, &jsonX509Config)
	if err != nil {
		return jsonX509Config, err
	}

	// Close JSON config file
	if jsonFile.Close(); err != nil {
		return jsonX509Config, err
	}

	return jsonX509Config, nil
}

// ----------------------------------------------------------
func dumpJSONConfig(x509config *X509Config) error {

	log.Println("")
	log.Println("Configuration Parameters:")
	log.Println("- create_new_rootca: " + x509config.CreateNewRootCA)
	log.Println("- working_dir      : " + x509config.WorkingDir)
	log.Println("- pki_setup_dir    : " + x509config.PKISetupDir)
	log.Println("- dump_config      : " + x509config.DumpConfig)
	log.Println("Key Schemes Parameters:")
	log.Println("- dump_keys        : " + x509config.KeyScheme.DumpKeys)
	log.Println("- rsa              : " + x509config.KeyScheme.RSA)
	log.Println("- rsa_key_size     : " + x509config.KeyScheme.RSAKeySize)
	log.Println("- ec               : " + x509config.KeyScheme.EC)
	log.Println("- ec_curve         : " + x509config.KeyScheme.ECCurve)
	log.Println("Root CA Parameters:")
	log.Println("- ca_name          : " + x509config.RootCA.CAName)
	log.Println("- ca_c             : " + x509config.RootCA.CACountry)
	log.Println("- ca_st            : " + x509config.RootCA.CAState)
	log.Println("- ca_l             : " + x509config.RootCA.CALocality)
	log.Println("- ca_o             : " + x509config.RootCA.CAOrg)
	log.Println("TLS Server Parameters:")
	log.Println("- tls_host         : " + x509config.TLSServer.TLSHost)
	log.Println("- tls_domain       : " + x509config.TLSServer.TLSDomain)
	log.Println("- tls_c            : " + x509config.TLSServer.TLSCountry)
	log.Println("- tls_st           : " + x509config.TLSServer.TLSSate)
	log.Println("- tls_l            : " + x509config.TLSServer.TLSLocality)
	log.Println("- tls_o            : " + x509config.TLSServer.TLSOrg)

	return nil
}

// ----------------------------------------------------------
func createEnv(x509config *X509Config) error {

	// Abs returns an absolute representation of path.
	// If the path is not absolute it will be joined with the current working directory
	// to turn it into an absolute path.
	wDir, err := filepath.Abs(x509config.WorkingDir)
	if err != nil {
		fatalIfErr(err, "ERROR: failed to build the working directory absolute path from JSON config")
	}
	log.Printf("Working directory: %s", wDir)

	// pkiCaDir: Concatenate working dir absolute path with PKI setup dir, using separator "/"
	pkiCaDir = strings.Join([]string{wDir, x509config.PKISetupDir, x509config.RootCA.CAName}, "/")

	// Convert create_new_ca JSON string "true|false" to boolean
	newCA, err = strconv.ParseBool(x509config.CreateNewRootCA)
	// Convert dump_config JSON string "true|false" to boolean
	dumpConfig, err = strconv.ParseBool(x509config.DumpConfig)
	// Convert dump_keys JSON string "true|flase| to boolean
	dumpKeys, err = strconv.ParseBool(x509config.KeyScheme.DumpKeys)
	// Convert rsa JSON string "true|false" to boolean
	rsaScheme, err = strconv.ParseBool(x509config.KeyScheme.RSA)
	// Convert rsa_key_size JSON string to integer
	rsaKeySize, err = strconv.Atoi(x509config.KeyScheme.RSAKeySize)
	// Convert ec JSON string "true|false" to boolean
	ecScheme, err = strconv.ParseBool(x509config.KeyScheme.EC)
	// EC chosen curve
	ecCurve = x509config.KeyScheme.ECCurve
	// Init: CA name and PEM key/cert filenames
	caName = x509config.RootCA.CAName
	caKeyFile = filepath.Join(pkiCaDir, caName+skFileExt)
	caCertFile = filepath.Join(pkiCaDir, caName+certFileExt)
	// Init: TLS host.domain and PEM key/cert filenames
	tlsHost = x509config.TLSServer.TLSHost
	tlsDomain = x509config.TLSServer.TLSDomain
	if tlsDomain == "local" {
		tlsFQDN = tlsHost
		tlsAltFQDN = tlsHost + "." + tlsDomain
	} else {
		tlsFQDN = tlsHost + "." + tlsDomain
		tlsAltFQDN = ""
	}
	tlsKeyFile = filepath.Join(pkiCaDir, tlsHost+skFileExt)
	tlsCertFile = filepath.Join(pkiCaDir, tlsHost+certFileExt)
	// CA subjects
	caCountry = x509config.RootCA.CACountry
	caState = x509config.RootCA.CAState
	caLocality = x509config.RootCA.CALocality
	caOrg = x509config.RootCA.CAOrg
	// TLS subjects
	tlsCountry = x509config.TLSServer.TLSCountry
	tlsState = x509config.TLSServer.TLSSate
	tlsLocality = x509config.TLSServer.TLSLocality
	tlsOrg = x509config.TLSServer.TLSOrg

	// Print the JSON parameters to console
	if dumpConfig {
		// inside createEnv, local x509config is already an address
		// as it was passed to createEnv w/ the & prefix
		_ = dumpJSONConfig(x509config)
	}

	// Creating a new fresh PKI setup dir, if new CA is requested
	if newCA {
		// Remove eventual previous PKI setup directory
		// Create a new empty PKI setup directory
		log.Printf("")
		log.Println("New CA creation requested by configuration")
		log.Println("Cleaning up CA PKI setup directory")

		err = os.RemoveAll(pkiCaDir) // Remove pkiCaDir
		if err != nil {
			return fmt.Errorf("Failed to remove the existing CA PKI configuration directory: %s (%s)", pkiCaDir, err)
		}

		log.Printf("Creating CA PKI setup directory: %s", pkiCaDir)
		err = os.MkdirAll(pkiCaDir, 0750) // Create pkiCaDir
		if err != nil {
			return fmt.Errorf("Failed to create the CA PKI configuration directory: %s (%s)", pkiCaDir, err)
		}
	} else { // Using an existing PKI setup directory, if new CA is *NOT* requested
		log.Println("No new CA creation requested by configuration")

		// Is the CA there? (if nil then OK... but could be something else than a directory)
		stat, err := os.Stat(pkiCaDir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("CA PKI setup directory does not exist: %s", pkiCaDir)
			}
			return fmt.Errorf("CA PKI setup directory cannot be reached: %s (%s)", pkiCaDir, err)
		}
		if stat.IsDir() {
			log.Printf("Existing CA PKI setup directory: %s", pkiCaDir)
		} else {
			//return errors.New("Existing CA PKI setup directory is not a directory")
			return fmt.Errorf("Existing CA PKI setup directory is not a directory: %s", pkiCaDir)
		}
	}
	return nil
}

// ----------------------------------------------------------
func fatalIfErr(err error, msg string) {
	if err != nil {
		log.Fatalf("ERROR: %s: %s", msg, err) // fatalf() =  Prinf() followed by a call to os.Exit(1)
	}
}
