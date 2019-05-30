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

  @author: Alain Pulluelo, ForgeRock AS (created: July 27, 2018)
  @author: Tingyu Zeng, DELL (updated: May 22, 2019)
  @version: 1.0.0
*/

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	model "github.com/edgexfoundry/go-mod-core-contracts/models"
	worker "github.com/edgexfoundry/security-secret-store/internal/pkg/vaultworker"
)

var debug = false
var lc = CreateLogging()

// CreateLogging Logger functionality
func CreateLogging() logger.LoggingClient {
	return logger.NewClient(worker.SecurityService, false, fmt.Sprintf("%s-%s.log", SecurityService, time.Now().Format("2006-01-02")), model.InfoLog)
}


func main() {

	lc.Info("-------------------- Vault Worker Cycle ------------------------")

	if len(os.Args) < 2 {
		worker.HelpCallback()
	}

	useConsul := flag.Bool("consul", false, "retrieve configuration from consul server")
	initNeeded := flag.Bool("init", false, "run init procedure for security service.")
	debugActive := flag.Bool("debug", false, "output sensitive debug informations for security service.")
	insecureSkipVerify := flag.Bool("insureskipverify", true, "skip server side SSL verification, mainly for self-signed cert.")
	configFileLocation := flag.String("configfile", "res/configuration.toml", "configuration file")
	waitInterval := flag.Int("wait", 30, "time to wait between checking Vault status in seconds.")

	flag.Usage = worker.HelpCallback
	flag.Parse()

	if *debugActive {
		lc.Info("Debugging mode activated.")
		debug = true
	}
	if *useConsul {
		lc.Info("Retrieving config data from Consul...")
	}

	if *initNeeded == false {
		lc.Info("skipping initlization and exit. Hint: are you trying to initialize the secret store ? please use the option with --init=true.")
		os.Exit(0)
	}

	config, err := worker.LoadTomlConfig(*configFileLocation)
	if err != nil {
		lc.Error("Failed to retrieve config data from local file. Please make sure res/configuration.toml file exists with correct format.")
		return
	}

	// Prepare the HTTP Client to use with Vault REST API
	// 1/2 Build Transport
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: *insecureSkipVerify,
		},
	}
	// Add TLS support if requested
	if *insecureSkipVerify == false {
		caCert, err := ioutil.ReadFile(config.SecretService.CAFilePath)
		if err != nil {
			lc.Error("Failed to load rootCA certificate.")
			os.Exit(0)
		}
		lc.Info("Successful loading the rootCA certificate.")
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tr = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: *insecureSkipVerify,
			},
			TLSHandshakeTimeout: 5 * time.Second,
		}
	}

	// 2/2 Build HTTP Client
	client := &http.Client{Transport: tr, Timeout: 10 * time.Second}

	// Loop duration interval between Vault init and unseal retries
	intervalDuration := time.Duration(*waitInterval) * time.Second
	// Loop exit condition
	loopExit := false

	for {
		sCode, _ := worker.VaultHealthCheck(config, client)

		switch sCode {
		case 200:
			lc.Info(fmt.Sprintf("Vault is initialized and unsealed (Status Code: %d).", sCode))
			loopExit = true
		case 429:
			lc.Error(fmt.Sprintf("Vault is unsealed and in standby mode (Status Code: %d).", sCode))
			loopExit = true
		case 501:
			lc.Info(fmt.Sprintf("Vault is not initialized (Status Code: %d). Starting initialisation and unseal phases.", sCode))
			_, err = worker.VaultInit(config, client, debug)
			if err == nil {
				_, err = worker.VaultUnseal(config, client, debug)
				if err == nil {
					loopExit = true
				}
			}
		case 503:
			lc.Info(fmt.Sprintf("Vault is sealed (Status Code: %d). Starting unseal phase...", sCode))
			_, err = worker.VaultUnseal(config, client, debug)
			if err == nil {
				loopExit = true
			}
		default:
			if sCode == 0 {
				lc.Error(fmt.Sprintf("Vault is in an unknown state. No Status code available."))
			} else {
				lc.Error(fmt.Sprintf("Vault is in an unknown state. Status code: %d", sCode))
			}
		}

		if loopExit {
			break
		}
		lc.Info(fmt.Sprintf("Next Vault Init/Unseal attempt in %d seconds.", *waitInterval))
		time.Sleep(intervalDuration)
	}

	// -----------------------------------------------------------------------------------
	// Importing Admin and Kong Policies in Vault + create corresponding tokens
	// -----------------------------------------------------------------------------------
	// Get the Vault Root Token generated after Vault initialization
	rootToken, err := worker.GetSecret(config.SecretService.TokenFolderPath + "/" + config.SecretService.VaultInitParm)
	if err != nil {
		lc.Error("Fatal Error fetching Vault root token.")
		worker.FatalIfErr(err, "Root token fetch failure")
	}

	/*
		Till Vault has completed the post unseal cluster/node/backend tasks,
		otherwise the REST API request returns a HTTP Status 500...

		edgex-vault-worker | INFO: 2018/10/20 10:52:55 Vault has been initialized successfully.
		edgex-vault-worker | INFO: 2018/10/20 10:52:55 Vault has been unsealed successfully.
		edgex-vault-worker | INFO: 2018/10/20 10:52:55 Reading Admin policy file.
		edgex-vault-worker | INFO: 2018/10/20 10:52:55 Importing Vault Admin policy.
		edgex-vault-worker | 2018/10/20 10:52:55 ERROR: Import policy failure: admin
		edgex-vault-worker | ERROR: 2018/10/20 10:52:55 Import Policy HTTP Status: 500 Internal Server Error (StatusCode: 500)
		edgex-vault-worker | ERROR: 2018/10/20 10:52:55 Fatal Error importing Admin policy in Vault.
	*/
	for {
		if sCode, _ := worker.VaultHealthCheck(config, client); sCode == 200 { // Healthcheck output (code 200 is OK status)
			break
		}
	}

	// ------------------ Admin Vault Policies and associated token ----------------------
	// Read the Admin HCL config file and build the policy request
	lc.Info("Verifying Admin policy file hash (SHA256).")
	policyFile := config.SecretService.PolicyPath4Admin
	_, err = worker.HashFile(&policyFile, debug)
	if err != nil {
		worker.FatalIfErr(err, "Calculating policy file hash (SHA256)")
	}
	lc.Info("Reading Admin policy file.")
	policyRequest, err := worker.GetPolicyFromFile(&policyFile)
	if err != nil {
		lc.Error("Fatal Error opening Admin policy file.")
		worker.FatalIfErr(err, "Opening policy file (Admin)")
	}

	// Import the Admin policy data into Vault
	lc.Info("Importing Vault Admin policy.")
	err = worker.ImportPolicy(config.SecretService.PolicyName4Admin, &policyRequest, rootToken.Token, config, client)
	if err != nil {
		lc.Error("Fatal Error importing Admin policy in Vault.")
		worker.FatalIfErr(err, "Import policy failure")
	}

	// Create Admin token associated with admin policy in Vault
	lc.Info("Creating Vault Admin token.")
	err = worker.CreateToken(config.SecretService.TokenName4Admin, config.SecretService.PolicyName4Admin, rootToken.Token, config, client)
	if err != nil {
		lc.Error("Fatal Error creating Admin token in Vault.")
		worker.FatalIfErr(err, "Create token failure (Admin)")
	}

	// ------------------ Kong Vault Policies and associated token ----------------------
	// Read the Kong HCL file and build the policy request
	lc.Info("Verifying Kong policy file hash (SHA256).")
	policyFile = config.SecretService.PolicyPath4Kong
	_, err = worker.HashFile(&policyFile, debug)
	if err != nil {
		worker.FatalIfErr(err, "Calculating policy file hash (SHA256)")
	}
	lc.Info("Reading Kong policy file.")
	policyFile = config.SecretService.PolicyPath4Kong
	policyRequest, err = worker.GetPolicyFromFile(&policyFile)
	if err != nil {
		lc.Error("Fatal Error opening Kong policy file.")
		worker.FatalIfErr(err, "Opening policy file (Kong)")
	}

	// Import the Kong policy data into Vault
	lc.Info("Importing Vault Kong policy.")
	err = worker.ImportPolicy(config.SecretService.PolicyName4Kong, &policyRequest, rootToken.Token, config, client)
	if err != nil {
		lc.Error("Fatal Error importing Kong policy in Vault.")
		worker.FatalIfErr(err, "Import policy failure")
	}

	// Create Kong token associated with kong policy in Vault
	lc.Info("Creating Vault Kong token.")
	err = worker.CreateToken(config.SecretService.TokenName4Kong, config.SecretService.PolicyName4Kong, rootToken.Token, config, client)
	if err != nil {
		lc.Error("Fatal Error creating Kong token in Vault.")
		worker.FatalIfErr(err, "Create token failure (Kong)")
	}	

	secretServiceBaseURL := fmt.Sprintf("https://%s:%s/", config.SecretService.Server, config.SecretService.Port)

	err = worker.CredentialsInit(config, secretServiceBaseURL, client )
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to create initlization parameters in the secret store: %s", err.Error()))
		os.Exit(1)
	}

	hasCertKeyPair, err := worker.CertKeyPairInStore(config, secretServiceBaseURL, client, debug)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to check if the API Gateway TLS certificate and key are in the secret store: %s", err.Error()))
		os.Exit(1)
	}

	if hasCertKeyPair == true {
		lc.Info("API Gateway TLS certificate and key already in the secret store, skip uploading phase.")
		os.Exit(0)
	}
	lc.Info("API Gateway TLS certificate and key are not in the secret store yet, uploading them.")

	cert, sk, err := worker.LoadCertKeyPair(config.SecretService.CertFilePath, config.SecretService.KeyFilePath)
	if err != nil {
		lc.Error("Failed to load API Gateway TLS certificate and key from volume:")
		lc.Error(fmt.Sprintf("--> Certificate path: %s", config.SecretService.CertFilePath))
		lc.Error(fmt.Sprintf("--> Private Key path: %s.", config.SecretService.KeyFilePath))
		os.Exit(1)
	}
	lc.Info("API Gateway TLS certificate and key successfully loaded from volume, now will upload to secret store.")

	for {
		done, _ := worker.UploadProxyCerts(config, secretServiceBaseURL, cert, sk, client)
		if done == true {
			os.Exit(0)
		} else {
			lc.Info(fmt.Sprintf("Will retry uploading in %d seconds.", *waitInterval))
		}
		time.Sleep(time.Second * time.Duration(*waitInterval))
	}
}
