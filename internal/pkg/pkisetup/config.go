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
	"log"
	"os"	
	"io/ioutil"
	"encoding/json"
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


/*ReadConfig load the configuration from filesystem and return X509Config struct*/
func ReadConfig(configFilePtr *string) (X509Config, error) {

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
