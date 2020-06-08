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

package internal

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	corev1 "k8s.io/api/core/v1"

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis/tags"
)

// PluginSPIImpl is the real implementation of PluginSPI interface
// that makes the calls to the provider SDK
type PluginSPIImpl struct{}

const providerPrefix = "vsphere://"

// CreateMachine creates a VM by cloning from a template
func (spi *PluginSPIImpl) CreateMachine(ctx context.Context, machineName string, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	client, err := createVsphereClient(ctx, secrets)
	if err != nil {
		return "", err
	}
	defer client.Logout(ctx)

	cmd := newClone(machineName, providerSpec, string(secrets.Data["userData"]))
	err = cmd.Run(ctx, client)
	if err != nil {
		return "", err
	}
	machineID := cmd.Clone.UUID(ctx)
	providerID := spi.encodeProviderID(providerSpec.Region, machineID)
	return providerID, nil
}

func (spi *PluginSPIImpl) encodeProviderID(region, machineID string) string {
	if machineID == "" {
		return ""
	}
	return fmt.Sprintf("%s%s/%s", providerPrefix, region, machineID)
}

func (spi *PluginSPIImpl) decodeProviderID(providerID string) (region, machineID string) {
	if !strings.HasPrefix(providerID, providerPrefix) {
		return
	}
	parts := strings.Split(providerID[len(providerPrefix):], "/")
	if len(parts) != 2 {
		return
	}
	region = parts[0]
	machineID = parts[1]
	return
}

// DeleteMachine deletes a VM by name
func (spi *PluginSPIImpl) DeleteMachine(ctx context.Context, machineName string, providerID string, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	client, err := createVsphereClient(ctx, secrets)
	if err != nil {
		return "", err
	}
	defer client.Logout(ctx)

	_, machineID := spi.decodeProviderID(providerID)
	foundMachineID, err := deleteVM(ctx, client, providerSpec, machineName, machineID)
	if err != nil {
		return "", err
	}

	foundProviderID := spi.encodeProviderID(providerSpec.Region, foundMachineID)
	return foundProviderID, nil
}

// ShutDownMachine shuts down a machine by name
func (spi *PluginSPIImpl) ShutDownMachine(ctx context.Context, machineName string, providerID string, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	client, err := createVsphereClient(ctx, secrets)
	if err != nil {
		return "", err
	}
	defer client.Logout(ctx)

	_, machineID := spi.decodeProviderID(providerID)
	foundMachineID, err := shutDownVM(ctx, client, providerSpec, machineName, machineID)
	if err != nil {
		return "", err
	}

	foundProviderID := spi.encodeProviderID(providerSpec.Region, foundMachineID)
	return foundProviderID, nil
}

// GetMachineStatus checks for existence of VM by name
func (spi *PluginSPIImpl) GetMachineStatus(ctx context.Context, machineName string, providerID string, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	client, err := createVsphereClient(ctx, secrets)
	if err != nil {
		return "", err
	}
	defer client.Logout(ctx)

	_, machineID := spi.decodeProviderID(providerID)
	vm, err := findVM(ctx, client, providerSpec, machineName, machineID)
	if err != nil {
		return "", err
	}

	foundMachineID := vm.UUID(ctx)

	foundProviderID := spi.encodeProviderID(providerSpec.Region, foundMachineID)
	return foundProviderID, nil
}

// ListMachines lists all VMs in the DC or folder
func (spi *PluginSPIImpl) ListMachines(ctx context.Context, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (map[string]string, error) {
	client, err := createVsphereClient(ctx, secrets)
	if err != nil {
		return nil, err
	}
	defer client.Logout(ctx)

	machineList := map[string]string{}

	relevantTags, _ := tags.NewRelevantTags(providerSpec.Tags)
	if relevantTags == nil {
		return machineList, nil
	}

	visitor := func(vm *object.VirtualMachine, obj mo.ManagedEntity, field object.CustomFieldDefList) error {
		customValues := map[string]string{}
		for _, cv := range obj.CustomValue {
			sv := cv.(*types.CustomFieldStringValue)
			customValues[field.ByKey(sv.Key).Name] = sv.Value
		}

		if relevantTags.Matches(customValues) {
			uuid := vm.UUID(ctx)
			providerID := spi.encodeProviderID(providerSpec.Region, uuid)
			machineList[providerID] = obj.Name
		}
		return nil
	}

	err = visitVirtualMachines(ctx, client, providerSpec, visitor)
	if err != nil {
		return nil, err
	}

	return machineList, nil
}

func createVsphereClient(ctx context.Context, secret *corev1.Secret) (*govmomi.Client, error) {
	clientURL, err := url.Parse("https://" + string(secret.Data["vsphereHost"]) + "/sdk")
	if err != nil {
		return nil, err
	}

	clientURL.User = url.UserPassword(string(secret.Data["vsphereUsername"]), string(secret.Data["vspherePassword"]))

	// Connect and log in to ESX or vCenter
	vsphereInsecureSSL := strings.ToLower(string(secret.Data["vsphereInsecureSSL"])) == "true" || string(secret.Data["vsphereInsecureSSL"]) == "1"
	return govmomi.NewClient(ctx, clientURL, vsphereInsecureSSL)
}
