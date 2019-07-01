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

package vaultworker

import (
	"fmt"
	"net/http"
)

type CredInfo struct {
	Path string
	Pair *UserPasswd
}

/* Need to creat credentials for all microservices and put them into the path of secret service
 */

func CredentialsInit(config *tomlConfig, secretBaseURL string, c *http.Client) error {

	creds := map[string]*CredInfo{}

	adminpasswd, _ := CreateCredential()
	creds["mongo"] = &CredInfo{Path: config.SecretService.MongoSecretPath, Pair: &UserPasswd{User: "admin", Passwd: adminpasswd}}

	corepasswd, _ := CreateCredential()
	creds["coredata"] = &CredInfo{Path: config.SecretService.CoredataSecretPath, Pair: &UserPasswd{User: "core", Passwd: corepasswd}}

	metapasswd, _ := CreateCredential()
	creds["metadata"] = &CredInfo{Path: config.SecretService.MetadataSecretPath, Pair: &UserPasswd{User: "meta", Passwd: metapasswd}}

	repasswd, _ := CreateCredential()
	creds["rulesengine"] = &CredInfo{Path: config.SecretService.RulesenginesecretPath, Pair: &UserPasswd{User: "rules_engine_user", Passwd: repasswd}}

	ntpasswd, _ := CreateCredential()
	creds["notifications"] = &CredInfo{Path: config.SecretService.NotificationsSecretPath, Pair: &UserPasswd{User: "notifications", Passwd: ntpasswd}}

	scpasswd, _ := CreateCredential()
	creds["scheduler"] = &CredInfo{Path: config.SecretService.SchedulerSecretPath, Pair: &UserPasswd{User: "scheduler", Passwd: scpasswd}}

	lgpasswd, _ := CreateCredential()
	creds["logging"] = &CredInfo{Path: config.SecretService.LoggingSecretPath, Pair: &UserPasswd{User: "logging", Passwd: lgpasswd}}

	for _, v := range creds {
		hasCred, err := CredentialInStore(config, secretBaseURL, v.Path, c)
		if err != nil {
			return err
		}

		if hasCred == true {
			lc.Info(fmt.Sprintf("Credential initialization parameters are in the secret store with path %s already. Skip creating this credential.", v.Path))
		} else {
			err = InitCredentials(config, secretBaseURL, v.Path, v.Pair, c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
