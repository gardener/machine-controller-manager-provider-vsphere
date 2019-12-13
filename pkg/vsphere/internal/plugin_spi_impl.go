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
	"net/url"
	"strings"

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// PluginSPIImpl is the real implementation of PluginSPI interface
// that makes the calls to the provider SDK
type PluginSPIImpl struct{}

// CreateMachine creates a VM by cloning from a template
func (spi *PluginSPIImpl) CreateMachine(ctx context.Context, machineName string, providerSpec *api.VsphereProviderSpec, secrets *api.Secrets) (string, error) {
	client, err := createVsphereClient(ctx, secrets)
	if err != nil {
		return "", err
	}
	defer client.Logout(ctx)

	cmd := newClone(machineName, providerSpec, secrets.UserData)
	err = cmd.Run(ctx, client)
	if err != nil {
		return "", err
	}
	machineID := cmd.Clone.UUID(ctx)
	return machineID, nil
}

// DeleteMachine deletes a VM by name
func (spi *PluginSPIImpl) DeleteMachine(ctx context.Context, machineName string, providerSpec *api.VsphereProviderSpec, secrets *api.Secrets) (string, error) {
	client, err := createVsphereClient(ctx, secrets)
	if err != nil {
		return "", err
	}
	defer client.Logout(ctx)

	machineID, err := deleteVM(ctx, client, providerSpec, machineName)
	if err != nil {
		return "", err
	}

	return machineID, nil
}

// ShutDownMachine shuts down a machine by name
func (spi *PluginSPIImpl) ShutDownMachine(ctx context.Context, machineName string, providerSpec *api.VsphereProviderSpec, secrets *api.Secrets) (string, error) {
	client, err := createVsphereClient(ctx, secrets)
	if err != nil {
		return "", err
	}
	defer client.Logout(ctx)

	machineID, err := shutDownVM(ctx, client, providerSpec, machineName)
	if err != nil {
		return "", err
	}

	return machineID, nil
}

// GetMachineStatus checks for existence of VM by name
func (spi *PluginSPIImpl) GetMachineStatus(ctx context.Context, machineName string, providerSpec *api.VsphereProviderSpec, secrets *api.Secrets) (string, error) {
	client, err := createVsphereClient(ctx, secrets)
	if err != nil {
		return "", err
	}
	defer client.Logout(ctx)

	vm, err := findByIPath(ctx, client, providerSpec, machineName)
	if err != nil {
		return "", err
	}

	machineID := vm.UUID(ctx)
	return machineID, nil
}

// ListMachines lists all VMs in the DC or folder
func (spi *PluginSPIImpl) ListMachines(ctx context.Context, providerSpec *api.VsphereProviderSpec, secrets *api.Secrets) (map[string]string, error) {
	client, err := createVsphereClient(ctx, secrets)
	if err != nil {
		return nil, err
	}
	defer client.Logout(ctx)

	machineList := map[string]string{}

	clusterName := ""
	nodeRole := ""
	for key := range providerSpec.Tags {
		if strings.HasPrefix(key, "kubernetes.io/cluster/") {
			clusterName = key
		} else if strings.HasPrefix(key, "kubernetes.io/role/") {
			nodeRole = key
		}
	}

	if clusterName == "" || nodeRole == "" {
		return machineList, nil
	}

	visitor := func(vm *object.VirtualMachine, obj mo.ManagedEntity, field object.CustomFieldDefList) error {
		matchedCluster := false
		matchedRole := false
		for _, cv := range obj.CustomValue {
			sv := cv.(*types.CustomFieldStringValue)
			switch field.ByKey(sv.Key).Name {
			case clusterName:
				matchedCluster = true
			case nodeRole:
				matchedRole = true
			}
		}
		if matchedCluster && matchedRole {
			uuid := vm.UUID(ctx)
			machineList[uuid] = obj.Name
		}
		return nil
	}

	err = visitVirtualMachines(ctx, client, providerSpec, visitor)
	if err != nil {
		return nil, err
	}

	return machineList, nil
}

func createVsphereClient(ctx context.Context, secret *api.Secrets) (*govmomi.Client, error) {
	clientURL, err := url.Parse("https://" + secret.VsphereHost + "/sdk")
	if err != nil {
		return nil, err
	}

	clientURL.User = url.UserPassword(secret.VsphereUsername, secret.VspherePassword)

	// Connect and log in to ESX or vCenter
	return govmomi.NewClient(ctx, clientURL, secret.VsphereInsecureSSL)
}
