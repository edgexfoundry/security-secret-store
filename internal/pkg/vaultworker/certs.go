/*******************************************************************************
 * Copyright 2018 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @author: Tingyu Zeng, Dell
 * @version: 1.0.0
 *******************************************************************************/
package vaultworker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/dghubble/sling"
	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	model "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// CertKeyPair X.509 TLS certioficate and associated private key
type CertKeyPair struct {
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
}

// CertKeyCollector X.509 TLS certificate and associated private key from Secret Store get req
type CertKeyCollector struct {
	Section CertKeyPair `json:"data"`
}

// CertInfo parm
type CertInfo struct {
	Cert string   `json:"cert,omitempty"`
	Key  string   `json:"key,omitempty"`
	Snis []string `json:"snis,omitempty"`
}

var lc = CreateLogging()

// CreateLogging Logger functionality
func CreateLogging() logger.LoggingClient {
	return logger.NewClient(SecurityService, false, fmt.Sprintf("%s-%s.log", SecurityService, time.Now().Format("2006-01-02")), model.InfoLog)
}

func LoadKongCerts(config *tomlConfig, url string, secretBaseURL string, c *http.Client, debug bool) error {
	cert, key, err := getCertKeyPair(config, secretBaseURL, c, debug)
	if err != nil {
		return err
	}
	body := &CertInfo{
		Cert: cert,
		Key:  key,
		Snis: []string{config.SecretService.SNIS},
	}
	lc.Info("Trying to upload cert to proxy server.")
	req, err := sling.New().Base(url).Post(CertificatesPath).BodyJSON(body).Request()
	resp, err := c.Do(req)
	if err != nil {
		lc.Error("Failed to upload cert to proxy server with error %s", err.Error())
		return err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusConflict {
		lc.Info("Successful to add certificate to the reverse proxy.")
	} else {
		s := fmt.Sprintf("Failed to add certificate with errorcode %d.", resp.StatusCode)
		return errors.New(s)
	}
	return nil
}

func getCertKeyPair(config *tomlConfig, secretBaseURL string, c *http.Client, debug bool) (string, string, error) {

	t, err := GetSecret(filepath.Join(config.SecretService.TokenFolderPath, config.SecretService.VaultInitParm))
	if err != nil {
		return "", "", err
	}

	s := sling.New().Set(VaultToken, t.Token)
	req, err := s.New().Base(secretBaseURL).Get(config.SecretService.CertPath).Request()
	resp, err := c.Do(req)
	if err != nil {
		errStr := fmt.Sprintf("Failed to retrieve certificate with path as %s with error %s", config.SecretService.CertPath, err.Error())
		return "", "", errors.New(errStr)
	}
	defer resp.Body.Close()

	collector := CertKeyCollector{}
	json.NewDecoder(resp.Body).Decode(&collector)

	switch resp.StatusCode {
	case http.StatusOK:
		lc.Info(fmt.Sprintf("API Gateway TLS certificate/key found in Secret Store @/%s (%s)", config.SecretService.CertPath, resp.Status))
		if debug {
			lc.Info(fmt.Sprintf("\n %s \n \n %s", collector.Section.Cert, collector.Section.Key))
		}

	case http.StatusNotFound:
		lc.Info(fmt.Sprintf("API Gateway TLS certificate/key NOT found in Secret Store @/%s (%s)", config.SecretService.CertPath, resp.Status))

	default:
		lc.Info(fmt.Sprintf("Failed reading API Gateway TLS certificate/key from Secret Store @/%s (%s)", config.SecretService.CertPath, resp.Status))
	}

	return collector.Section.Cert, collector.Section.Key, nil
}

func CertKeyPairInStore(config *tomlConfig, secretBaseURL string, c *http.Client, debug bool) (bool, error) {
	cert, key, err := getCertKeyPair(config, secretBaseURL, c, debug)
	if err != nil {
		return false, err
	}
	if len(cert) > 0 && len(key) > 0 {
		return true, nil
	}
	return false, nil
}

func LoadCACert(caPath string) (string, error) {
	certPEMBlock, err := ioutil.ReadFile(caPath)
	if err != nil {
		return "", err
	}
	cert := string(certPEMBlock[:])

	return cert, nil
}

func LoadCertKeyPair(certPath string, keyPath string) (string, string, error) {
	certPEMBlock, err := ioutil.ReadFile(certPath)
	if err != nil {
		return "", "", err
	}
	cert := string(certPEMBlock[:])

	keyPEMBlock, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return "", "", err
	}
	key := string(keyPEMBlock[:])

	return cert, key, nil
}
