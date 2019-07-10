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
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
)

/*GenSK creates a new RSA or EC based private key (sk)*/
func genSK(cf *CertConfig) (crypto.PrivateKey, error) {

	if cf.rsaScheme {
		log.Printf("- Generating private key with RSA scheme %d", cf.rsaKeySize)
		return rsa.GenerateKey(rand.Reader, cf.rsaKeySize)
	}

	if cf.ecScheme {
		log.Printf("- Generating private key with EC scheme %s", cf.ecCurve)
		var curve elliptic.Curve
		switch cf.ecCurve {
		case "224": // secp224r1 NIST P-224
			curve = elliptic.P224()
		case "256": // secp256v1 NIST P-256
			curve = elliptic.P256()
		case "384": // secp384r1 NIST P-384
			curve = elliptic.P384()
		case "521": // secp521r1 NIST P-521
			curve = elliptic.P521()
		default:
			return nil, fmt.Errorf("Unknown elliptic curve: %q", cf.ecCurve)
		}
		return ecdsa.GenerateKey(curve, rand.Reader)
	}

	return nil, fmt.Errorf("Unknown key scheme: RSA[%t] EC[%t]", cf.rsaScheme, cf.ecScheme)
}

/*dumpKeyPair output sk,pk keypair (RSA or EC) to console. !!! Debug only for obvious security reasons...*/
func dumpKeyPair(sk crypto.PrivateKey, pk crypto.PublicKey) error {

	log.Println("")
	switch sk.(type) {
	case *rsa.PrivateKey:
		log.Printf(">> RSA SK: %q", sk)
	case *ecdsa.PrivateKey:
		log.Printf(">> ECDSA SK: %q", sk)
	default:
		log.Println("Unsupported Private Key")
	}

	log.Println("")
	switch pk.(type) {
	case *rsa.PublicKey:
		log.Printf(">> RSA PK: %q", pk)
	case *ecdsa.PublicKey:
		log.Printf(">> ECDSA PK: %q", pk)
	default:
		log.Println("Unsupported Public Key")
	}
	log.Println("")

	return nil
}
