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

package integration

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/errors"
	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/internal"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// TODO: Update secret field from api.Secrets to corev1.Secret in integration tests

type integrationConfig struct {
	MachineName  string                   `json:"machineName"`
	ProviderSpec *api.VsphereProviderSpec `json:"providerSpec"`
	Secrets      *corev1.Secret           `json:"secrets"`
}

// TestPluginSPIImpl tests creation and deleting of a VM via vSphere API.
// Path to configuration needs to be specified as environment variable MCM_PROVIDER_VSPHERE_CONFIG.
func TestPluginSPIImpl(t *testing.T) {
	configPath := os.Getenv("MCM_PROVIDER_VSPHERE_CONFIG")
	if configPath == "" {
		t.Skipf("No path to integrationConfig specified by environmental variable MCM_PROVIDER_VSPHERE_CONFIG")
		return
	}

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Errorf("reading integrationConfig from %s failed with %s", configPath, err)
		return
	}

	cfg := integrationConfig{}
	err = yaml.Unmarshal([]byte(content), &cfg)
	if err != nil {
		t.Errorf("Unmarshalling integrationConfig failed with %s", err)
		return
	}

	spi := &internal.PluginSPIImpl{}
	ctx := context.TODO()

	providerID, err := spi.GetMachineStatus(ctx, cfg.MachineName, "", cfg.ProviderSpec, cfg.Secrets)
	if err == nil {
		t.Errorf("Machine name %s already existing", cfg.MachineName)
		return
	}
	switch err.(type) {
	case *errors.MachineNotFoundError:
		// expected
	default:
		t.Errorf("Unexpected error on GetMachineStatus %v", err)
		return
	}

	providerID, err = spi.DeleteMachine(ctx, cfg.MachineName, providerID, cfg.ProviderSpec, cfg.Secrets)
	switch err.(type) {
	case *errors.MachineNotFoundError:
		// expected
	default:
		t.Errorf("Unexpected error on DeleteMachine")
		return
	}

	providerID, err = spi.CreateMachine(ctx, cfg.MachineName, cfg.ProviderSpec, cfg.Secrets)
	if err != nil {
		t.Errorf("CreateMachine failed with %s", err)
		return
	}

	providerID2, err := spi.GetMachineStatus(ctx, cfg.MachineName, "", cfg.ProviderSpec, cfg.Secrets)
	if err != nil {
		t.Errorf("GetMachineStatus by machine name failed with %s", err)
		return
	}
	if providerID != providerID2 {
		t.Errorf("ProviderID mismatch %s != %s", providerID, providerID2)
	}

	providerID2, err = spi.GetMachineStatus(ctx, cfg.MachineName, providerID, cfg.ProviderSpec, cfg.Secrets)
	if err != nil {
		t.Errorf("GetMachineStatus by providerID failed with %s", err)
		return
	}
	if providerID != providerID2 {
		t.Errorf("ProviderID mismatch %s != %s", providerID, providerID2)
	}

	providerIDList, err := spi.ListMachines(ctx, cfg.ProviderSpec, cfg.Secrets)
	if err != nil {
		t.Errorf("ListMachines failed with %s", err)
	}

	found := false
	for id, name := range providerIDList {
		if id == providerID {
			if name != cfg.MachineName {
				t.Errorf("MachineName mismatch %s != %s", providerID, id)
			}
			found = true
		}
	}
	if !found {
		t.Errorf("Created machine with ID %s not found", providerID)
	}

	providerID2, err = spi.ShutDownMachine(ctx, cfg.MachineName, providerID, cfg.ProviderSpec, cfg.Secrets)
	if err != nil {
		t.Errorf("ShutDownMachine failed with %s", err)
	}
	if providerID != providerID2 {
		t.Errorf("ProviderID mismatch %s != %s", providerID, providerID2)
	}

	providerID2, err = spi.DeleteMachine(ctx, cfg.MachineName, providerID, cfg.ProviderSpec, cfg.Secrets)
	if err != nil {
		t.Errorf("DeleteMachine failed with %s", err)
	}
	if providerID != providerID2 {
		t.Errorf("ProviderID mismatch %s != %s", providerID, providerID2)
	}
}
