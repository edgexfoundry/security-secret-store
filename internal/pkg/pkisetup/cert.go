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
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"time"
)

//GenCA creates a new CA certificate, saves it to PEM file and returns the x509 certificate and crypto private key.*/
func GenCA(cf *CertConfig) (*x509.Certificate, crypto.PrivateKey, error) {

	log.Println("")
	log.Println("<Phase 1> Generating CA PKI materials")
	log.Println("Generating Root CA key pair (sk,pk)")

	// Generate RSA or EC based SK
	caSK, err := genSK(cf)
	if err != nil {
		return nil, nil, err
	}
	// Extract PK from RSA or EC generated SK
	caPK := caSK.(crypto.Signer).Public()
	// Debug the key pair generation/extraction
	if cf.dumpKeys {
		dumpKeyPair(caSK, caPK)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	FatalIfErr(err, "failed to generate serial number")

	spkiASN1, err := x509.MarshalPKIXPublicKey(caPK)
	FatalIfErr(err, "failed to encode public key")

	var spki struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	_, err = asn1.Unmarshal(spkiASN1, &spki)
	FatalIfErr(err, "failed to decode public key")

	skid := sha1.Sum(spki.SubjectPublicKey.Bytes)

	caCertTemplate := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         cf.caName,
			Organization:       []string{cf.caName},
			OrganizationalUnit: []string{cf.caOrg},
			Locality:           []string{cf.caLocality},
			Province:           []string{cf.caState},
			Country:            []string{cf.caCountry},
		},

		EmailAddresses: []string{cf.caName + "@" + cf.tlsDomain},

		SubjectKeyId: skid[:],

		NotAfter:  time.Now().AddDate(10, 0, 0),
		NotBefore: time.Now(),

		KeyUsage: x509.KeyUsageCertSign,

		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}

	log.Printf("Generating Root CA certificate")
	caDER, err := x509.CreateCertificate(rand.Reader, caCertTemplate, caCertTemplate, caPK, caSK)
	FatalIfErr(err, "failed to generate CA certificate (DER)")

	caCert, err := x509.ParseCertificate(caDER)
	FatalIfErr(err, "failed to parse Root CA certificate")

	log.Printf("Saving Root CA private key to PEM file: %s", cf.caKeyFile)
	skPKCS8, err := x509.MarshalPKCS8PrivateKey(caSK)
	FatalIfErr(err, "failed to encode CA private key")

	err = ioutil.WriteFile(cf.caKeyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: skPKCS8}), 0400)
	FatalIfErr(err, "failed to save CA private key")

	log.Printf("Saving Root CA certificate to PEM file: %s", cf.caCertFile)
	err = ioutil.WriteFile(cf.caCertFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}), 0644)
	FatalIfErr(err, "failed to save CA certificate")

	log.Printf("New local Root CA successfully created!")

	return caCert, caSK, nil
}

/*GenCert creates a new TLS server certificate, saves it to PEM file and returns the x509 certificate and crypto private key. */
func GenCert(cf *CertConfig) (*x509.Certificate, crypto.PrivateKey, error) {

	log.Println("")
	log.Println("<Phase 2> Generating TLS server PKI materials")

	// Root CA certificate fetch --------------------------------------------------------
	log.Printf("Loading Root CA certificate: %s", cf.caCertFile)
	certPEMBlock, err := ioutil.ReadFile(cf.caCertFile) // Load Root CA certificate
	FatalIfErr(err, "failed to read the Root CA certificate")

	log.Println("- Decoding the Root CA certificate")
	certDERBlock, _ := pem.Decode(certPEMBlock) // Decode Root CA certificate
	if certDERBlock == nil || certDERBlock.Type != "CERTIFICATE" {
		log.Fatalln("ERROR: failed to read the Root CA certificate: unexpected content")
	}

	log.Println("- Parsing the Root CA certificate")
	caCert, err := x509.ParseCertificate(certDERBlock.Bytes) // Parse Root CA certificate
	FatalIfErr(err, "failed to parse the Root CA certificate")

	// Root CA private key fetch --------------------------------------------------------
	log.Printf("Loading the Root CA private key: %s", cf.caKeyFile)
	keyPEMBlock, err := ioutil.ReadFile(cf.caKeyFile)
	FatalIfErr(err, "failed to read the Root CA private key")

	log.Println("- Decoding the Root CA private key")
	keyDERBlock, _ := pem.Decode(keyPEMBlock) // Decode Root CA private key
	if keyDERBlock == nil || keyDERBlock.Type != "PRIVATE KEY" {
		log.Fatalln("ERROR: failed to read the Root CA key: unexpected content")
	}

	log.Println("- Parsing the Root CA private key")
	caSK, err := x509.ParsePKCS8PrivateKey(keyDERBlock.Bytes) // Parse Root CA private key
	FatalIfErr(err, "failed to parse the Root CA key")

	// TLS server certificate preparation -----------------------------------------------
	log.Println("Generating TLS server key pair (sk,pk)")

	// Generate RSA or EC based SK
	tlsSK, err := genSK(cf)
	if err != nil {
		return nil, nil, err
	}
	// Extract PK from RSA or EC generated SK
	tlsPK := tlsSK.(crypto.Signer).Public()
	// Debug the key pair generation/extraction
	if cf.dumpKeys {
		dumpKeyPair(tlsSK, tlsPK)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	FatalIfErr(err, "failed to generate serial number")

	tlsCertTemplate := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         cf.tlsFQDN,
			Organization:       []string{cf.tlsHost},
			OrganizationalUnit: []string{cf.tlsOrg},
			Locality:           []string{cf.tlsLocality},
			Province:           []string{cf.tlsState},
			Country:            []string{cf.tlsCountry},
		},

		EmailAddresses: []string{"admin@" + cf.tlsDomain},
		DNSNames:       []string{cf.tlsFQDN, cf.tlsAltFQDN}, // Alternative Names

		NotAfter:  time.Now().AddDate(10, 0, 0),
		NotBefore: time.Now(),

		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		BasicConstraintsValid: true,
	}

	log.Printf("Generating TLS server certificate (Self-signed with our local Root CA)")
	tlsDER, err := x509.CreateCertificate(rand.Reader, tlsCertTemplate, caCert, tlsPK, caSK)
	FatalIfErr(err, "failed to generate TLS server certificate - DER")

	tlsCert, err := x509.ParseCertificate(tlsDER)
	FatalIfErr(err, "failed to parse TLS server certificate")

	log.Printf("Saving TLS server private key to PEM file: %s", cf.tlsKeyFile)
	skPKCS8, err := x509.MarshalPKCS8PrivateKey(tlsSK)
	FatalIfErr(err, "failed to encode TLS server key")

	err = ioutil.WriteFile(cf.tlsKeyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: skPKCS8}), 0600)
	FatalIfErr(err, "failed to save TLS server private key")

	log.Printf("Saving Root CA certificate to PEM file: %s", cf.caCertFile)
	err = ioutil.WriteFile(cf.tlsCertFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: tlsDER}), 0644)
	FatalIfErr(err, "failed to save TLS server certificate")

	log.Printf("New TLS server certificate/key successfully created!")

	return tlsCert, tlsSK, nil
}
