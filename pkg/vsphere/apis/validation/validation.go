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

// Package validation is used to validate all the machine CRD objects
package validation

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/machine-controller-manager/pkg/apis/machine"
)

// ValidateVsphereMachineClass validates a VsphereMachineClass and returns a list of errors.
func ValidateVsphereMachineClass(VsphereMachineClass *machine.VsphereMachineClass) field.ErrorList {
	return internalValidateVsphereMachineClass(VsphereMachineClass)
}

func internalValidateVsphereMachineClass(VsphereMachineClass *machine.VsphereMachineClass) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateVsphereMachineClassSpec(&VsphereMachineClass.Spec, field.NewPath("spec"))...)
	return allErrs
}

func validateVsphereMachineClassSpec(spec *machine.VsphereMachineClassSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if "" == spec.Datastore && "" == spec.DatastoreCluster {
		allErrs = append(allErrs, field.Required(fldPath.Child("datastoreCluster"), "DatastoreCluster or Datastore is required"))
	}
	if "" == spec.TemplateVM {
		allErrs = append(allErrs, field.Required(fldPath.Child("templateVM"), "TemplateVM is required"))
	}
	if "" == spec.ComputeCluster && "" == spec.Pool && "" == spec.HostSystem {
		allErrs = append(allErrs, field.Required(fldPath.Child("computeCluster"), "ComputeCluster or Pool or HostSystem is required"))
	}
	if "" == spec.Network {
		allErrs = append(allErrs, field.Required(fldPath.Child("network"), "Network is required"))
	}
	// TODO martin: complete VsphereMachineClassSpec validation

	allErrs = append(allErrs, validateSecretRef(spec.SecretRef, field.NewPath("spec.secretRef"))...)
	allErrs = append(allErrs, validateVsphereClassSpecTags(spec.Tags, field.NewPath("spec.tags"))...)

	return allErrs
}

func validateVsphereClassSpecTags(tags map[string]string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	clusterName := ""
	nodeRole := ""

	for key := range tags {
		if strings.Contains(key, "kubernetes.io/cluster/") {
			clusterName = key
		} else if strings.Contains(key, "kubernetes.io/role/") {
			nodeRole = key
		}
	}

	if clusterName == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("kubernetes.io/cluster/"), "Tag required of the form kubernetes.io/cluster/****"))
	}
	if nodeRole == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("kubernetes.io/role/"), "Tag required of the form kubernetes.io/role/****"))
	}

	return allErrs
}
