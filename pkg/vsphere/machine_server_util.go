/*
 * Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package vsphere

import (
	"encoding/json"
	"fmt"
	"strings"

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis/validation"
	errors2 "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/errors"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// decodeProviderSpecAndSecret converts request parameters to api.ProviderSpec & api.Secrets
func decodeProviderSpecAndSecret(providerSpecBytes []byte, secretMap map[string][]byte, checkUserData bool) (*api.VsphereProviderSpec, *api.Secrets, error) {
	var (
		providerSpec *api.VsphereProviderSpec
	)

	// Extract providerSpec
	err := json.Unmarshal(providerSpecBytes, &providerSpec)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	// Extract secrets from secretMap
	secrets, err := getSecretsFromSecretMap(secretMap, checkUserData)
	if err != nil {
		return nil, nil, err
	}

	//Validate the Spec and Secrets
	ValidationErr := validation.ValidateVsphereProviderSpec(providerSpec, secrets)
	if ValidationErr != nil {
		err = fmt.Errorf("Error while validating ProviderSpec %v", ValidationErr)
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	return providerSpec, secrets, nil
}

// getSecretsFromSecretMap converts secretMap to api.secrets object
func getSecretsFromSecretMap(secretMap map[string][]byte, checkUserData bool) (*api.Secrets, error) {
	host, hostExists := secretMap["vsphereHost"]
	username, usernameExists := secretMap["vsphereUsername"]
	password, passwordExists := secretMap["vspherePassword"]
	insecureSSL := secretMap["vsphereInsecureSSL"]
	userData, userDataExists := secretMap["userData"]
	missingKeys := []string{}
	if !hostExists {
		missingKeys = append(missingKeys, "vsphereHost")
	}
	if !usernameExists {
		missingKeys = append(missingKeys, "vsphereUsername")
	}
	if !passwordExists {
		missingKeys = append(missingKeys, "vspherePassword")
	}
	if checkUserData && !userDataExists {
		missingKeys = append(missingKeys, "userData")
	}
	if len(missingKeys) > 0 {
		return nil, status.Error(codes.Internal, fmt.Sprintf("invalid secret map. Missing keys: '%s'", strings.Join(missingKeys, "', '")))
	}

	var secrets api.Secrets
	secrets.VsphereHost = string(host)
	secrets.VsphereUsername = string(username)
	secrets.VspherePassword = string(password)
	secrets.VsphereInsecureSSL = strings.ToLower(string(insecureSSL)) == "true" || string(insecureSSL) == "1"
	secrets.UserData = string(userData)

	return &secrets, nil
}

func prepareErrorf(err error, format string, args ...interface{}) error {
	var (
		code    codes.Code
		wrapped error
	)
	switch err.(type) {
	case *errors2.MachineNotFoundError:
		code = codes.NotFound
		wrapped = err
	default:
		code = codes.Internal
		wrapped = errors.Wrap(err, fmt.Sprintf(format, args...))
	}
	glog.V(2).Infof(wrapped.Error())
	return status.Error(code, wrapped.Error())
}
