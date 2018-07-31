package main

import (
	"crypto/tls"
	"flag"
	"fmt"
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
		TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecureSkipVerify},
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}

	a, err := api.NewClient(&api.Config{
		Address:    secretServiceBaseURL,
		HttpClient: client,
	})
	s := a.Sys()
	k, err := initVault(s, config.SecretService.TokenPath)
	if err != nil {
		lc.Error(fmt.Sprintf("Error while initializing the vault with error info: %s", err.Error()))
		os.Exit(0)
	}
	sealed, err := unsealVault(s, k)

	if sealed == false {
		lc.Info("Vault has been unsealed successfully.")
	} else {
		lc.Error("Vault is still under sealed status.")
	}

	cert, sk, err := generateCerKeyPair()
	uploadProxyCerts(config, secretServiceBaseURL, cert, sk, client)
}
