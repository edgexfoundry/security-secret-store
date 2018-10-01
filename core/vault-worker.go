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
 * @version: 0.1.0
 *******************************************************************************/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/hashicorp/vault/api"
)

func initVault(c *api.Sys, path string, inited bool) (string, error) {

	if inited == false {
		ir := &api.InitRequest{
			SecretShares:    1,
			SecretThreshold: 1,
		}

		resp, err := c.Init(ir)
		r, _ := json.Marshal(resp)
		ioutil.WriteFile(path, r, 0644)
		lc.Info(string(r))
		lc.Info("Vault has been initialized successfully.")

		return resp.KeysB64[0], err
	}
	s, err := getSecret(path)
	if err != nil {
		return "", err
	}
	lc.Info("Vault has been initialized previously. Loading the access token for unsealling.")
	return s.Token, nil
}

func unsealVault(c *api.Sys, token string) (bool, error) {
	if len(token) == 0 {
		return true, errors.New("error:empty token")
	}
	r, err := c.SealStatus()
	if err != nil {
		lc.Error(err.Error())
		return true, err
	}
	if r.Sealed == false {
		lc.Info("Vault is in unseal status, nothing to do.")
		return false, err
	}
	resp, err := c.Unseal(token)
	if err != nil {
		fmt.Println(err.Error())
		return true, err
	}
	return resp.Sealed, err
}

func checkProxyCerts(config *tomlConfig, secretBaseURL string, c *http.Client) (bool, error) {
	cert, key, err := getCertKeyPair(config, secretBaseURL, c)
	if err != nil {
		return false, err
	}
	if len(cert) > 0 && len(key) > 0 {
		return true, nil
	}
	return false, nil
}

/*
 curl --header "X-Vault-Token: ${_ROOT_TOKEN}" \
            --header "Content-Type: application/json" \
            --request POST \
            --data @${_PAYLOAD_KONG} \
            http://localhost:8200/v1/secret/edgex/pki/tls/edgex-kong
*/
func uploadProxyCerts(config *tomlConfig, secretBaseURL string, cert string, sk string, c *http.Client) (bool, error) {
	body := &CertPair{
		Cert: cert,
		Key:  sk,
	}

	t, err := getSecret(config.SecretService.TokenPath)
	if err != nil {
		lc.Error(err.Error())
		return false, err
	}
	lc.Info("Trying to upload cert&key to secret store.")
	s := sling.New().Set(VaultToken, t.Token)
	req, err := s.New().Base(secretBaseURL).Post(config.SecretService.CertPath).BodyJSON(body).Request()
	resp, err := c.Do(req)
	if err != nil {
		lc.Error("Failed to upload cert to secret store with error %s", err.Error())
		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 204 {
		lc.Info("Successful to add certificate to the secret store.")
	} else {
		b, _ := ioutil.ReadAll(resp.Body)
		s := fmt.Sprintf("Failed to add certificate to the secret store with error %s,%s.", resp.Status, string(b))
		lc.Error(s)
		return false, errors.New(s)
	}
	return true, nil
}
