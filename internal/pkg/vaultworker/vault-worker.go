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
 * @author: Tingyu Zeng, Dell / Alain Pulluelo, ForgeRock AS
 * @version: 1.0.0
 *******************************************************************************/
package vaultworker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/dghubble/sling"
)

// InitRequest contains a Vault init request regarding the Shamir Secret Sharing (SSS) parameters
type InitRequest struct {
	SecretShares    int `json:"secret_shares"`
	SecretThreshold int `json:"secret_threshold"`
}

// InitResponse contains a Vault init response
type InitResponse struct {
	Keys       []string `json:"keys"`
	KeysBase64 []string `json:"keys_base64"`
	RootToken  string   `json:"root_token"`
}

// UnsealRequest contains a Vault unseal request
type UnsealRequest struct {
	Key   string `json:"key"`
	Reset bool   `json:"reset"`
}

// UnsealResponse contains a Vault unseal response
type UnsealResponse struct {
	Sealed   bool `json:"sealed"`
	T        int  `json:"t"`
	N        int  `json:"n"`
	Progress int  `json:"progress"`
}

func VaultHealthCheck(config *tomlConfig, httpClient *http.Client) (sCode int, err error) {

	// Build Vault API full URL
	url, err := url.Parse(config.SecretService.Scheme + "://" + config.SecretService.Server + ":" + config.SecretService.Port + vaultHealthAPI)
	// Build Vault HTTP/GET Request
	jsonBlock := []byte(`{}`)
	req, err := http.NewRequest(http.MethodGet, url.String(), bytes.NewBuffer(jsonBlock))

	// Prepare the Header
	req.Header.Set("Content-Type", contentType)

	// GET the request
	resp, err := httpClient.Do(req)
	if err != nil {
		lc.Error(fmt.Sprintf("Failure sending the Vault health check request: %s", err.Error()))
		return 0, err
	}
	defer resp.Body.Close()

	lc.Info(fmt.Sprintf("Vault Health Check HTTP Status: %s (StatusCode: %d)", resp.Status, resp.StatusCode))

	return resp.StatusCode, nil
}

func VaultInit(config *tomlConfig, httpClient *http.Client, debug bool) (sCode int, err error) {

	// Shamir Secret Sharing parameters to apply during Vault initialization
	initRequest := InitRequest{
		SecretShares:    config.SecretService.VaultSecretShares,
		SecretThreshold: config.SecretService.VaultSecretThreshold,
	}

	lc.Info(fmt.Sprintf("Vault Init Strategy (SSS parameters): Shares=%d Threshold=%d", initRequest.SecretShares, initRequest.SecretThreshold))

	// Build Vault API full URL
	url, err := url.Parse(config.SecretService.Scheme + "://" + config.SecretService.Server + ":" + config.SecretService.Port + vaultInitAPI)
	// Build Vault HTTP/POST Request
	jsonBlock, err := json.Marshal(&initRequest) //jsonBlock := []byte(`{}`)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to build the Vault init request (SSS parameters): %s", err.Error()))
		return 0, err
	}
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(jsonBlock))

	// Prepare the Header
	req.Header.Set("Content-Type", contentType)

	// POST the request
	resp, err := httpClient.Do(req)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to send the Vault init request: %s", err.Error()))
		return 0, err
	}
	defer resp.Body.Close()

	// Init request OK/KO ?
	if resp.StatusCode != http.StatusOK {
		lc.Error(fmt.Sprintf("Vault init request failed with status code: %s", resp.Status))
		return resp.StatusCode, err
	}

	// Grab the Vault init request response from HTTP body
	initRequestResponseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to fetch the Vault init request response body: %s", err.Error()))
		return 0, err
	}

	// Build a JSON structure from the init request response HTTP body
	var initResponse InitResponse
	if err = json.Unmarshal(initRequestResponseBody, &initResponse); err != nil {
		lc.Error(fmt.Sprintf("Failed to build the JSON structure from the init request response body: %s", err.Error()))
		return 0, err
	}

	if debug {
		lc.Info(fmt.Sprintf("Vault Init Response: %s", initRequestResponseBody))
	}

	// Save the JSON structure to a file system JSON file
	err = ioutil.WriteFile(config.SecretService.TokenFolderPath+"/"+config.SecretService.VaultInitParm, initRequestResponseBody, 0600)
	if err != nil {
		lc.Error(fmt.Sprintf("Fatal error creating Vault init response %s file, HTTP status: %s", config.SecretService.TokenFolderPath+"/"+config.SecretService.VaultInitParm, err.Error()))
		return 0, err
	}

	lc.Info("Vault Initialization complete.")

	return resp.StatusCode, nil
}

func VaultUnseal(config *tomlConfig, httpClient *http.Client, debug bool) (sCode int, err error) {

	lc.Info(fmt.Sprintf("Vault Unsealing Process. Applying key shares."))

	// Get the resp-init.json file to fetch the Vault key shares
	var initResponse InitResponse
	rawBytes, err := ioutil.ReadFile(config.SecretService.TokenFolderPath + "/" + config.SecretService.VaultInitParm)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to read the Vault JSON response init file: %s", err.Error()))
		return 0, err
	}

	// Build the JSON structure from the file informations
	if err = json.Unmarshal(rawBytes, &initResponse); err != nil {
		lc.Error(fmt.Sprintf("Failed to build the JSON structure from the init response body: %s", err.Error()))
		return 0, err
	}

	// Build Vault API full URL
	url, err := url.Parse(config.SecretService.Scheme + "://" + config.SecretService.Server + ":" + config.SecretService.Port + vaultUnsealAPI)

	// Iterate the JSON init response key array (key shares) and build/send a unseal request each time
	keyCounter := 1
	for _, key := range initResponse.KeysBase64 {

		// Key share n to apply during the Vault unseal process (until threshold reached)
		unsealRequest := UnsealRequest{
			Key: key,
		}

		if debug {
			lc.Info(fmt.Sprintf("Vault Key Share to apply (SSS): %s", key))
		}

		// Build Vault HTTP/POST Request
		jsonBlock, err := json.Marshal(&unsealRequest) //jsonBlock := []byte(`{}`)
		if err != nil {
			lc.Error(fmt.Sprintf("Failed to build the Vault unseal request (key shares parameter): %s", err.Error()))
			return 0, err
		}
		req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(jsonBlock))

		// Prepare the Header
		req.Header.Set("Content-Type", contentType)

		// POST the request
		resp, err := httpClient.Do(req)
		if err != nil {
			lc.Error(fmt.Sprintf("Failed to send the Vault init request: %s", err.Error()))
			return 0, err
		}
		defer resp.Body.Close()

		// Unseal request OK/KO ?
		if resp.StatusCode != http.StatusOK {
			lc.Error(fmt.Sprintf("Vault unseal request failed with status code: %s", resp.Status))
			return resp.StatusCode, err
		}

		// Grab the Vault unseal request response from HTTP body
		unsealRequestResponseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			lc.Error(fmt.Sprintf("Failed to fetch the Vault unseal request response body: %s", err.Error()))
			return 0, err
		}

		// Build a JSON structure from the unseal request response HTTP body
		var unsealResponse UnsealResponse
		if err = json.Unmarshal(unsealRequestResponseBody, &unsealResponse); err != nil {
			lc.Error(fmt.Sprintf("Failed to build the JSON structure from the unseal request response body: %s", err.Error()))
			return 0, err
		}

		lc.Info(fmt.Sprintf("Vault Key Share %d/%d successfully applied.", keyCounter, config.SecretService.VaultSecretShares))

		// Check if unsealing threshold has been successfully reached?
		if !unsealResponse.Sealed {
			lc.Info("Vault Key Share Threshold Reached. Unsealing complete.")
			return resp.StatusCode, nil
		}

		keyCounter++
	}

	// Unattented Vault unsealing failure
	return 0, fmt.Errorf("%d", 1)
}

// ----------------------------------------------------------
/*
 curl --header "X-Vault-Token: ${_ROOT_TOKEN}" \
            --header "Content-Type: application/json" \
            --request POST \
            --data @${_PAYLOAD_KONG} \
            http://localhost:8200/v1/secret/edgex/pki/tls/edgex-kong
*/
func UploadProxyCerts(config *tomlConfig, secretBaseURL string, cert string, sk string, c *http.Client) (bool, error) {
	body := &CertKeyPair{
		Cert: cert,
		Key:  sk,
	}

	t, err := GetSecret(config.SecretService.TokenFolderPath + "/" + config.SecretService.VaultInitParm)
	if err != nil {
		lc.Error(err.Error())
		return false, err
	}
	lc.Info("Trying to upload API Gateway TLS certificate and key to the secret store.")
	s := sling.New().Set(VaultToken, t.Token)
	req, err := s.New().Base(secretBaseURL).Post(config.SecretService.CertPath).BodyJSON(body).Request()
	resp, err := c.Do(req)
	if err != nil {
		lc.Error("Failed to upload API Gateway TLS certificate and key to secret store: %s", err.Error())
		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent {
		lc.Info("API Gateway TLS certificate and key successfully loaded in the secret store.")
	} else {
		b, _ := ioutil.ReadAll(resp.Body)
		s := fmt.Sprintf("Failed to load the TLS certificate and key to the secret store: %s,%s.", resp.Status, string(b))
		lc.Error(s)
		return false, errors.New(s)
	}
	return true, nil
}
