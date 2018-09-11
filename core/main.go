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
	waitInterval := flag.Int("wait", 180, "time to wait between checking the vault status in seconds.")

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
		time.Sleep(time.Second * time.Duration(*waitInterval))
	}

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

	cert, sk, err := loadCerKeyPair(config.SecretService.CertFilePath, config.SecretService.KeyFilePath)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to load cert&key pair from volume with path of cert - %s, key - %s.", config.SecretService.CertFilePath, config.SecretService.KeyFilePath))
		os.Exit(0)
	}
	lc.Info("Load cert&key pair from volume successfully, now will upload to secret store.")
	uploadProxyCerts(config, secretServiceBaseURL, cert, sk, client)
}
