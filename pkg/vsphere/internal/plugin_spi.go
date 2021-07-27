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

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/internal/vmomi"
	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/internal/vmop"
	corev1 "k8s.io/api/core/v1"
)

// PluginSPISwitch is the real implementation of PluginSPI interface
// that makes the calls to the provider SDK
type PluginSPISwitch struct {
	spec1 *vmomi.PluginSPIImpl
	spec2 *vmop.PluginSPIImpl
}

// NewPluginSPISwitch creates a new PluginSPISwitch
func NewPluginSPISwitch() *PluginSPISwitch {
	return &PluginSPISwitch{
		spec1: &vmomi.PluginSPIImpl{},
		spec2: &vmop.PluginSPIImpl{},
	}
}

// CreateMachine creates a VM by cloning from a template
func (spi *PluginSPISwitch) CreateMachine(ctx context.Context, machineName string, providerSpec api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	switch providerSpec.SpecVersion() {
	case 1:
		return spi.spec1.CreateMachine(ctx, machineName, providerSpec.(*api.VsphereProviderSpec1), secrets)
	case 2:
		return spi.spec2.CreateMachine(ctx, machineName, providerSpec.(*api.VsphereProviderSpec2), secrets)
	default:
		return "", fmt.Errorf("invalid spec version")
	}
}

// DeleteMachine deletes a VM by name
func (spi *PluginSPISwitch) DeleteMachine(ctx context.Context, machineName string, providerID string, providerSpec api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	switch providerSpec.SpecVersion() {
	case 1:
		return spi.spec1.DeleteMachine(ctx, machineName, providerID, providerSpec.(*api.VsphereProviderSpec1), secrets)
	case 2:
		return spi.spec2.DeleteMachine(ctx, machineName, providerID, providerSpec.(*api.VsphereProviderSpec2), secrets)
	default:
		return "", fmt.Errorf("invalid spec version")
	}
}

// ShutDownMachine shuts down a machine by name
func (spi *PluginSPISwitch) ShutDownMachine(ctx context.Context, machineName string, providerID string, providerSpec api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	switch providerSpec.SpecVersion() {
	case 1:
		return spi.spec1.ShutDownMachine(ctx, machineName, providerID, providerSpec.(*api.VsphereProviderSpec1), secrets)
	case 2:
		return spi.spec2.ShutDownMachine(ctx, machineName, providerID, providerSpec.(*api.VsphereProviderSpec2), secrets)
	default:
		return "", fmt.Errorf("invalid spec version")
	}
}

// GetMachineStatus checks for existence of VM by name
func (spi *PluginSPISwitch) GetMachineStatus(ctx context.Context, machineName string, providerID string, providerSpec api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	switch providerSpec.SpecVersion() {
	case 1:
		return spi.spec1.GetMachineStatus(ctx, machineName, providerID, providerSpec.(*api.VsphereProviderSpec1), secrets)
	case 2:
		return spi.spec2.GetMachineStatus(ctx, machineName, providerID, providerSpec.(*api.VsphereProviderSpec2), secrets)
	default:
		return "", fmt.Errorf("invalid spec version")
	}
}

// ListMachines lists all VMs in the DC or folder
func (spi *PluginSPISwitch) ListMachines(ctx context.Context, providerSpec api.VsphereProviderSpec, secrets *corev1.Secret) (map[string]string, error) {
	switch providerSpec.SpecVersion() {
	case 1:
		return spi.spec1.ListMachines(ctx, providerSpec.(*api.VsphereProviderSpec1), secrets)
	case 2:
		return spi.spec2.ListMachines(ctx, providerSpec.(*api.VsphereProviderSpec2), secrets)
	default:
		return nil, fmt.Errorf("invalid spec version")
	}
}
