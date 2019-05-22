/*
   Copyright 2019 DELL Technologies.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
  
  @author: Tingyu Zeng, DELL (created: May 21th, 2019)
  @version: 1.0.0
*/

package pkisetup

import (
	"testing"
)

var testConfig = CertConfig {        
        configFile: "testconfigfile",
        pkiCaDir: "pki",
        dumpConfig: true,
        newCA: true,  
    
        caName     :"testCA",
        caKeyFile  :"testCAKeyFile.test",
        caCertFile :"testCACertFile.test",
        caCountry  :"testCountry",
        caState     :"testcaState",
        caLocality: "testcaLocality",
        caOrg  :    "testcaOrg",
    
        // TLS Server Certificate
        tlsHost     : "testtlsHost",
        tlsDomain   : "testtlsDomain",
        tlsFQDN     : "testtlsFQDN",
        tlsAltFQDN  : "testtlsAltFQDN",
        tlsKeyFile  : "testtlsKeyFile",
        tlsCertFile : "testtlsCertFile",
        tlsCountry  : "testtlsCountry",
        tlsState    : "testtlsState",
        tlsLocality : "testtlsLocality",
        tlsOrg      : "testtlsOrg",
    
        // Key Generation
        dumpKeys  : true,
        rsaScheme  : true,
        rsaKeySize : 4096,
        ecScheme   : true,
        ecCurve   : "224",    
}

func TestGenCAWithValidConfig(t *testing.T) {
    myconfig := testConfig
    _, _, err := GenCA(&myconfig)
    if err != nil {
        t.Errorf("Failed to create CA with correct configuration data.")
    }
}

func TestGenCertWithValidConfig(t *testing.T) {
    myconfig := testConfig
    _, _, err := GenCert(&myconfig)
    if err != nil {
        t.Errorf("Failed to create cert with correct configuration data.")
    }
}