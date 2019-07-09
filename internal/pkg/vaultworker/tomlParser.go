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
	"github.com/BurntSushi/toml"
)

type tomlConfig struct {
	Title         string
	SecretService secretservice
}

type secretservice struct {
	Scheme                  string
	Server                  string
	Port                    string
	CAFilePath              string
	CertPath                string
	CertFilePath            string
	KeyFilePath             string
	VaultInitParm           string
	VaultSecretShares       int
	VaultSecretThreshold    int
	TokenFolderPath         string
	PolicyPath4Admin        string
	PolicyName4Admin        string
	TokenName4Admin         string
	PolicyPath4Kong         string
	PolicyName4Kong         string
	TokenName4Kong          string
	SNIS                    string
}

// LoadTomlConfig Loading the TOML configuration into structure
func LoadTomlConfig(path string) (*tomlConfig, error) {
	config := tomlConfig{}
	_, err := toml.DecodeFile(path, &config)
	return &config, err
}
