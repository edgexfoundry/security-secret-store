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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dghubble/sling"
)

type MongoCredentials struct {
	AdminUser           string `json:"admin,omitempty"`
	AdminPasswd         string `json:"adminpasswd,omitempty"`
	MetadataUser        string `json:"metadata,omitempty"`
	MetadataPasswd      string `json:"metadatapasswd,omitempty"`
	CoredataUser        string `json:"coredata,omitempty"`
	CoredataPasswd      string `json:"coredatapasswd,omitempty"`
	RulesengineUser     string `json:"rulesengine,omitempty"`
	RulesenginePasswd   string `json:"rulesenginepasswd,omitempty"`
	NotificationsUser   string `json:"notifications,omitempty"`
	NotificationsPasswd string `json:"notificationspasswd,omitempty"`
	SchedulerUser       string `json:"scheduler,omitempty"`
	SchedulerPasswd     string `json:"schedulerpasswd,omitempty"`
	LoggingUser         string `json:"logging,omitempty"`
	LoggingPasswd       string `json:"loggingpasswd,omitempty"`
}

func credentialInStore(config *tomlConfig, secretBaseURL string, credPath string, c *http.Client) (bool, error) {

	t, err := getSecret(config.SecretService.TokenFolderPath + "/" + config.SecretService.VaultInitParm)
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

	var result map[string]interface{}

	by, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(by, &result)

	credentials := result["data"].(map[string]interface{})

	if len(credentials) > 0 {
		return true, nil
	}
	return false, nil
}

func initMongoDBCredentials(config *tomlConfig, secretBaseURL string, c *http.Client) error {

	adminpasswd, _ := createCredential()
	metadatapasswd, _ := createCredential()
	coreadatapasswd, _ := createCredential()
	rulesenginepasswd, _ := createCredential()
	notificationspasswd, _ := createCredential()
	schedulerpasswd, _ := createCredential()
	loggingpasswd, _ := createCredential()

	body := &MongoCredentials{
		AdminUser:           "admin",
		AdminPasswd:         adminpasswd,
		MetadataUser:        "metadata",
		MetadataPasswd:      metadatapasswd,
		CoredataUser:        "coredata",
		CoredataPasswd:      coreadatapasswd,
		RulesengineUser:     "rules_engine_user",
		RulesenginePasswd:   rulesenginepasswd,
		NotificationsUser:   "notifications",
		NotificationsPasswd: notificationspasswd,
		SchedulerUser:       "scheduler",
		SchedulerPasswd:     schedulerpasswd,
		LoggingUser:         "logging",
		LoggingPasswd:       loggingpasswd,
	}

	t, err := getSecret(config.SecretService.TokenFolderPath + "/" + config.SecretService.VaultInitParm)
	if err != nil {
		lc.Error(err.Error())
		return err
	}

	s := sling.New().Set(VaultToken, t.Token)

	lc.Info("Trying to upload mongoDBinit credentials to secret service server.")
	req, err := s.New().Base(secretBaseURL).Post(config.SecretService.MongodbinitSecretPath).BodyJSON(body).Request()
	resp, err := c.Do(req)
	if err != nil {
		lc.Error("Failed to upload mongoDBinit credentials to secret with error %s", err.Error())
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
		lc.Info("Successful to add mongoDBinit credentials to secret service.")
	} else {
		s := fmt.Sprintf("Failed to add mongoDBinit credentials with errorcode %d.", resp.StatusCode)
		return errors.New(s)
	}
	return nil
}
