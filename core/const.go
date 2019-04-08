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
package main

// Global constants
const (
	CertificatesPath = "certificates/"
	SecurityService  = "securityservice"
	EdgeXService     = "edgex"
	VaultToken       = "X-Vault-Token"

	// Vault API endpoints: v1
	vaultHealthAPI      = "/v1/sys/health"
	vaultInitAPI        = "/v1/sys/init"
	vaultUnsealAPI      = "/v1/sys/unseal"
	vaultPolicyAPI      = "/v1/sys/policy/"
	vaultTokenCreateAPI = "/v1/auth/token/create"
	vaultTokenDeleteAPI = "/v1/auth/token/delete"

	vaultDefaultPolicy = "default"
	vaultTokenTTL      = "168h"
	// Vault Configuration defaults/limits: local.hcl
	// If create token w/o ttl then the default will be default_lease_ttl="168h" (7 days)
	// If specified the ttl cannot exceed max_lease_ttl="720h" (30 days)

	tokenFileSuffix = "-token.json"      // When saving service tokens to filesystem
	contentType     = "application/json" // Vault API requests
)
