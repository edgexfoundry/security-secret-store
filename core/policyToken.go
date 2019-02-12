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

  @author: Alain Pulluelo, ForgeRock AS (created: October 19, 2018)
  @version: 0.2.2
*/

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// ----------------------------------------------------------
// Information:
//    https://www.vaultproject.io/api/system/policy.html
//    https://www.vaultproject.io/api/auth/token/index.html
// ----------------------------------------------------------

// TokenData structure to serialize a token create data
/*
	{
		"policies": [
		  "admin",
		  "default"
		],
		"metadata": {
		  "user": "admin user"
		},
		"display_name": "admin",
		"ttl": "1h",
		"renewable": true
	  }
*/
type TokenData struct {
	Policies    []string `json:"policies"`
	Metadata    Metadata `json:"metadata"`
	DisplayName string   `json:"display_name"`
	TTL         string   `json:"ttl"`
	Renewable   string   `json:"renewable"`
}

// Metadata structure from token create data structure
type Metadata struct {
	User string `json:"user"`
}

const (
	vaultPolicyAPI      = "/v1/sys/policy/"
	vaultDefaultPolicy  = "default"
	vaultTokenTTL       = "1h"
	vaultTokenCreateAPI = "/v1/auth/token/create"
	vaultTokenDeleteAPI = "/v1/auth/token/delete"

	tokenFileSuffix = "-token.json"

	contentType = "application/json"
)

// ----------------------------------------------------------
func getPolicyFromFile(policyFilePtr *string) ([]byte, error) {

	var (
		readLine      string
		fullRequest   string
		policyRequest []byte
	)

	// Open JSON policy config file
	hclFile, err := os.Open(*policyFilePtr)
	if err != nil {
		return nil, err
	}
	defer hclFile.Close() // Defers Close till func return

	// Format the request from HCL (add escape char, operation and remove comments)
	/*
		To read a file line-by-line, using bufio.Scanner seems easier.
		And Scanner won't includes endline (\n or \r\n) into the line.
	*/
	scanbuf := bufio.NewScanner(hclFile)
	for scanbuf.Scan() {
		readLine = scanbuf.Text()
		readLine = strings.Replace(readLine, "\"", "\\\"", -1) // escape the double quote
		readLine = strings.TrimSpace(readLine)                 // Trim leading/trailing white spaces
		indexComment := strings.Index(readLine, "#")           // Get the index of a comment char
		if indexComment == -1 || indexComment > 0 {            // Index set to -1 if # not found
			fullRequest = fullRequest + " " + readLine // Discard commented lines
		}
	}
	fullRequest = "{ \"policy\": \"" + fullRequest + "\"}" // Encapsules the request w/ operation

	// Put the request string into a byteArray
	policyRequest = []byte(fullRequest)

	return policyRequest, nil
}

// ----------------------------------------------------------
func importPolicy(policyName string, policyRequest *[]byte, rootToken string, config *tomlConfig, httpClient *http.Client) (err error) {

	// Build Vault API full URL
	url, err := url.Parse(config.SecretService.Scheme + "://" + config.SecretService.Server + ":" + config.SecretService.Port + vaultPolicyAPI + policyName)
	// Build Vault HTTP/POST Request
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(*policyRequest))

	// Prepare the Header
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Vault-Token", rootToken)

	// POST the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		lc.Info("Import Policy Successfull.")
	} else {
		lc.Error(fmt.Sprintf("Import Policy HTTP Status: %s (StatusCode: %s)", resp.Status, strconv.Itoa(resp.StatusCode)))
		return fmt.Errorf("%s", policyName)
	}

	return nil
}

// ----------------------------------------------------------
func createToken(tokenName string, policyName string, rootToken string, config *tomlConfig, httpClient *http.Client) (err error) {

	// Prepare the JSON to be POST'ed
	userData := Metadata{tokenName + " user"}

	tokenData := TokenData{
		Policies:    []string{policyName, vaultDefaultPolicy},
		Metadata:    userData,
		DisplayName: tokenName,
		TTL:         vaultTokenTTL,
		Renewable:   "True",
	}

	tokenDataReq, err := json.Marshal(tokenData)
	if err != nil {
		fatalIfErr(err, "Token data request creation failure")
	}

	// Build Vault API full URL
	url, err := url.Parse(config.SecretService.Scheme + "://" + config.SecretService.Server + ":" + config.SecretService.Port + vaultTokenCreateAPI)
	// Build Vault HTTP/POST Request
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(tokenDataReq))

	// Prepare the Header
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Vault-Token", rootToken)

	// POST the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Get response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fatalIfErr(err, "Read Body failure")
	}

	if resp.StatusCode == 200 {
		lc.Info("Create Token Successfull.")
	} else {
		lc.Error(fmt.Sprintf("Fatal Error Creating Token in Vault, HTTP Status: %s", resp.Status))
		return fmt.Errorf("%d", 1)
	}

	// Save created token data to a JSON file
	err = ioutil.WriteFile(config.SecretService.TokenFolderPath+"/"+tokenName+tokenFileSuffix, body, 0644)
	if err != nil {
		lc.Error(fmt.Sprintf("Fatal Error Writing %s Token in Vault, HTTP Status: %s", tokenName, resp.Status))
		return err
	}

	return nil
}

// ----------------------------------------------------------
func vaultHealthCheck(config *tomlConfig, httpClient *http.Client) (sCode int, err error) {

	// Build Vault API full URL
	url, err := url.Parse(config.SecretService.Scheme + "://" + config.SecretService.Server + ":" + config.SecretService.Port + "/sys/health")
	// Build Vault HTTP/GET Request
	var jsonBlock = []byte(`{}`)
	req, err := http.NewRequest(http.MethodGet, url.String(), bytes.NewBuffer(jsonBlock))

	// Prepare the Header
	req.Header.Set("Content-Type", contentType)

	// GET the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	lc.Info(fmt.Sprintf("Vault Health Check HTTP Status: %s (StatusCode: %s)", resp.Status, strconv.Itoa(resp.StatusCode)))

	return resp.StatusCode, nil
}

// ----------------------------------------------------------
func fatalIfErr(err error, msg string) {
	if err != nil {
		log.Fatalf("ERROR: %s: %s", msg, err) // fatalf() =  Prinf() followed by a call to os.Exit(1)
	}
}
