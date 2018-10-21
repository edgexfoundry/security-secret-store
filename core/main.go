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

	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/hashicorp/vault/api"
)

var lc = CreateLogging()

// CreateLogging Logger functionality
func CreateLogging() logger.LoggingClient {
	return logger.NewClient(SecurityService, false, fmt.Sprintf("%s-%s.log", SecurityService, time.Now().Format("2006-01-02")))
}

func main() {

	if len(os.Args) < 2 {
		HelpCallback()
	}

	useConsul := flag.Bool("consul", false, "retrieve configuration from consul server")
	initNeeded := flag.Bool("init", false, "run init procedure for security service.")
	insecureSkipVerify := flag.Bool("insureskipverify", true, "skip server side SSL verification, mainly for self-signed cert.")
	configFileLocation := flag.String("configfile", "res/configuration.toml", "configuration file")
	waitInterval := flag.Int("wait", 30, "time to wait between checking the vault status in seconds.")

	flag.Usage = HelpCallback
	flag.Parse()

	if *useConsul {
		lc.Info("Retrieving config data from Consul...")
	}

	if *initNeeded == false {
		lc.Info("skipping initlization and exit. Hint: are you trying to initialize the secret store ? please use the option with --init=true.")
		os.Exit(0)
	}

	config, err := LoadTomlConfig(*configFileLocation)
	if err != nil {
		lc.Error("Failed to retrieve config data from local file. Please make sure res/configuration.toml file exists with correct formats.")
		return
	}
	secretServiceBaseURL := fmt.Sprintf("https://%s:%s/", config.SecretService.Server, config.SecretService.Port)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: *insecureSkipVerify,
		},
	}
	if *insecureSkipVerify == false {
		caCert, err := ioutil.ReadFile(config.SecretService.CAFilePath)
		if err != nil {
			lc.Error("Failed to load rootCA cert.")
			os.Exit(0)
		}
		lc.Info("successful loading the rootCA cert.")
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tr = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: *insecureSkipVerify,
			},
		}
	}

	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}

	a, err := api.NewClient(&api.Config{
		Address:    secretServiceBaseURL,
		HttpClient: client,
	})
	s := a.Sys()
	inited := false
	sealed := true

	for {
		inited, err = s.InitStatus()
		if err != nil {
			lc.Error(fmt.Sprintf("Error while checking the initialization status: %s", err.Error()))
		} else {
			k, err := initVault(s, config.SecretService.TokenPath, inited)
			if err != nil {
				lc.Error(fmt.Sprintf("Error while initializing the vault with info: %s", err.Error()))
			}

			sealed, err = unsealVault(s, k)
			if err != nil {
				lc.Error(err.Error())
			}

			if sealed == false {
				lc.Info("Vault has been unsealed successfully.")
				break
			} else {
				lc.Error("Vault is still under sealed status. Will retry again.")
			}
		}
		lc.Info(fmt.Sprintf("waiting %d seconds to retry checking the vault status.", *waitInterval))
		time.Sleep(time.Second * time.Duration(*waitInterval))
	}

	_, _ = vaultHealthCheck(config, client)

	// -----------------------------------------------------------------------------------
	// Importing Admin and Kong Policies in Vault + create corresponding tokens
	// -----------------------------------------------------------------------------------
	// Get the Vault Root Token generated after Vault initialization
	rootToken, err := getSecret(config.SecretService.TokenPath)
	if err != nil {
		lc.Error("Fatal Error fetching Vault root token.")
		fatalIfErr(err, "Root token fetch failure")
	}

	/*
		Wait 5" till Vault has completed the post unseal cluster/node/backend tasks,
		otherwise the REST API request returns a HTTP Status 500...

		edgex-vault-worker | INFO: 2018/10/20 10:52:55 Vault has been initialized successfully.
		edgex-vault-worker | INFO: 2018/10/20 10:52:55 Vault has been unsealed successfully.
		edgex-vault-worker | INFO: 2018/10/20 10:52:55 Reading Admin policy file.
		edgex-vault-worker | INFO: 2018/10/20 10:52:55 Importing Vault Admin policy.
		edgex-vault-worker | 2018/10/20 10:52:55 ERROR: Import policy failure: admin
		edgex-vault-worker | ERROR: 2018/10/20 10:52:55 Import Policy HTTP Status: 500 Internal Server Error (StatusCode: 500)
		edgex-vault-worker | ERROR: 2018/10/20 10:52:55 Fatal Error importing Admin policy in Vault.
	*/
	time.Sleep(5 * time.Second)

	// Read the Admin HCL config file and build the policy request
	lc.Info("Reading Admin policy file.")
	policyFile := config.SecretService.PolicyPath4Admin
	policyRequest, err := getPolicyFromFile(&policyFile)
	if err != nil {
		lc.Error("Fatal Error opening Admin policy file.")
		fatalIfErr(err, "Opening policy file (Admin)")
	}

	// Import the Admin policy data into Vault
	lc.Info("Importing Vault Admin policy.")
	err = importPolicy(config.SecretService.PolicyName4Admin, &policyRequest, rootToken.Token, config, client)
	if err != nil {
		lc.Error("Fatal Error importing Admin policy in Vault.")
		fatalIfErr(err, "Import policy failure")
	}

	// Read the Kong HCL file and build the policy request
	lc.Info("Reading Kong policy file.")
	policyFile = config.SecretService.PolicyPath4Kong
	policyRequest, err = getPolicyFromFile(&policyFile)
	if err != nil {
		lc.Error("Fatal Error opening Kong policy file.")
		fatalIfErr(err, "Opening policy file (Kong)")
	}

	// Import the Kong policy data into Vault
	lc.Info("Importing Vault Kong policy.")
	err = importPolicy(config.SecretService.PolicyName4Kong, &policyRequest, rootToken.Token, config, client)
	if err != nil {
		lc.Error("Fatal Error importing Kong policy in Vault.")
		fatalIfErr(err, "Import policy failure")
	}

	// Create Admin token associated with admin policy in Vault
	lc.Info("Creating Vault Admin token.")
	err = createToken(config.SecretService.TokenName4Admin, config.SecretService.PolicyName4Admin, rootToken.Token, config, client)
	if err != nil {
		lc.Error("Fatal Error creating Admin token in Vault.")
		fatalIfErr(err, "Create token failure (Admin)")
	}

	// Create Kong token associated with kong policy in Vault
	lc.Info("Creating Vault Kong token.")
	err = createToken(config.SecretService.TokenName4Kong, config.SecretService.PolicyName4Kong, rootToken.Token, config, client)
	if err != nil {
		lc.Error("Fatal Error creating Kong token in Vault.")
		fatalIfErr(err, "Create token failure (Kong)")
	}
	// -----------------------------------------------------------------------------------

	hasCertKeyPair, err := certKeyPairInStore(config, secretServiceBaseURL, client)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to check if the cert&key pair is in secret store with error %s.", err.Error()))
		os.Exit(0)
	}

	if hasCertKeyPair == true {
		lc.Info("Cert&key pair is already in secret store, skip uploading cert step.")
		os.Exit(0)
	}
	lc.Info("Cert&key are not in the secret store yet, will need to upload them.")

	cert, sk, err := loadCertKeyPair(config.SecretService.CertFilePath, config.SecretService.KeyFilePath)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to load cert&key pair from volume with path of cert - %s, key - %s.", config.SecretService.CertFilePath, config.SecretService.KeyFilePath))
		os.Exit(0)
	}
	lc.Info("Load cert&key pair from volume successfully, now will upload to secret store.")

	for {
		done, _ := uploadProxyCerts(config, secretServiceBaseURL, cert, sk, client)
		if done == true {
			// Entropy
			os.Exit(0)
		} else {
			lc.Info(fmt.Sprintf("will retry uploading in %d seconds.", *waitInterval))
		}
		time.Sleep(time.Second * time.Duration(*waitInterval))
	}
}
