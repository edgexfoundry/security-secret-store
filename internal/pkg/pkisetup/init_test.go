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

var testKeyScheme = KeyScheme {
	DumpKeys   : "false",
	RSA        : "false",
	RSAKeySize : "4096",
	EC         : "true",
	ECCurve   : "384",
}

var testRootCA = RootCA {
	CAName    : "testCA.test",
	CACountry : "testCACountry",
	CAState  : "testCAState",
	CALocality : "testCALocality",
	CAOrg    : "testCAOrg",
}

var testTLSServer = TLSServer {
	TLSHost    : "testtlshost",
	TLSDomain  : "testtlsdomain",
	TLSCountry  : "testtlscountry",
	TLSSate    : "testtlsstate",
	TLSLocality : "testtlslocality",
	TLSOrg     : "testtlsorg",
}

var testX509Config = X509Config {
	CreateNewRootCA : "true",
	WorkingDir     : "./testconfig",
	PKISetupDir    : "pki",
	DumpConfig    :  "true",
	KeyScheme     : testKeyScheme,
	RootCA       :  testRootCA,
	TLSServer    :   testTLSServer,
}

func TestCreateEnvWithValidConfig(t *testing.T) {
    myconfig := testX509Config
    _, err := CreateEnv(&myconfig)

    if err != nil {
        t.Errorf("Failed to create X509 env with correct configuration data.")
    }
}
