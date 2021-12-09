/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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
 *
 */

package validation

import (
	"fmt"

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis/tags"

	corev1 "k8s.io/api/core/v1"
)

// ValidateVsphereProviderSpec1 validates Vsphere provider spec
func ValidateVsphereProviderSpec1(spec *api.VsphereProviderSpec, secrets *corev1.Secret) []error {
	var allErrs []error

	v1 := spec.V1
	if "" == v1.Datastore && "" == v1.DatastoreCluster {
		allErrs = append(allErrs, fmt.Errorf("either datastoreCluster or datastore field is required"))
	}
	if "" == v1.TemplateVM {
		allErrs = append(allErrs, fmt.Errorf("templateVM is a required field"))
	}
	if "" == v1.ComputeCluster && "" == v1.ResourcePool && "" == v1.HostSystem {
		allErrs = append(allErrs, fmt.Errorf("either computeCluster or resourcePool or hostSystem field is required"))
	}
	if "" == v1.Network {
		allErrs = append(allErrs, fmt.Errorf("network is a required field"))
	}

	allErrs = append(allErrs, validateSecrets(secrets)...)
	_, tagErrs := tags.NewRelevantTags(spec.Tags)
	allErrs = append(allErrs, tagErrs...)

	return allErrs
}

func validateSecrets(secret *corev1.Secret) []error {
	var allErrs []error

	if secret == nil {
		allErrs = append(allErrs, fmt.Errorf("Secret object that has been passed by the MCM is nil"))
	} else {
		_, hostExists := secret.Data["vsphereHost"]
		_, usernameExists := secret.Data["vsphereUsername"]
		_, passwordExists := secret.Data["vspherePassword"]
		_, userDataExists := secret.Data["userData"]

		if !hostExists {
			allErrs = append(allErrs, fmt.Errorf("Secret vsphereHost is required field"))
		}
		if !usernameExists {
			allErrs = append(allErrs, fmt.Errorf("Secret vsphereUsername is required field"))
		}
		if !passwordExists {
			allErrs = append(allErrs, fmt.Errorf("Secret vspherePassword is required field"))
		}
		if !userDataExists {
			allErrs = append(allErrs, fmt.Errorf("Secret userData is required field"))
		}
	}

	return allErrs
}

// ValidateVsphereProviderSpec2 validates Vsphere provider spec2
func ValidateVsphereProviderSpec2(spec *api.VsphereProviderSpec, secrets *corev1.Secret) []error {
	var allErrs []error

	v2 := spec.V2
	if "" == v2.Namespace {
		allErrs = append(allErrs, fmt.Errorf("namespace is a required field"))
	}
	if "" == v2.ImageName {
		allErrs = append(allErrs, fmt.Errorf("imageName is a required field"))
	}
	if "" == v2.NetworkType {
		allErrs = append(allErrs, fmt.Errorf("networkType is a required field"))
	}
	if "" == v2.NetworkName {
		allErrs = append(allErrs, fmt.Errorf("networkName is a required field"))
	}
	if "" == v2.ClassName {
		allErrs = append(allErrs, fmt.Errorf("className is a required field"))
	}

	allErrs = append(allErrs, validateSecrets2(secrets)...)
	_, tagErrs := tags.NewRelevantTags(spec.Tags)
	allErrs = append(allErrs, tagErrs...)

	return allErrs
}

func validateSecrets2(secret *corev1.Secret) []error {
	var allErrs []error

	if secret == nil {
		allErrs = append(allErrs, fmt.Errorf("Secret object that has been passed by the MCM is nil"))
	} else {
		_, kubeconfigExists := secret.Data[api.VSphereKubeconfig]
		_, userDataExists := secret.Data["userData"]

		if !kubeconfigExists {
			allErrs = append(allErrs, fmt.Errorf("Secret %s is required field", api.VSphereKubeconfig))
		}
		if !userDataExists {
			allErrs = append(allErrs, fmt.Errorf("Secret userData is required field"))
		}
	}

	return allErrs
}
