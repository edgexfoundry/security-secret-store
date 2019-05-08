/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
)

type CoredataCredentials struct {
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
}

func coredataCredentialInStore(config *tomlConfig, secretBaseURL string, credPath string, c *http.Client) (bool, error) {

	return credentialInStore(config, secretBaseURL, credPath, c)
}

func initCoredataCredentials(config *tomlConfig, secretBaseURL string, c *http.Client) error {

	password, _ := createCredential()

	body := &CoredataCredentials{
		Name:     Coredata,
		Password: password,
	}

	t, err := getSecret(config.SecretService.TokenFolderPath + "/" + config.SecretService.VaultInitParm)
	if err != nil {
		lc.Error(err.Error())
		return err
	}

	s := sling.New().Set(VaultToken, t.Token)

	lc.Info("Trying to upload coredata initial credentials to secret service server.")
	req, err := s.New().Base(secretBaseURL).Post(config.SecretService.CoredataSecretPath).BodyJSON(body).Request()
	resp, err := c.Do(req)
	if err != nil {
		lc.Error("Failed to upload initial credentials to secret with error %s", err.Error())
		return err
	}

	defer resp.Body.Close()

	lc.Info(fmt.Sprintf("%s - %d", config.SecretService.CoredataSecretPath, resp.StatusCode))

	if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
		lc.Info("Successful to add Coredata initial credentials to secret service.")
	} else {
		s := fmt.Sprintf("Failed to add Coredata initial credentials with errorcode %d.", resp.StatusCode)
		return errors.New(s)
	}
	return nil
}
