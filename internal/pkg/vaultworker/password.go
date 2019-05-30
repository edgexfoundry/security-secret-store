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

	"github.com/dghubble/sling"
	"github.com/sethvargo/go-password/password"
)

type UserPasswd struct {
    User string
    Passwd string
}

func CreateCredential() (string, error) {
	pass, err := password.Generate(8, 4, 4, false, false)
	if err != nil {
		return "", err
	}
	return pass, nil
}

func CredentialInStore(config *tomlConfig, secretBaseURL string, credPath string, c *http.Client) (bool, error) {

	t, err := GetSecret(config.SecretService.TokenFolderPath + "/" + config.SecretService.VaultInitParm)
	if err != nil {
		lc.Error(err.Error())
		return false, err
	}

	s := sling.New().Set(VaultToken, t.Token)

	req, err := s.New().Base(secretBaseURL).Get(credPath).Request()
	resp, err := c.Do(req)
	if err != nil {
		errStr := fmt.Sprintf("Failed to retrieve credentials with path as %s with error %s", credPath, err.Error())
		return false, errors.New(errStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		lc.Info(fmt.Sprintf("No credential path found, setting up the credentials for %s", credPath))
		return false, nil
	}

	lc.Info(fmt.Sprintf("%s - %d", credPath, resp.StatusCode))

	var result map[string]interface{}

	by, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(by, &result)

	credentials := result["data"].(map[string]interface{})

	if len(credentials) > 0 {
		return true, nil
	}
	return false, nil
}

func InitCredentials(config *tomlConfig, secretBaseURL string, secretPath string, cred *UserPasswd, c *http.Client) error {
	
	t, err := GetSecret(config.SecretService.TokenFolderPath + "/" + config.SecretService.VaultInitParm)
	if err != nil {
		lc.Error(err.Error())
		return err
	}

	s := sling.New().Set(VaultToken, t.Token)

	lc.Info("Trying to upload init credentials to secret service server.")
	req, err := s.New().Base(secretBaseURL).Post(secretPath).BodyJSON(cred).Request()
	resp, err := c.Do(req)
	if err != nil {
		lc.Error("Failed to upload init credentials to secret with error %s", err.Error())
		return err
	}

	defer resp.Body.Close()

	lc.Info(fmt.Sprintf("%s - %d", secretPath, resp.StatusCode))

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info(fmt.Sprintf("Successful to add init credentials to secret service with path %s.", secretPath))
	} else {
		s := fmt.Sprintf("Failed to add init credentials on path %s with errorcode %d.", secretPath, resp.StatusCode)
		return errors.New(s)
	}
	return nil
}
