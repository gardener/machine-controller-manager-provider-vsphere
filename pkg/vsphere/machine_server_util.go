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

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis/validation"
	errors2 "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/errors"
	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// decodeProviderSpecAndSecret converts request parameters to api.VsphereProviderSpec
func decodeProviderSpecAndSecret(machineClass *v1alpha1.MachineClass, secret *corev1.Secret) (api.VsphereProviderSpec, error) {
	if _, ok := secret.Data[api.VSphereKubeconfig]; ok {
		spec, err := decodeProviderSpec2AndSecret(machineClass, secret)
		return spec, err
	}
	spec, err := decodeProviderSpec1AndSecret(machineClass, secret)
	return spec, err
}

// decodeProviderSpec1AndSecret converts request parameters to api.VsphereProviderSpec
func decodeProviderSpec1AndSecret(machineClass *v1alpha1.MachineClass, secret *corev1.Secret) (*api.VsphereProviderSpec1, error) {
	var (
		providerSpec *api.VsphereProviderSpec1
	)

	// Extract providerSpec
	err := json.Unmarshal(machineClass.ProviderSpec.Raw, &providerSpec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	//Validate the Spec and Secrets
	ValidationErr := validation.ValidateVsphereProviderSpec1(providerSpec, secret)
	if ValidationErr != nil {
		err = fmt.Errorf("Error while validating ProviderSpec %v", ValidationErr)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return providerSpec, nil
}

// decodeProviderSpec2AndSecret converts request parameters to api.VsphereProviderSpec2
func decodeProviderSpec2AndSecret(machineClass *v1alpha1.MachineClass, secret *corev1.Secret) (*api.VsphereProviderSpec2, error) {
	var (
		providerSpec *api.VsphereProviderSpec2
	)

	// Extract providerSpec
	err := json.Unmarshal(machineClass.ProviderSpec.Raw, &providerSpec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	//Validate the Spec and Secrets
	ValidationErr := validation.ValidateVsphereProviderSpec2(providerSpec, secret)
	if ValidationErr != nil {
		err = fmt.Errorf("Error while validating ProviderSpec2 %v", ValidationErr)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return providerSpec, nil
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
	klog.V(2).Infof(wrapped.Error())
	return status.Error(code, wrapped.Error())
}
