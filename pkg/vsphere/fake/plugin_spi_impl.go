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

package fake

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
)

// PluginSPIImpl is the fake implementation of PluginSPI interface
// that makes the calls to the provider SDK
type PluginSPIImpl struct{}

// CreateMachine creates a VM by cloning from a template
func (spi *PluginSPIImpl) CreateMachine(ctx context.Context, machineName string, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	return "", fmt.Errorf("fake not implemented yet")
}

// DeleteMachine deletes a VM by name
func (spi *PluginSPIImpl) DeleteMachine(ctx context.Context, machineName string, providerID string, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	return "", fmt.Errorf("fake not implemented yet")
}

// ShutDownMachine shuts down a machine by name
func (spi *PluginSPIImpl) ShutDownMachine(ctx context.Context, machineName string, providerID string, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	return "", fmt.Errorf("fake not implemented yet")
}

// GetMachineStatus checks for existence of VM by name
func (spi *PluginSPIImpl) GetMachineStatus(ctx context.Context, machineName string, providerID string, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	return "", fmt.Errorf("fake not implemented yet")
}

// ListMachines lists all VMs in the DC or folder
func (spi *PluginSPIImpl) ListMachines(ctx context.Context, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (map[string]string, error) {
	return nil, fmt.Errorf("fake not implemented yet")
}
